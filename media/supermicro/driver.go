// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package supermicro

import (
	"bytes"
	"github.com/vmware/goipmi"
	"github.com/vmware/goipmi/media"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
)

const (
	smcLogin     = "/cgi/login.cgi"
	smcInsertIMG = "/cgi/uimapin.cgi"
	smcEjectIMG  = "/cgi/uimapout.cgi"
	smcConfigISO = "/cgi/virtual_media_share_img.cgi"
	smcInsertISO = "/cgi/uisopin.cgi"
	smcEjectISO  = "/cgi/uisopout.cgi"
)

var (
	smcScheme = "http"
	smcPort   = "80"
)

type driver struct {
	http.Client
	c *ipmi.Client
}

func init() {
	media.Register(ipmi.OemSupermicro, New)
}

// New driver instance
func New(c *ipmi.Client) (media.Driver, error) {
	return &driver{c: c}, nil
}

func (d *driver) Insert(m media.DeviceMap) error {
	if err := d.login(); err != nil {
		return err
	}

	devices := map[string]struct {
		boot ipmi.BootDevice
		call func(*media.Device) error
	}{
		media.ISO: {ipmi.BootDeviceCdrom, d.insertISO},
		media.IMG: {ipmi.BootDeviceFloppy, d.insertIMG},
	}

	for id, device := range m {
		if err := devices[id].call(device); err != nil {
			return err
		}
		if device.Boot {
			if err := d.c.SetBootDevice(devices[id].boot); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *driver) Eject() error {
	if err := d.login(); err != nil {
		return err
	}

	paths := map[string]string{
		media.ISO: smcEjectISO,
		media.IMG: smcEjectIMG,
	}

	for _, path := range paths {
		res, err := d.Get(d.url(path).String())
		if err != nil {
			return err
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return errors.New(res.Status)
		}
	}

	return nil
}

func (d *driver) login() error {
	val := url.Values{
		"name": []string{d.c.Username},
		"pwd":  []string{d.c.Password},
	}

	res, err := d.PostForm(d.url(smcLogin).String(), val)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	_ = res.Body.Close()

	cookies := res.Cookies()
	if len(cookies) == 0 {
		return http.ErrNoCookie
	}
	d.Jar, _ = cookiejar.New(nil)
	d.Jar.SetCookies(d.url("/"), cookies)

	return nil
}

func (d *driver) insertIMG(device *media.Device) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(device.Path))
	if err != nil {
		return err
	}

	file, err := os.Open(device.Path)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	_ = file.Close()
	if err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	res, err := d.Post(d.url(smcInsertIMG).String(), writer.FormDataContentType(), body)
	_ = res.Body.Close()
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}

	return nil
}

func (d *driver) configISO(device *media.Device) error {
	smb := device.SambaURL(d.c.LocalIP())
	pwd, _ := smb.User.Password()
	val := url.Values{
		"host": []string{smb.Host},
		"path": []string{smb.SharePath},
		"user": []string{smb.User.Username()},
		"pwd":  []string{pwd},
	}

	res, err := d.PostForm(d.url(smcConfigISO).String(), val)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	_ = res.Body.Close()
	return nil
}

func (d *driver) insertISO(device *media.Device) error {
	if err := d.configISO(device); err != nil {
		return err
	}

	res, err := d.Get(d.url(smcInsertISO).String())
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	return nil
}

func (d *driver) url(path string) *url.URL {
	return &url.URL{
		Scheme: smcScheme,
		Host:   net.JoinHostPort(d.c.Hostname, smcPort),
		Path:   path,
	}
}
