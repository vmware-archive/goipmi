// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"fmt"
)

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
		cmds = append(cmds, fmt.Sprintf("vm %s insert %s", devices[id], device.URL))
		if device.Boot {
			cmds = append(cmds, fmt.Sprintf("vm %s set boot_once", devices[id]))
		}
	}

	return m.c.RunSSH(cmds...)
}

func (m *hp) UnMount() error {
	return nil
}
