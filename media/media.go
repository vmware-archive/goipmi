// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"fmt"
	"log"
	"net/url"
)

const (
	// ISO to be inserted into the virtual CD-ROM/DVD-ROM device
	ISO = "iso"
	// IMG to be inserted into the virtual Floppy/USB device
	IMG = "img"
)

var drivers = make(map[ipmi.OemID]func(c *ipmi.Client) (Driver, error))

// The Driver interface defines methods for managing virtual media
type Driver interface {
	Insert(DeviceMap) error
	Eject() error
}

// Device contains the local Path to media, Boot (once) flag and
// optional URL where the file can be accessed by the BMC
type Device struct {
	Path string
	Boot bool
	URL  *url.URL
}

// DeviceMap specifies paths for inserting iso and/or img files into virtual media devices.
// Boot flag can be set to specify an image type to boot from on next boot.
type DeviceMap map[string]*Device

// Register driver constructor for the given OemID
func Register(id ipmi.OemID, driver func(c *ipmi.Client) (Driver, error)) {
	drivers[id] = driver
}

// New creates a Driver instance based on the ipmi device id.
// A error is returned if the device is not supported.
func New(c *ipmi.Client, id ipmi.OemID) (Driver, error) {
	if driver, ok := drivers[id]; ok {
		return driver(c)
	}
	return nil, fmt.Errorf("OEM not supported: %s", id)
}

// Boot a machine with the given virtual media inserted.
// The images will be inserted via remote virtual media,
// bios flag set to boot once for the appropriate media device
// and machine will be power cycled.
// The given handler will be called after the power cycled.
func Boot(conn *ipmi.Connection, m DeviceMap, handler func(*ipmi.Client) error) error {
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

	d, err := New(c, id.ManufacturerID)
	if err != nil {
		return err
	}

	if err := d.Insert(m); err != nil {
		return err
	}

	defer func() {
		if err := d.Eject(); err != nil {
			log.Printf("Error ejecting: %s", err)
		}
	}()

	if err := c.Control(ipmi.ControlPowerCycle); err != nil {
		return err
	}

	return handler(c)
}
