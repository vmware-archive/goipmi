// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	s := &server{}
	err := s.Mount("server_test.go")
	assert.NoError(t, err)

	url := *s.url
	r, err := http.Get(url.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.StatusCode)

	url.Path = "/server.go"
	r, err = http.Get(url.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, r.StatusCode)

	err = s.UnMount()
	assert.NoError(t, err)
}
