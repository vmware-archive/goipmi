// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"os/exec"
)

var defaultDellVmcli = "vmcli"

type dell struct {
	process
	c *ipmi.Connection
}

func newDellMedia(c *ipmi.Connection, id *ipmi.DeviceIDResponse) (Media, error) {
	return &dell{c: c}, nil
}

// Note that Dell vmcli only supports 1 active session, but supports mounting both
// a floppy/usb and cdrom within the single session.
func (m *dell) Mount(media *VirtualMedia) error {
	args := []string{"-r", m.c.RemoteIP(), "-u", m.c.Username, "-p", m.c.Password}

	devices := map[string]string{
		"-c": media.CdromImage,
		"-f": media.FloppyImage,
	}

	for flag, file := range devices {
		if file != "" {
			args = append(args, flag, file)
		}
	}

	m.Cmd = exec.Command(m.cli(), args...)

	return m.start()
}

func (m *dell) cli() string {
	return defaultDellVmcli // TODO: racvmcli for v5
}
