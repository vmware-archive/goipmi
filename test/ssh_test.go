// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package test

import (
	"github.com/vmware/goipmi"
	"testing"

	"code.google.com/p/go.crypto/ssh"
	"github.com/stretchr/testify/assert"
)

func TestDialSSH(t *testing.T) {
	var lastCmd string
	var status int

	c := &ipmi.Connection{
		Username: "multi",
		Password: "none",
		Hostname: "127.0.0.1",
	}

	wg := StartSSHExecServer(c, func(ch ssh.Channel, cmd string) int {
		lastCmd = cmd
		return status
	})

	client, err := c.DialSSH()
	assert.NoError(t, err)

	tests := []struct {
		cmd    string
		status int
	}{
		{"cal", 1},
		{"date", 0},
	}

	for _, test := range tests {
		status = test.status

		session, err := client.NewSession()
		assert.NoError(t, err)

		err = session.Run(test.cmd)
		if test.status == 0 {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
		assert.Equal(t, test.cmd, lastCmd)
		session.Close()
	}

	client.Close()
	wg.Wait()
}

func TestRunSSH(t *testing.T) {
	conns := []*ipmi.Connection{
		// pure-go
		{
			Username: "gossh",
			Password: "none",
		},
		// exec ssh
		{
			Username: "ssh",
			Password: "",
			SSHOpts:  []string{"-o", "SendEnv=HOME"},
		},
	}

	tests := []struct {
		cmd    string
		status int
	}{
		{"cal", 1},
		{"date", 0},
	}

	for _, c := range conns {
		for _, test := range tests {
			cmds := []string{}
			status := test.status

			c.Hostname = "127.0.0.1"

			wg := StartSSHExecServer(c, func(ch ssh.Channel, cmd string) int {
				cmds = append(cmds, cmd)
				return status
			})

			err := c.RunSSH(test.cmd)
			if test.status == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, 1, len(cmds))
			assert.Equal(t, test.cmd, cmds[0])

			wg.Wait()
		}
	}
}
