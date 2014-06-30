// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import "github.com/vmware/goipmi"

type hp struct {
	*server
	c *ipmi.Connection
}

func newHPMedia(c *ipmi.Connection, id *ipmi.DeviceIDResponse) (Media, error) {
	return &hp{
		server: newServer(c),
		c:      c,
	}, nil
}

func (*hp) BootDevice(image string) ipmi.BootDevice {
	return ipmi.BootDeviceNone
}

func (m *hp) Mount(image string) error {
	err := m.server.Mount(image)
	if err != nil {
		return err
	}

	mediaType := "floppy"
	if isISO(image) {
		mediaType = "cdrom"
	}

	return runSSH(m.c,
		sshCommand("vm", mediaType, "insert", m.url.String()),
		// iLO v2 does not support setting this over IPMI
		sshCommand("vm", mediaType, "set", "boot_once"))
}
