// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ListenAndServe starts and HTTP server that can serve virtual media files.
func (m DeviceMap) ListenAndServe(host string) error {
	l, err := net.ListenTCP("tcp4", &net.TCPAddr{})
	if err != nil {
		return err
	}

	port := l.Addr().(*net.TCPAddr).Port

	for id, device := range m {
		device.URL = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", host, port),
			Path:   fmt.Sprintf("/%s%s", id, filepath.Ext(device.Path)),
		}
	}

	go (&http.Server{Handler: m}).Serve(l)
	return nil
}

func (m DeviceMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, device := range m {
		if device.URL.Path == r.URL.Path {
			http.ServeFile(w, r, device.Path)
			return
		}
	}

	http.NotFound(w, r)
}

// SambaURL extends url.URL, adding a SharePath in UNC form
type SambaURL struct {
	url.URL
	SharePath string
}

// SambaURL returns a URL for which the device media can be accessed
// over samba by the remote BMC.
func (d *Device) SambaURL(host string) *SambaURL {
	var u url.URL

	if d.URL == nil {
		path := d.Path
		share := os.Getenv("IPMI_MEDIA_SHARE_PATH")
		// make path relative to share path if set
		if share != "" {
			if strings.HasPrefix(path, share) {
				path = path[len(share):]
			}
		}

		u = url.URL{
			Host: host,
			Path: path,
			User: url.User("guest"),
		}
	} else {
		u = *d.URL
	}

	return &SambaURL{
		URL:       u,
		SharePath: strings.Replace(u.Path, "/", "\\", -1),
	}
}
