// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import "fmt"

type transport interface {
	open() error
	close() error
	send(*Request, Response) error
	// Console enters Serial Over LAN mode
	Console() error
}

func newTransport(c *Connection) (transport, error) {
	switch c.Interface {
	case "lan":
		if c.Path == "" {
			return newLanTransport(c), nil
		}
		return newToolTransport(c), nil
	case "lanplus":
		return newToolTransport(c), nil
	default:
		return nil, fmt.Errorf("unsupported interface: %s", c.Interface)
	}
}
