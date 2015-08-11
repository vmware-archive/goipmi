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
			return ErrUnspecified
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
