// Copyright 2017 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"syscall"

	"golang.org/x/time/rate"
)

type runnerStresser struct {
	cmd     *exec.Cmd
	cmdStr  string
	args    []string
	rl      *rate.Limiter
	reqRate int

	errc  chan error
	donec chan struct{}
}

func newRunnerStresser(cmdStr string, args []string, rl *rate.Limiter, reqRate int) *runnerStresser {
	rl.SetLimit(rl.Limit() - rate.Limit(reqRate))
	return &runnerStresser{
		cmdStr:  cmdStr,
		args:    args,
		rl:      rl,
		reqRate: reqRate,
		errc:    make(chan error, 1),
		donec:   make(chan struct{}),
	}
}

func (rs *runnerStresser) setupOnce() (err error) {
	if rs.cmd != nil {
		return nil
	}

	rs.cmd = exec.Command(rs.cmdStr, rs.args...)
	stderr, err := rs.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		defer close(rs.donec)
		out, err := ioutil.ReadAll(stderr)
		if err != nil {
			rs.errc <- err
		} else {
			rs.errc <- fmt.Errorf("(%v %v) stderr %v", rs.cmdStr, rs.args, string(out))
		}
	}()

	return rs.cmd.Start()
}

func (rs *runnerStresser) Stress() (err error) {
	if err = rs.setupOnce(); err != nil {
		return err
	}
	return syscall.Kill(rs.cmd.Process.Pid, syscall.SIGCONT)
}

func (rs *runnerStresser) Pause() {
	syscall.Kill(rs.cmd.Process.Pid, syscall.SIGSTOP)
}

func (rs *runnerStresser) Close() {
	syscall.Kill(rs.cmd.Process.Pid, syscall.SIGINT)
	rs.cmd.Wait()
	<-rs.donec
	rs.rl.SetLimit(rs.rl.Limit() + rate.Limit(rs.reqRate))
}

func (rs *runnerStresser) ModifiedKeys() int64 {
	return 1
}

func (rs *runnerStresser) Checker() Checker {
	return &runnerChecker{rs.errc}
}
