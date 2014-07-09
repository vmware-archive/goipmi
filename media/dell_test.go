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

	calledSetBoot ipmi.BootDevice
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
	s.SetHandler(ipmi.NetworkFunctionChassis, ipmi.CommandSetSystemBootOptions, func(m *ipmi.Message) ipmi.Response {
		if m.Data[0] == ipmi.BootParamBootFlags {
			s.calledSetBoot = ipmi.BootDevice(m.Data[2])
		}
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
	vm := &VirtualMedia{
		CdromImage: "dell_test.go",
		BootDevice: ipmi.BootDeviceRemoteCdrom,
	}
	err = Boot(s.c, vm, func() error {
		calledHandler = true
		return nil
	})

	assert.NoError(t, err)

	assert.Equal(t, vm.BootDevice, s.calledSetBoot)
	assert.True(t, s.calledControl)
	assert.True(t, calledHandler)
}
