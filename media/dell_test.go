// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dellSim struct {
	*ipmi.Simulator
	c *ipmi.Connection

	calledSetBoot bool
	calledControl bool
}

func (s *dellSim) Run() error {
	s.Simulator = ipmi.NewSimulator(net.UDPAddr{})
	if err := s.Simulator.Run(); err != nil {
		return err
	}

	s.SetHandler(ipmi.NetworkFunctionApp, ipmi.CommandGetDeviceID, func(*ipmi.Message) ipmi.Response {
		return &ipmi.DeviceIDResponse{
			CompletionCode: ipmi.CommandCompleted,
			ManufacturerID: ipmi.OemDell,
		}
	})
	s.SetHandler(ipmi.NetworkFunctionChassis, ipmi.CommandSetSystemBootOptions, func(*ipmi.Message) ipmi.Response {
		s.calledSetBoot = true
		return ipmi.CommandCompleted
	})
	s.SetHandler(ipmi.NetworkFunctionChassis, ipmi.CommandChassisControl, func(*ipmi.Message) ipmi.Response {
		s.calledControl = true
		return ipmi.CommandCompleted
	})

	s.c = s.NewConnection()

	return nil
}

func TestDell(t *testing.T) {
	defaultDellVmcli = "echo"

	s := &dellSim{}
	err := s.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	calledHandler := false

	err = Boot(s.c, "dell_test.go", func() error {
		calledHandler = true
		return nil
	})

	assert.NoError(t, err)

	assert.True(t, s.calledSetBoot)
	assert.True(t, s.calledControl)
	assert.True(t, calledHandler)
}
