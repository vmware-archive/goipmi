// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"

	"code.google.com/p/go.crypto/ssh"
)

// Connection properties for a Client
type Connection struct {
	Path      string
	Hostname  string
	Port      int
	Username  string
	Password  string
	Interface string
	SSHPort   int
	SSHOpts   []string
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

func (c *Connection) sshPort() string {
	if c.SSHPort == 0 {
		return "22"
	}
	return strconv.Itoa(c.SSHPort)
}

// DialSSH calls ssh.Dial with the given Connection
func (c *Connection) DialSSH() (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: c.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
	}

	address := net.JoinHostPort(c.Hostname, c.sshPort())
	return ssh.Dial("tcp", address, config)
}

// RunSSH runs the given commands over ssh
func (c *Connection) RunSSH(commands ...string) error {
	if len(c.SSHOpts) > 0 {
		// go ssh and HP iLO v2 have no cipher in common..
		// assuming for now that SSHOpts includes '-i'
		for _, cmd := range commands {
			opts := []string{
				"-p", c.sshPort(),
				"-o", "UserKnownHostsFile=/dev/null",
				"-o", "StrictHostKeyChecking=no",
				"-o", "BatchMode=yes",
				c.Hostname,
				cmd,
			}
			ssh := exec.Command("ssh", append(c.SSHOpts, opts...)...)
			if err := ssh.Run(); err != nil {
				return err
			}
		}
		return nil
	}

	// with no SSHOpts, assume we can do password auth with pure go ssh client
	client, err := c.DialSSH()
	if err != nil {
		return err
	}
	defer client.Close()

	for _, cmd := range commands {
		session, err := client.NewSession()
		if err != nil {
			return err
		}

		err = session.Run(cmd)
		_ = session.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
