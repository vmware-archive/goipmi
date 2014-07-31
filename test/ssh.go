// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package test

import (
	"github.com/vmware/goipmi"
	"errors"
	"net"
	"sync"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/testdata"
)

// StartSSHExecServer runs an ssh server that accepts a single connection and handles only "exec" requests.
// The handler callback is given the ssh.Channel for io and command string.  The handler return
// value is propagated to the client via "exit-status".
func StartSSHExecServer(ic *ipmi.Connection, handler func(ssh.Channel, string) int) *sync.WaitGroup {
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
		defer l.Close()
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		defer c.Close()

		conn, chans, _, err := ssh.NewServerConn(c, config)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		for newChan := range chans {
			ch, requests, err := newChan.Accept()
			if err != nil {
				panic(err)
			}

			for req := range requests {
				if req.Type == "env" {
					_ = req.Reply(true, nil)
					continue
				} else if req.Type != "exec" {
					panic(req.Type)
				}
				_ = req.Reply(true, nil)
				rc := handler(ch, string(req.Payload[4:]))
				status := struct {
					Status uint32
				}{uint32(rc)}
				_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(&status))
				_ = ch.Close() // 1 exec per session (see ssh.Session.Start)
			}
		}

		wg.Done()
	}()
	return &wg
}
