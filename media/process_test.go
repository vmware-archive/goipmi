// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessNoError(t *testing.T) {
	p := &process{
		Cmd: exec.Command("sh", "-c", "sleep 1"),
	}
	err := p.start()
	assert.NoError(t, err)
	err = p.UnMount()
	assert.NoError(t, err)
	assert.False(t, p.ProcessState.Exited())
}

func TestProcessExitError(t *testing.T) {
	p := &process{
		Cmd: exec.Command("sh", "-c", "exit 2"),
	}
	err := p.start()
	assert.NoError(t, err)
	p.wg.Wait()
	err = p.waitState()
	assert.Error(t, err)
	assert.True(t, p.ProcessState.Exited())
}

func TestProcessEnoent(t *testing.T) {
	p := &process{
		Cmd: exec.Command("__ENOENT__"),
	}
	err := p.start()
	assert.Error(t, err)
}
