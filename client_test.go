// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	s := NewSimulator(net.UDPAddr{Port: 0})
	err := s.Run()
	assert.NoError(t, err)

	client, err := NewClient(s.NewConnection())
	assert.NoError(t, err)

	err = client.Open()
	assert.NoError(t, err)

	err = client.SetBootDevice(BootDevicePxe)
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
	s.Stop()
}

func TestDeviceID(t *testing.T) {
	s := NewSimulator(net.UDPAddr{})
	err := s.Run()
	assert.NoError(t, err)

	client, err := NewClient(s.NewConnection())
	assert.NoError(t, err)

	err = client.Open()
	assert.NoError(t, err)

	tests := []OemID{
		OemDell, OemHP,
	}

	for _, test := range tests {
		s.SetHandler(NetworkFunctionApp, CommandGetDeviceID, func(*Message) Response {
			return &DeviceIDResponse{
				CompletionCode: CommandCompleted,
				ManufacturerID: test,
			}
		})

		id, err := client.DeviceID()
		assert.NoError(t, err)
		assert.Equal(t, test, id.ManufacturerID)
		assert.Equal(t, test.String(), id.ManufacturerID.String())
	}

	err = client.Close()
	assert.NoError(t, err)
	s.Stop()
}
