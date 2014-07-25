// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	s := newServer(&ipmi.Connection{Hostname: "localhost"})
	vm := &VirtualMedia{
		FloppyImage: "server.go",
		CdromImage:  "server_test.go",
		BootDevice:  ipmi.BootDeviceRemoteCdrom,
	}
	err := s.Mount(vm)
	assert.NoError(t, err)

	tests := map[ipmi.BootDevice]string{
		ipmi.BootDeviceRemoteCdrom:  vm.CdromImage,
		ipmi.BootDeviceRemoteFloppy: vm.FloppyImage,
	}

	for dev, file := range tests {
		st, _ := os.Stat(file)
		surl := s.url[dev]
		r, err := http.Get(surl)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.Equal(t, r.ContentLength, st.Size())
	}

	u, _ := url.Parse(s.url[ipmi.BootDeviceRemoteCdrom])
	u.Path = "/dell.go"
	r, err := http.Get(u.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, r.StatusCode, "file exists but should 404")

	u.Path = "/enoent.go"
	r, err = http.Get(u.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, r.StatusCode)

	err = s.UnMount()
	assert.NoError(t, err)
}
