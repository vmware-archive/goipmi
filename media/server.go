// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"
)

type server struct {
	wg       sync.WaitGroup
	addr     net.TCPAddr
	listener net.Listener
	url      map[ipmi.BootDevice]string
	files    map[string]string
	c        *ipmi.Connection
}

func newServer(c *ipmi.Connection) *server {
	return &server{c: c}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if file, ok := s.files[r.URL.Path]; ok {
		http.ServeFile(w, r, file)
	} else {
		http.NotFound(w, r)
	}
}

func (s *server) Mount(media *VirtualMedia) error {
	listener, err := net.ListenTCP("tcp4", &s.addr)
	if err != nil {
		return err
	}

	host := s.c.LocalIP()
	port := listener.Addr().(*net.TCPAddr).Port

	s.listener = listener
	s.url = make(map[ipmi.BootDevice]string)
	s.files = make(map[string]string)

	handlers := map[ipmi.BootDevice]string{
		ipmi.BootDeviceRemoteCdrom:  media.CdromImage,
		ipmi.BootDeviceRemoteFloppy: media.FloppyImage,
	}

	for dev, file := range handlers {
		if file == "" {
			continue
		}
		path := fmt.Sprintf("/%s%s", dev, filepath.Ext(file))
		s.url[dev] = fmt.Sprintf("http://%s:%d%s", host, port, path)
		s.files[path] = file
	}

	s.wg.Add(1)

	go func() {
		_ = (&http.Server{Handler: s}).Serve(s.listener)
		s.wg.Done()
	}()

	return nil
}

func (s *server) UnMount() error {
	ioclose(s.listener)
	s.wg.Wait()
	return nil
}
