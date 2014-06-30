// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package media

import (
	"os/exec"
	"sync"
	"syscall"
)

type process struct {
	*exec.Cmd
	wg    sync.WaitGroup
	state error
}

func (p *process) wait() {
	p.state = p.Wait()
	p.wg.Done()
}

func (p *process) start() error {
	err := p.Start()
	if err != nil {
		return err
	}
	p.wg.Add(1)
	go p.wait()
	return nil
}

func (p *process) waitState() error {
	if err, ok := p.state.(*exec.ExitError); ok {
		status := err.ProcessState.Sys().(syscall.WaitStatus)
		if status.Signaled() && status.Signal() == syscall.SIGKILL {
			// expected - we just Killed the process
			return nil
		}
	}

	return p.state
}

func (p *process) UnMount() error {
	_ = p.Process.Kill()
	p.wg.Wait()
	return p.waitState()
}
