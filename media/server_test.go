// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"net/http"
	"net/url"
	"os"
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
