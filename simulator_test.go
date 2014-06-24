// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimulator(t *testing.T) {
	s := NewSimulator(net.UDPAddr{Port: 0})
	err := s.Run()
	assert.NoError(t, err)

	c := s.NewConnection()
	client, err := NewClient(c)
	assert.NoError(t, err)
	err = client.Open()
	assert.NoError(t, err)

	for _, cmd := range []Command{CommandChassisControl, CommandSetSystemBootOptions} {
		s.SetHandler(NetworkFunctionChassis, cmd, func(*Message) Response {
			return UnspecifiedError
		})
	}

	err = client.SetBootDevice(BootDevicePxe)
	assert.Error(t, err)

	err = client.Control(ControlPowerCycle)
	assert.Error(t, err)

	var calledControl, calledOptions bool

	s.SetHandler(NetworkFunctionChassis, CommandChassisControl, func(m *Message) Response {
		calledControl = true
		assert.Equal(t, c.Username, m.RequestID)
		return CommandCompleted
	})

	s.SetHandler(NetworkFunctionChassis, CommandSetSystemBootOptions, func(m *Message) Response {
		calledOptions = true
		assert.Equal(t, c.Username, m.RequestID)
		return CommandCompleted
	})

	err = client.SetBootDevice(BootDevicePxe)
	assert.NoError(t, err)
	assert.True(t, calledOptions)
	err = client.Control(ControlPowerCycle)
	assert.NoError(t, err)
	assert.True(t, calledControl)

	client.Close()
	s.Stop()
}
