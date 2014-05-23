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
	c.Username = "foo"

	err = c.EnableNetworkBoot()
	assert.Error(t, err)

	err = c.PowerCycle()
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

	err = c.EnableNetworkBoot()
	assert.NoError(t, err)
	assert.True(t, calledOptions)
	err = c.PowerCycle()
	assert.NoError(t, err)
	assert.True(t, calledControl)

	s.Stop()
}
