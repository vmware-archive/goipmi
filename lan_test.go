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
	assert.Equal(t, ErrInvalidCommand, err)

	err = tr.close()
	assert.NoError(t, err)
	s.Stop()
}
