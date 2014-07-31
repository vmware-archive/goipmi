// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"fmt"
	"log"
	"net/url"
)

// VirtualMedia types
const (
	ISO = "iso" // CD-ROM/DVD-ROM
	IMG = "img" // Floppy/USB
)

// The Media interface defines methods for using remote virtual media
type Media interface {
	Mount(VirtualMedia) error
	UnMount() error
}

// VirtualDevice contains the local Path to media, Boot (once) flag and
// optional URL where the file can be accessed by the BMC
type VirtualDevice struct {
	Path string
	Boot bool
	URL  *url.URL
}

// VirtualMedia specifies paths for mounting cdrom and/or floppy/usb images.
// BootDevice can be set to specify an image type to boot from on next boot.
type VirtualMedia map[string]*VirtualDevice

// New creates a Media instance based on the ipmi device id.
// A error is returned if the device is not supported.
func New(c *ipmi.Client, id *ipmi.DeviceIDResponse) (Media, error) {
	switch id.ManufacturerID {
	case ipmi.OemDell:
		return newDellMedia(c)
	case ipmi.OemHP:
		return newHPMedia(c)
	case ipmi.OemSupermicro:
		return newSupermicroMedia(c)
	default:
		return nil, fmt.Errorf("OEM not supported: %s", id.ManufacturerID)
	}
}

// Boot a machine with the given image.
// The image will be mounted via remote virtual media, bios flag set to boot once
// for the appropriate media device and machine will be power cycled.
// The given handler will be called after the power cycled.
func Boot(conn *ipmi.Connection, vm VirtualMedia, handler func(*ipmi.Client) error) error {
	c, err := ipmi.NewClient(conn)
	if err != nil {
		return err
	}
	if err := c.Open(); err != nil {
		return err
	}
	defer c.Close()

	id, err := c.DeviceID()
	if err != nil {
		return err
	}

	m, err := New(c, id)
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

	if err := c.Control(ipmi.ControlPowerCycle); err != nil {
		return err
	}

	return handler(c)
}
