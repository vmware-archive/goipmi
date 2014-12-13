// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"fmt"
	"net"
)

// Connection properties for a Client
type Connection struct {
	Path      string
	Hostname  string
	Port      int
	Username  string
	Password  string
	Interface string
}

// RemoteIP returns the remote (bmc) IP address of the Connection
func (c *Connection) RemoteIP() string {
	if net.ParseIP(c.Hostname) == nil {
		addrs, err := net.LookupHost(c.Hostname)
		if err != nil && len(addrs) > 0 {
			return addrs[0]
		}
	}
	return c.Hostname
}

// LocalIP returns the local (client) IP address of the Connection
func (c *Connection) LocalIP() string {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", c.Hostname, c.Port))
	if err != nil {
		// don't bother returning an error, since this value will never
		// make it to the bmc if we can't connect to it.
		return c.Hostname
	}
	_ = conn.Close()
	host, _, _ := net.SplitHostPort(conn.LocalAddr().String())
	return host
}
