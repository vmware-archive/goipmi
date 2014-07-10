// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"fmt"
	"io"
	"log"
)

// The Media interface defines methods for using remote virtual media
type Media interface {
	Mount(*VirtualMedia) error
	UnMount() error
}

// VirtualMedia specifies paths for mounting cdrom and/or floppy/usb images.
// BootDevice can be set to specify an image type to boot from on next boot.
type VirtualMedia struct {
	BootDevice  ipmi.BootDevice
	CdromImage  string
	FloppyImage string
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
func Boot(conn *ipmi.Connection, vm *VirtualMedia, handler func(*ipmi.Client) error) error {
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

	if err := m.Mount(vm); err != nil {
		return err
	}

	defer func() {
		if err := m.UnMount(); err != nil {
			log.Printf("Error unmounting: %s", err)
		}
	}()

	if vm.BootDevice != ipmi.BootDeviceNone {
		if err := c.SetBootDevice(vm.BootDevice); err != nil {
			return err
		}
	}

	if err := c.Control(ipmi.ControlPowerCycle); err != nil {
		return err
	}

	return handler(c)
}

// avoid common errcheck warning
func ioclose(c io.Closer) {
	_ = c.Close()
}
