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
	c        *ipmi.Connection
}

func newServer(c *ipmi.Connection) *server {
	return &server{c: c}
}

func (s *server) hostIP() (string, error) {
	if s.c == nil {
		return "localhost", nil
	}

	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", s.c.Hostname, s.c.Port))
	if err != nil {
		return "", err
	}
	defer ioclose(conn)
	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	return host, err
}

func (s *server) Mount(media *VirtualMedia) error {
	listener, err := net.ListenTCP("tcp4", &s.addr)
	if err != nil {
		return err
	}

	port := listener.Addr().(*net.TCPAddr).Port
	host, err := s.hostIP()
	if err != nil {
		return err
	}

	s.listener = listener
	mux := http.NewServeMux()
	s.url = make(map[ipmi.BootDevice]string)

	handlers := map[ipmi.BootDevice]string{
		ipmi.BootDeviceRemoteCdrom:  media.CdromImage,
		ipmi.BootDeviceRemoteFloppy: media.FloppyImage,
	}

	for dev, file := range handlers {
		if file == "" {
			continue
		}
		path := fmt.Sprintf("/%s", filepath.Base(file))
		s.url[dev] = fmt.Sprintf("http://%s:%d%s", host, port, path)

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, file)
		})
	}

	s.wg.Add(1)

	go func() {
		_ = (&http.Server{Handler: mux}).Serve(s.listener)
		s.wg.Done()
	}()

	return nil
}

func (s *server) UnMount() error {
	ioclose(s.listener)
	s.wg.Wait()
	return nil
}
