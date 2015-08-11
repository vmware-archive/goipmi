/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
			&Connection{
				Path:      "",
				Hostname:  "h",
				Port:      0,
				Username:  "u",
				Password:  "p",
				Interface: "",
			},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lanplus"},
		},
		{
			"should append port",
			&Connection{
				Path:      "",
				Hostname:  "h",
				Port:      1623,
				Username:  "u",
				Password:  "p",
				Interface: "",
			},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lanplus", "-p", "1623"},
		},
		{
			"should override default interface",
			&Connection{
				Path:      "",
				Hostname:  "h",
				Port:      0,
				Username:  "u",
				Password:  "p",
				Interface: "lan",
			},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lan"},
		},
	}

	for _, test := range tests {
		result := newToolTransport(test.conn).(*tool).options()
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

	// Set Boot Options
	data := []uint8{0x80, uint8(BootDevicePxe) | 0x40}
	req = &Request{
		NetworkFunctionChassis,
		CommandSetSystemBootOptions,
		&SetSystemBootOptionsRequest{
			Param: BootParamBootFlags,
			Data:  data,
		},
	}
	err = tr.send(req, &SetSystemBootOptionsResponse{})
	assert.Error(t, err) // ErrShortPacket
	// resend with valid Data length
	req.Data.(*SetSystemBootOptionsRequest).Data = append(data, 0x00, 0x00, 0x00)
	err = tr.send(req, &SetSystemBootOptionsResponse{})
	assert.NoError(t, err)

	// Get Boot Options
	req = &Request{
		NetworkFunctionChassis,
		CommandGetSystemBootOptions,
		&SystemBootOptionsRequest{
			Param: BootParamBootFlags,
		},
	}
	bor := &SystemBootOptionsResponse{}
	err = tr.send(req, bor)
	assert.NoError(t, err)
	assert.Equal(t, uint8(BootParamBootFlags), bor.Param)
	assert.Equal(t, uint8(BootDevicePxe), bor.BootDeviceSelector())
	assert.Equal(t, uint8(0x40), bor.Data[1]&0x40)

	// Invalid command
	req.Command = 0xff
	err = tr.send(req, &DeviceIDResponse{})
	assert.Error(t, err)

	err = tr.close()
	assert.NoError(t, err)
	s.Stop()
}
