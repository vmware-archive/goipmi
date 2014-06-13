// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLAN(t *testing.T) {
	s := NewSimulator(net.UDPAddr{Port: 0})
	err := s.Run()
	assert.NoError(t, err)

	c := &Connection{
		Hostname:  "127.0.0.1",
		Port:      s.LocalAddr().Port,
		Username:  "vmware",
		Password:  "cow",
		Interface: "lan",
	}

	tr, err := newTransport(c)
	assert.NoError(t, err)

	err = tr.open()
	assert.NoError(t, err)

	req := &Request{
		NetworkFunctionApp,
		CommandGetDeviceID,
		&DeviceIDRequest{},
	}
	res := &DeviceIDResponse{}

	err = tr.send(req, res)
	assert.NoError(t, err)

	assert.Equal(t, uint8(0x51), res.IPMIVersion)

	req.Command = 0xff
	err = tr.send(req, res)
	assert.Equal(t, InvalidCommand, err)

	err = tr.close()
	assert.NoError(t, err)
	s.Stop()
}
