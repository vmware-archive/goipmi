// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package dell

import (
	"github.com/vmware/goipmi"
	"github.com/vmware/goipmi/media"
	"os/exec"
)

var defaultDellVmcli = "vmcli"

type driver struct {
	*exec.Cmd
	c   *ipmi.Client
	cli string
}

func init() {
	media.Register(ipmi.OemDell, New)
}

// New driver instance
func New(c *ipmi.Client) (media.Driver, error) {
	cli := defaultDellVmcli // TODO: racvmcli for v5
	return &driver{c: c, cli: cli}, nil
}

// Note that Dell vmcli only supports 1 active session, but supports inserting
// both an iso and img within the single session.
func (d *driver) Insert(m media.DeviceMap) error {
	args := []string{"-r", d.c.RemoteIP(), "-u", d.c.Username, "-p", d.c.Password}

	devices := map[string]struct {
		boot ipmi.BootDevice
		flag string
	}{
		media.ISO: {ipmi.BootDeviceRemoteCdrom, "-c"},
		media.IMG: {ipmi.BootDeviceRemoteFloppy, "-f"},
	}

	for id, device := range m {
		args = append(args, devices[id].flag, device.Path)
		if device.Boot {
			if err := d.c.SetBootDevice(devices[id].boot); err != nil {
				return err
			}
		}
	}

	d.Cmd = exec.Command(d.cli, args...)

	return d.Start()
}

func (d *driver) Eject() error {
	return d.Process.Kill()
}
