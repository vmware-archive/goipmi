// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"github.com/vmware/goipmi"
	"errors"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/testdata"
)

func sshPort(c *ipmi.Connection) string {
	if c.SSHPort == 0 {
		return "22"
	}
	return strconv.Itoa(c.SSHPort)
}

func sshCommand(args ...string) string {
	return strings.Join(args, " ")
}

func dialSSH(c *ipmi.Connection) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: c.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
	}

	address := net.JoinHostPort(c.Hostname, sshPort(c))
	return ssh.Dial("tcp", address, config)
}

func runSSH(c *ipmi.Connection, cmds ...string) error {
	if len(c.SSHOpts) > 0 {
		// go ssh and HP iLO v2 have no cipher in common..
		// assuming for now that SSHOpts includes '-i'
		for _, cmd := range cmds {
			opts := []string{
				"-p", sshPort(c),
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
	client, err := dialSSH(c)
	if err != nil {
		return err
	}
	defer ioclose(client)

	for _, cmd := range cmds {
		session, err := client.NewSession()
		if err != nil {
			return err
		}

		err = session.Run(cmd)
		ioclose(session)
		if err != nil {
			return err
		}
	}

	return nil
}

// run an ssh server that accepts a single connection and handles only "exec" requests.
// handler callback is given the ssh.Channel for io and command string.  The handler return
// value is propagated to the client via "exit-status".
func sshTestExecServer(ic *ipmi.Connection, handler func(ssh.Channel, string) int) *sync.WaitGroup {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == ic.Username && string(pass) == ic.Password {
				return nil, nil
			}
			return nil, errors.New("auth fail")
		},
	}
	if ic.Password == "" {
		config.NoClientAuth = true
	}
	signer, err := ssh.ParsePrivateKey(testdata.PEMBytes["dsa"])
	if err != nil {
		panic(err)
	}
	config.AddHostKey(signer)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	ic.SSHPort = l.Addr().(*net.TCPAddr).Port

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer ioclose(l)
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		defer ioclose(c)

		conn, chans, _, err := ssh.NewServerConn(c, config)
		if err != nil {
			panic(err)
		}
		defer ioclose(conn)

		for newChan := range chans {
			ch, requests, err := newChan.Accept()
			if err != nil {
				panic(err)
			}

			for req := range requests {
				if req.Type != "exec" {
					panic(req.Type)
				}
				_ = req.Reply(true, nil)
				rc := handler(ch, string(req.Payload[4:]))
				status := struct {
					Status uint32
				}{uint32(rc)}
				_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(&status))
				ioclose(ch) // 1 exec per session (see ssh.Session.Start)
			}
		}

		wg.Done()
	}()
	return &wg
}
