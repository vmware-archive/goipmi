// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package ipmi

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		should string
		conn   *Connection
		expect []string
	}{
		{
			"should use default port and interface",
			&Connection{"", "h", 0, "u", "p", ""},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lanplus"},
		},
		{
			"should append port",
			&Connection{"", "h", 1623, "u", "p", ""},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lanplus", "-p", "1623"},
		},
		{
			"should override default interface",
			&Connection{"", "h", 0, "u", "p", "lan"},
			[]string{"-H", "h", "-U", "u", "-P", "p", "-I", "lan"},
		},
	}

	for _, test := range tests {
		result := test.conn.options()
		assert.Equal(t, test.expect, result, test.should)
	}
}

type toolMock struct {
	Connection
}

func (m *toolMock) cleanup() {
	err := os.Remove(m.Path)
	if err != nil {
		log.Printf("Remove(%s): %s", m.Path, err)
	}
}

func newToolMock(output string, rc int) *toolMock {
	file, err := ioutil.TempFile("", "ipmitool")
	if err != nil {
		panic(err)
	}
	// just enough to test exec related code paths
	file.WriteString("#!/usr/bin/env bash\n")
	file.WriteString(fmt.Sprintf("echo -n '%s'\n", output))
	if rc != 0 {
		file.WriteString("echo 'Mock Failure' 1>&2\n")
		file.WriteString(fmt.Sprintf("exit %d\n", rc))
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	err = os.Chmod(file.Name(), 0755)
	if err != nil {
		panic(err)
	}

	return &toolMock{Connection{Path: file.Name()}}
}

func TestExec(t *testing.T) {
	tool := newToolMock("stuff", 0)
	defer tool.cleanup()
	output, err := tool.run("bmc", "info")

	assert.Nil(t, err)
	assert.Equal(t, "stuff", output)
}

func TestExecErr(t *testing.T) {
	tool := newToolMock("nothing", 1)
	defer tool.cleanup()
	output, err := tool.run("foo", "bar")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Mock Failure")
	assert.NotContains(t, err.Error(), "nothing")
	assert.Equal(t, "", output)
}

func TestChassisStatus(t *testing.T) {
	tool := newToolMock("21 10 40 54\n", 0)
	defer tool.cleanup()
	status, err := tool.ChassisStatus()

	assert.Nil(t, err)
	assert.Equal(t, true, status.IsSystemPowerOn())
}

func TestGetBootFlags(t *testing.T) {
	tool := newToolMock("01 05 80 3c 00 00 00", 0)
	defer tool.cleanup()
	flags, err := tool.GetBootFlags()

	assert.Nil(t, err)
	assert.Equal(t, BootDeviceFloppy, flags.BootDeviceSelector)
}
