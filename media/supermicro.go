// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"bytes"
	"github.com/vmware/goipmi"
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

type supermicro struct {
	http.Client
	c *ipmi.Client
}

func newSupermicroMedia(c *ipmi.Client) (Media, error) {
	return &supermicro{c: c}, nil
}

func (s *supermicro) Mount(media VirtualMedia) error {
	if err := s.login(); err != nil {
		return err
	}

	devices := map[string]struct {
		boot ipmi.BootDevice
		call func(*VirtualDevice) error
	}{
		ISO: {ipmi.BootDeviceCdrom, s.insertISO},
		IMG: {ipmi.BootDeviceFloppy, s.insertIMG},
	}

	for id, device := range media {
		if err := devices[id].call(device); err != nil {
			return err
		}
		if device.Boot {
			if err := s.c.SetBootDevice(devices[id].boot); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *supermicro) UnMount() error {
	if err := s.login(); err != nil {
		return err
	}

	paths := map[string]string{
		ISO: smcEjectISO,
		IMG: smcEjectIMG,
	}

	for _, path := range paths {
		res, err := s.Get(s.url(path).String())
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

func (s *supermicro) login() error {
	val := url.Values{
		"name": []string{s.c.Username},
		"pwd":  []string{s.c.Password},
	}

	res, err := s.PostForm(s.url(smcLogin).String(), val)
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
	s.Jar, _ = cookiejar.New(nil)
	s.Jar.SetCookies(s.url("/"), cookies)

	return nil
}

func (s *supermicro) insertIMG(device *VirtualDevice) error {
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

	res, err := s.Post(s.url(smcInsertIMG).String(), writer.FormDataContentType(), body)
	_ = res.Body.Close()
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}

	return nil
}

func (s *supermicro) configISO(device *VirtualDevice) error {
	smb := device.SambaURL(s.c.LocalIP())
	pwd, _ := smb.User.Password()
	val := url.Values{
		"host": []string{smb.Host},
		"path": []string{smb.SharePath},
		"user": []string{smb.User.Username()},
		"pwd":  []string{pwd},
	}

	res, err := s.PostForm(s.url(smcConfigISO).String(), val)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	_ = res.Body.Close()
	return nil
}

func (s *supermicro) insertISO(device *VirtualDevice) error {
	if err := s.configISO(device); err != nil {
		return err
	}

	res, err := s.Get(s.url(smcInsertISO).String())
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	return nil
}

func (s *supermicro) url(path string) *url.URL {
	return &url.URL{
		Scheme: smcScheme,
		Host:   net.JoinHostPort(s.c.Hostname, smcPort),
		Path:   path,
	}
}
