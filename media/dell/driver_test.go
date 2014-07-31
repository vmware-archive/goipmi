// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package dell

import (
	"github.com/vmware/goipmi"
	"github.com/vmware/goipmi/media"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type sim struct {
	*ipmi.Simulator
	c *ipmi.Connection

	calledSetBoot ipmi.BootDevice
	calledControl bool
}

func (s *sim) Run() error {
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

	s := &sim{}
	err := s.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	calledHandler := false
	vm := media.DeviceMap{
		media.ISO: &media.Device{
			Path: "driver_test.go",
			Boot: true,
		},
	}
	err = media.Boot(s.c, vm, func(*ipmi.Client) error {
		calledHandler = true
		return nil
	})

	assert.NoError(t, err)

	assert.Equal(t, ipmi.BootDeviceRemoteCdrom, s.calledSetBoot)
	assert.True(t, s.calledControl)
	assert.True(t, calledHandler)
}
