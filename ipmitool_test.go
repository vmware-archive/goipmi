// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		should string
		conn   *Connection
		expect []string
	}{
		{
			"should use default port and interface",
			&Connection{"", "h", 0, "u", "p", ""},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lanplus"},
		},
		{
			"should append port",
			&Connection{"", "h", 1623, "u", "p", ""},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lanplus", "-p", "1623"},
		},
		{
			"should override default interface",
			&Connection{"", "h", 0, "u", "p", "lan"},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lan"},
		},
	}

	for _, test := range tests {
		result := test.conn.options()
		assert.Equal(t, test.expect, result, test.should)
	}
}

func TestTool(t *testing.T) {
	s := NewSimulator(net.UDPAddr{Port: 0})
	err := s.Run()
	assert.NoError(t, err)

	c := &Connection{
		Hostname:  "127.0.0.1",
		Port:      s.LocalAddr().Port,
		Username:  "vmware",
		Password:  "cow",
		Interface: "lan",
		Path:      "ipmitool",
	}

	tr, err := newTransport(c)
	assert.NoError(t, err)

	err = tr.open()
	assert.NoError(t, err)

	// Device ID
	req := &Request{
		NetworkFunctionApp,
		CommandGetDeviceID,
		&DeviceIDRequest{},
	}
	dir := &DeviceIDResponse{}
	err = tr.send(req, dir)
	assert.NoError(t, err)
	assert.Equal(t, uint8(0x51), dir.IPMIVersion)

	// Chassis Status
	req = &Request{
		NetworkFunctionChassis,
		CommandChassisStatus,
		&DeviceIDRequest{},
	}
	csr := &ChassisStatusResponse{}
	err = tr.send(req, csr)
	assert.NoError(t, err)
	assert.Equal(t, uint8(SystemPower), csr.PowerState)

	// Invalid command
	req.Command = 0xff
	err = tr.send(req, &DeviceIDResponse{})
	assert.Error(t, err)

	err = tr.close()
	assert.NoError(t, err)
	s.Stop()
}
