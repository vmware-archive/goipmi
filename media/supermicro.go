// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"bytes"
	"github.com/vmware/goipmi"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var (
	smcLoginPath           = "cgi/login.cgi"
	smcVMFloppyPath        = "cgi/uimapin.cgi"
	smcVMCDSharePath       = "cgi/virtual_media_share_img.cgi"
	smcVMCDMountPath       = "cgi/uisopin.cgi"
	smcVMStatusPath        = "cgi/vmstatus.cgi"
	smcVMFloppyUnmountPath = "cgi/uimapout.cgi"
	smcVMUnmountCDPath     = "cgi/uisopout.cgi"
)

type supermicro struct {
	c  *http.Client
	u  *url.URL
	ic *ipmi.Client
}

func newSupermicroMedia(c *ipmi.Client) (Media, error) {
	clnt := &supermicro{
		c: &http.Client{},
		u: &url.URL{
			Host:   c.Hostname,
			Scheme: "http",
		},
		ic: c,
	}

	v := url.Values{
		"name": []string{c.Username},
		"pwd":  []string{c.Password},
	}

	res, err := clnt.postForm(clnt.url(smcLoginPath), v)
	if err != nil {
		return nil, err
	}

	// Check if SID is set in the jar, meaning authentication succeeded.
	//
	// NOTE: The BMC will send back SID more than once.  One of the SID entries is
	// empty and must be pruned from future queries or the BMC will fail the
	// request silently.
	jar, _ := cookiejar.New(nil)
	for _, c := range res.Cookies() {
		if c.Name == "SID" && len(c.Value) > 0 {
			jar.SetCookies(clnt.u, []*http.Cookie{c})
			clnt.c.Jar = jar
			break
		}
	}

	if clnt.c.Jar == nil {
		return nil, errors.New("auth failed")
	}

	return clnt, nil
}

// Mount mounts the floppy and CD virtual devices
func (s *supermicro) Mount(media VirtualMedia) error {
	if device, ok := media[ISO]; ok {
		if err := s.mountCD(s.ic.LocalIP(), device.Path, "Guest", "Guest"); err != nil {
			return err
		}
		if err := s.ic.SetBootDevice(ipmi.BootDeviceCdrom); err != nil {
			return err
		}
	}

	if device, ok := media[IMG]; ok {
		if err := s.mountFloppy(device.Path); err != nil {
			return err
		}
		if err := s.ic.SetBootDevice(ipmi.BootDeviceFloppy); err != nil {
			return err
		}
	}

	return nil
}

func (s *supermicro) UnMount() error {
	if err := s.unmountFloppy(); err != nil {
		return err
	}

	if err := s.unmountCD(); err != nil {
		return err
	}
	return nil
}

func (s *supermicro) mountFloppy(path string) error {
	if filepath.Ext(path) != ".img" {
		return errors.New("floppy image file must have a .img extension")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// NOTE: The BMC does not like chunked encoding, and as such, requires a
	// Content-Length field in the request.  The http.Client will set the
	// Content-Length field if the body is non nil and set with the data to be
	// POSTed.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return err
	}

	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	var req *http.Request
	req, err = http.NewRequest("POST", s.url(smcVMFloppyPath).String(), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	var res *http.Response
	res, err = s.c.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", res.Status)
	}
	res.Body.Close()

	isMounted := false
	// The BMC will not mount the floppy immediately in a transaction but rather
	// after a few seconds.  Poll here (sigh) for a few seconds before throwing
	// an error.
	for i := 0; i < 10; i++ {
		if isMounted, err = s.floppy(); err != nil {
			return err
		}

		if isMounted {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if isMounted == false {
		return errors.New("floppy failed to mount")
	}

	return nil
}

func (s *supermicro) setCDSharePath(host, path, user, password string) error {
	if filepath.Ext(path) != ".iso" {
		return errors.New("CD image file must have a .iso extension")
	}

	v := url.Values{
		"host": []string{host},
		"path": []string{path},
		"user": []string{user},
		"pwd":  []string{password},
	}

	res, err := s.postForm(s.url(smcVMCDSharePath), v)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

func (s *supermicro) mountCD(host, path, user, password string) error {
	if err := s.setCDSharePath(host, path, user, password); err != nil {
		return err
	}

	res, err := s.get(s.url(smcVMCDMountPath))
	if err != nil {
		return err
	}
	res.Body.Close()

	isMounted := false
	// The BMC will not mount the CD immediately in a transaction but rather
	// after a few seconds.  Poll here (sigh) for a few seconds before throwing
	// an error.
	for i := 0; i < 10; i++ {
		if isMounted, err = s.cd(); err != nil {
			return err
		}

		if isMounted {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if isMounted == false {
		return errors.New("CD failed to mount")
	}

	return nil
}

type device struct {
	ID     uint8 `xml:"ID,attr"`
	Status uint8 `xml:"STATUS,attr"`
}

type code struct {
	NO uint8 `xml:"NO,attr"`
}

type vmmtab struct {
	C      code     `xml:"CODE"`
	Drives []device `xml:"DEVICE"`
}

func (s *supermicro) mtab() (*vmmtab, error) {
	res, err := s.get(s.url(smcVMStatusPath))
	if err != nil {
		return nil, err
	}

	respBody := &bytes.Buffer{}
	if _, err := respBody.ReadFrom(res.Body); err != nil {
		return nil, err
	}
	res.Body.Close()

	var mtab vmmtab
	if err := xml.Unmarshal(respBody.Bytes(), &mtab); err != nil {
		return nil, err
	}

	return &mtab, nil
}

func (s *supermicro) floppy() (bool, error) {
	mtab, err := s.mtab()
	if err != nil {
		return false, err
	}

	// The device ID 0 in the array is the floppy.
	// Unmounted is 255, mounted is 0.
	for _, m := range mtab.Drives {
		if m.ID == 0 {
			if m.Status == 0 {
				return true, nil
			}
			break
		}
	}
	return false, nil
}

func (s *supermicro) cd() (bool, error) {
	mtab, err := s.mtab()
	if err != nil {
		return false, err
	}

	// The device ID 1 in the array is the CD drive.
	// Unmounted is 255, mounted is 4.
	for _, m := range mtab.Drives {
		if m.ID == 1 {
			if m.Status == 4 {
				return true, nil
			}
			break
		}
	}
	return false, nil
}

func (s *supermicro) unmountFloppy() error {
	res, err := s.get(s.url(smcVMFloppyUnmountPath))
	if err != nil {
		return err
	}
	res.Body.Close()

	var floppyMounted bool
	floppyMounted, err = s.floppy()
	if err != nil {
		return err
	}

	if floppyMounted {
		return errors.New("cannot unmount floppy")
	}
	return nil
}

func (s *supermicro) unmountCD() error {
	res, err := s.get(s.url(smcVMUnmountCDPath))
	if err != nil {
		return err
	}
	res.Body.Close()

	var mounted bool
	mounted, err = s.cd()
	if err != nil {
		return err
	}

	if mounted {
		return errors.New("cannot unmount CD")
	}
	return nil
}

func (s *supermicro) url(path string) *url.URL {
	u := *s.u
	u.Path = path
	return &u
}

func (s *supermicro) get(u *url.URL) (*http.Response, error) {
	res, err := s.c.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", res.Status)
	}
	return res, nil
}

func (s *supermicro) postForm(u *url.URL, v url.Values) (*http.Response, error) {
	res, err := s.c.PostForm(u.String(), v)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", res.Status)
	}
	return res, nil
}
