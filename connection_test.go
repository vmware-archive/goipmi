// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteIP(t *testing.T) {
	c := Connection{Hostname: "127.0.0.1"}
	assert.Equal(t, c.Hostname, c.RemoteIP())
}

func TestLocalIP(t *testing.T) {
	c := Connection{Hostname: "127.0.0.1"}
	assert.Equal(t, c.Hostname, c.LocalIP())
}
