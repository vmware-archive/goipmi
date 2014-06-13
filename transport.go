// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import "fmt"

type transport interface {
	open() error
	close() error
	send(*Request, Response) error
}

func newTransport(c *Connection) (transport, error) {
	switch c.Interface {
	case "lan":
		return newLanTransport(c), nil
	default:
		return nil, fmt.Errorf("unsupported interface: %s", c.Interface)
	}
}
