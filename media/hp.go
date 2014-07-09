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

func (m *hp) Mount(media *VirtualMedia) error {
	cmds := []string{}
	err := m.server.Mount(media)
	if err != nil {
		return err
	}

	devices := map[ipmi.BootDevice]string{
		ipmi.BootDeviceRemoteCdrom:  "cdrom",
		ipmi.BootDeviceRemoteFloppy: "floppy",
	}

	for dev, name := range devices {
		if url, ok := m.url[dev]; ok {
			cmds = append(cmds, sshCommand("vm", name, "insert", url))
		}
	}

	// iLO v2 does not support setting this over IPMI
	if media.BootDevice != ipmi.BootDeviceNone {
		cmds = append(cmds, sshCommand("vm", devices[media.BootDevice], "set", "boot_once"))
	}

	return runSSH(m.c, cmds...)
}
