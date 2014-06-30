// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"fmt"
	"io"
	"log"
	"path/filepath"
)

// The Media interface defines methods for using remote virtual media
type Media interface {
	BootDevice(image string) ipmi.BootDevice
	Mount(string) error
	UnMount() error
}

// New creates a Media instance based on the ipmi device id.
// A error is returned if the device is not supported.
func New(c *ipmi.Connection, id *ipmi.DeviceIDResponse) (Media, error) {
	switch id.ManufacturerID {
	case ipmi.OemDell:
		return newDellMedia(c, id)
	case ipmi.OemHP:
		return newHPMedia(c, id)
	default:
		return nil, fmt.Errorf("OEM not supported: %s", id.ManufacturerID)
	}
}

// Boot a machine with the given image.
// The image will be mounted via remote virtual media, bios flag set to boot once
// for the appropriate media device and machine will be power cycled.
// The given handler will be called after the power cycled.
func Boot(conn *ipmi.Connection, image string, handler func() error) error {
	c, err := ipmi.NewClient(conn)
	if err != nil {
		return err
	}
	if err := c.Open(); err != nil {
		return err
	}
	defer ioclose(c)

	id, err := c.DeviceID()
	if err != nil {
		return err
	}

	m, err := New(conn, id)
	if err != nil {
		return err
	}

	if err := m.Mount(image); err != nil {
		return err
	}

	defer func() {
		if err := m.UnMount(); err != nil {
			log.Printf("Error unmounting %s: %s", image, err)
		}
	}()

	if dev := m.BootDevice(image); dev != ipmi.BootDeviceNone {
		if err := c.SetBootDevice(dev); err != nil {
			return err
		}
	}

	if err := c.Control(ipmi.ControlPowerCycle); err != nil {
		return err
	}

	return handler()
}

// avoid common errcheck warning
func ioclose(c io.Closer) {
	_ = c.Close()
}

type bootDevice struct{}

func (bootDevice) BootDevice(image string) ipmi.BootDevice {
	if isISO(image) {
		return ipmi.BootDeviceRemoteCdrom
	}
	return ipmi.BootDeviceRemoteFloppy
}

func isISO(image string) bool {
	return filepath.Ext(image) == ".iso"
}
