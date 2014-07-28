// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import "github.com/vmware/goipmi"

type hp struct {
	c *ipmi.Connection
}

func newHPMedia(c *ipmi.Client) (Media, error) {
	return &hp{
		c: c.Connection,
	}, nil
}

func (m *hp) Mount(media VirtualMedia) error {
	cmds := []string{}
	err := media.ListenAndServe(m.c.LocalIP())
	if err != nil {
		return err
	}

	devices := map[string]string{
		ISO: "cdrom",
		IMG: "floppy",
	}

	for id, device := range media {
		cmds = append(cmds, sshCommand("vm", devices[id], "insert", device.URL.String()))
		if device.Boot {
			cmds = append(cmds, sshCommand("vm", devices[id], "set", "boot_once"))
		}
	}

	return runSSH(m.c, cmds...)
}

func (m *hp) UnMount() error {
	return nil
}
