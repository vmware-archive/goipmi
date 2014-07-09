// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	s := &server{}
	vm := &VirtualMedia{
		CdromImage: "server_test.go",
		BootDevice: ipmi.BootDeviceRemoteCdrom,
	}
	err := s.Mount(vm)
	assert.NoError(t, err)

	surl := s.url[ipmi.BootDeviceRemoteCdrom]
	r, err := http.Get(surl)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.StatusCode)

	u, _ := url.Parse(surl)
	u.Path = "/server.go"
	r, err = http.Get(u.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, r.StatusCode, "file exists but should 404")

	err = s.UnMount()
	assert.NoError(t, err)
}
