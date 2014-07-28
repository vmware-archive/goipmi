// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"net"
	"regexp"
	"sort"
	"sync"
	"testing"

	"code.google.com/p/go.crypto/ssh"
	"github.com/stretchr/testify/assert"
)

type hpSim struct {
	*ipmi.Simulator
	wg   *sync.WaitGroup
	c    *ipmi.Connection
	cmds []string

	calledSetBoot ipmi.BootDevice
	calledControl bool
}

func (s *hpSim) Run() error {
	s.Simulator = ipmi.NewSimulator(net.UDPAddr{})
	if err := s.Simulator.Run(); err != nil {
		return err
	}

	s.SetHandler(ipmi.NetworkFunctionApp, ipmi.CommandGetDeviceID, func(*ipmi.Message) ipmi.Response {
		return &ipmi.DeviceIDResponse{
			CompletionCode: ipmi.CommandCompleted,
			ManufacturerID: ipmi.OemHP,
		}
	})
	s.SetHandler(ipmi.NetworkFunctionChassis, ipmi.CommandChassisControl, func(*ipmi.Message) ipmi.Response {
		s.calledControl = true
		return ipmi.CommandCompleted
	})

	s.c = s.NewConnection()

	s.wg = sshTestExecServer(s.c, func(ch ssh.Channel, cmd string) int {
		s.cmds = append(s.cmds, cmd)
		return 0
	})

	return nil
}

func (s *hpSim) Stop() {
	s.wg.Wait()
	s.Simulator.Stop()
}

func TestHP(t *testing.T) {
	s := &hpSim{}
	err := s.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	calledHandler := false
	vm := VirtualMedia{
		ISO: &VirtualDevice{
			Path: "hp.go",
		},
		IMG: &VirtualDevice{
			Path: "hp_test.go",
			Boot: true,
		},
	}
	err = Boot(s.c, vm, func(*ipmi.Client) error {
		calledHandler = true
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, len(s.cmds))
	sort.Strings(s.cmds)
	matches, _ := regexp.MatchString("vm cdrom insert http://127.0.0.1:[0-9]{4,5}/iso.go", s.cmds[0])
	assert.True(t, matches, s.cmds[0])
	matches, _ = regexp.MatchString("vm floppy insert http://127.0.0.1:[0-9]{4,5}/img.go", s.cmds[1])
	assert.True(t, matches, s.cmds[1])

	assert.Equal(t, "vm floppy set boot_once", s.cmds[2])
	assert.True(t, s.calledControl)
	assert.True(t, calledHandler)
}
