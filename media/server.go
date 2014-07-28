// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
)

// ListenAndServe starts and HTTP server that can serve VirtualMedia files.
func (m VirtualMedia) ListenAndServe(host string) error {
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

func (m VirtualMedia) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, device := range m {
		if device.URL.Path == r.URL.Path {
			http.ServeFile(w, r, device.Path)
			return
		}
	}

	http.NotFound(w, r)
}
