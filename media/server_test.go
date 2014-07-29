// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	vm := VirtualMedia{
		IMG: &VirtualDevice{
			Path: "server.go",
		},
		ISO: &VirtualDevice{
			Path: "server_test.go",
		},
	}
	err := vm.ListenAndServe("localhost")
	assert.NoError(t, err)

	for _, device := range vm {
		st, _ := os.Stat(device.Path)
		r, err := http.Get(device.URL.String())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.Equal(t, r.ContentLength, st.Size(), device.URL.String())
	}

	u, _ := url.Parse(vm[ISO].URL.String())
	u.Path = "/dell.go"
	r, err := http.Get(u.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, r.StatusCode, "file exists but should 404")

	u.Path = "/enoent.go"
	r, err = http.Get(u.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, r.StatusCode)
}

func TestSambaURL(t *testing.T) {
	os.Setenv("IPMI_MEDIA_SHARE_PATH", "/usr/share")
	p := "/images/foo.iso"
	u := url.User("guest")
	x, err := url.Parse("smb://user:pass@xhost/isos/x.iso")
	assert.NoError(t, err)

	tests := []struct {
		device VirtualDevice
		host   string
		expect url.URL
	}{
		{VirtualDevice{Path: p}, "h", url.URL{Host: "h", Path: p, User: u}},
		{VirtualDevice{Path: "/usr/share" + p}, "", url.URL{Path: p, User: u}},
		{VirtualDevice{URL: x}, "", *x},
		{VirtualDevice{URL: x}, "h", *x},
	}

	for _, test := range tests {
		tu := test.device.SambaURL(test.host)
		assert.Equal(t, test.expect.String(), tu.URL.String())
		assert.Equal(t, strings.Replace(tu.URL.Path, "/", "\\", -1), tu.SharePath)
	}
}
