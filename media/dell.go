// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"net"
	"os/exec"
)

var defaultDellVmcli = "vmcli"

type dell struct {
	process
	bootDevice
	c *ipmi.Connection
}

func newDellMedia(c *ipmi.Connection, id *ipmi.DeviceIDResponse) (Media, error) {
	return &dell{c: c}, nil
}

func (m *dell) Mount(image string) error {
	args := []string{"-r", ipmiAddress(m.c), "-u", m.c.Username, "-p", m.c.Password}
	flag := "-f"
	if isISO(image) {
		flag = "-c"
	}

	m.Cmd = exec.Command(m.cli(), append(args, flag, image)...)

	return m.start()
}

func (m *dell) cli() string {
	return defaultDellVmcli // TODO: racvmcli for v5
}

// the Dell vmcli tools do not resolve hostnames, so make sure we give it an IP address
func ipmiAddress(c *ipmi.Connection) string {
	if net.ParseIP(c.Hostname) == nil {
		addrs, err := net.LookupHost(c.Hostname)
		if err != nil && len(addrs) > 0 {
			return addrs[0]
		}
	}
	return c.Hostname
}
