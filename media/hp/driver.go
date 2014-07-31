// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package hp

import (
	"github.com/vmware/goipmi"
	"github.com/vmware/goipmi/media"
	"fmt"
)

type driver struct {
	c *ipmi.Client
}

func init() {
	media.Register(ipmi.OemHP, New)
}

// New driver instance
func New(c *ipmi.Client) (media.Driver, error) {
	return &driver{c: c}, nil
}

func (d *driver) Insert(m media.DeviceMap) error {
	cmds := []string{}
	err := m.ListenAndServe(d.c.LocalIP())
	if err != nil {
		return err
	}

	devices := map[string]string{
		media.ISO: "cdrom",
		media.IMG: "floppy",
	}

	for id, device := range m {
		cmds = append(cmds, fmt.Sprintf("vm %s insert %s", devices[id], device.URL))
		if device.Boot {
			cmds = append(cmds, fmt.Sprintf("vm %s set boot_once", devices[id]))
		}
	}

	return d.c.RunSSH(cmds...)
}

func (*driver) Eject() error {
	return nil
}
