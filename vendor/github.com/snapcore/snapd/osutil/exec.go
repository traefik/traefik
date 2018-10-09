// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package osutil

import (
	"bufio"
	"bytes"
	"debug/elf"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gopkg.in/tomb.v2"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/strutil"
)

func parseCoreLdSoConf(confPath string) []string {
	root := filepath.Join(dirs.SnapMountDir, "/core/current")

	f, err := os.Open(filepath.Join(root, confPath))
	if err != nil {
		return nil
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "#"):
			// nothing
		case strings.TrimSpace(line) == "":
			// nothing
		case strings.HasPrefix(line, "include "):
			l := strings.SplitN(line, "include ", 2)
			files, err := filepath.Glob(filepath.Join(root, l[1]))
			if err != nil {
				return nil
			}
			for _, f := range files {
				out = append(out, parseCoreLdSoConf(f[len(root):])...)
			}
		default:
			out = append(out, filepath.Join(root, line))
		}

	}
	if err := scanner.Err(); err != nil {
		return nil
	}

	return out
}

func elfInterp(cmd string) (string, error) {
	el, err := elf.Open(cmd)
	if err != nil {
		return "", err
	}
	defer el.Close()

	for _, prog := range el.Progs {
		if prog.Type == elf.PT_INTERP {
			r := prog.Open()
			interp, err := ioutil.ReadAll(r)
			if err != nil {
				return "", nil
			}

			return string(bytes.Trim(interp, "\x00")), nil
		}
	}

	return "", fmt.Errorf("cannot find PT_INTERP header")
}

// CommandFromCore runs a command from the core snap using the proper
// interpreter and library paths.
//
// At the moment it can only run ELF files, expects a standard ld.so
// interpreter, and can't handle RPATH.
func CommandFromCore(name string, cmdArgs ...string) (*exec.Cmd, error) {
	root := filepath.Join(dirs.SnapMountDir, "/core/current")

	cmdPath := filepath.Join(root, name)
	interp, err := elfInterp(cmdPath)
	if err != nil {
		return nil, err
	}
	coreLdSo := filepath.Join(root, interp)
	// we cannot use EvalSymlink here because we need to resolve
	// relative and absolute symlinks differently. A absolute
	// symlink is relative to root of the core snap.
	seen := map[string]bool{}
	for IsSymlink(coreLdSo) {
		link, err := os.Readlink(coreLdSo)
		if err != nil {
			return nil, err
		}
		if filepath.IsAbs(link) {
			coreLdSo = filepath.Join(root, link)
		} else {
			coreLdSo = filepath.Join(filepath.Dir(coreLdSo), link)
		}
		if seen[coreLdSo] {
			return nil, fmt.Errorf("cannot run command from core: symlink cycle found")
		}
		seen[coreLdSo] = true
	}

	ldLibraryPathForCore := parseCoreLdSoConf("/etc/ld.so.conf")

	ldSoArgs := []string{"--library-path", strings.Join(ldLibraryPathForCore, ":"), cmdPath}
	allArgs := append(ldSoArgs, cmdArgs...)
	return exec.Command(coreLdSo, allArgs...), nil
}

var cmdWaitTimeout = 5 * time.Second

// KillProcessGroup kills the process group associated with the given command.
//
// If the command hasn't had Setpgid set in its SysProcAttr, you'll probably end
// up killing yourself.
func KillProcessGroup(cmd *exec.Cmd) error {
	pgid, err := syscallGetpgid(cmd.Process.Pid)
	if err != nil {
		return err
	}
	if pgid == 1 {
		return fmt.Errorf("cannot kill pgid 1")
	}
	return syscallKill(-pgid, syscall.SIGKILL)
}

// RunAndWait runs a command for the given argv with the given environ added to
// os.Environ, killing it if it reaches timeout, or if the tomb is dying.
func RunAndWait(argv []string, env []string, timeout time.Duration, tomb *tomb.Tomb) ([]byte, error) {
	if len(argv) == 0 {
		return nil, fmt.Errorf("internal error: osutil.RunAndWait needs non-empty argv")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("internal error: osutil.RunAndWait needs positive timeout")
	}
	if tomb == nil {
		return nil, fmt.Errorf("internal error: osutil.RunAndWait needs non-nil tomb")
	}

	command := exec.Command(argv[0], argv[1:]...)

	// setup a process group for the command so that we can kill parent
	// and children on e.g. timeout
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Env = append(os.Environ(), env...)

	// Make sure we can obtain stdout and stderror. Same buffer so they're
	// combined.
	buffer := strutil.NewLimitedBuffer(100, 10*1024)
	command.Stdout = buffer
	command.Stderr = buffer

	// Actually run the command.
	if err := command.Start(); err != nil {
		return nil, err
	}

	// add timeout handling
	killTimerCh := time.After(timeout)

	commandCompleted := make(chan struct{})
	var commandError error
	go func() {
		// Wait for hook to complete
		commandError = command.Wait()
		close(commandCompleted)
	}()

	var abortOrTimeoutError error
	select {
	case <-commandCompleted:
		// Command completed; it may or may not have been successful.
		return buffer.Bytes(), commandError
	case <-tomb.Dying():
		// Hook was aborted, process will get killed below
		abortOrTimeoutError = fmt.Errorf("aborted")
	case <-killTimerCh:
		// Max timeout reached, process will get killed below
		abortOrTimeoutError = fmt.Errorf("exceeded maximum runtime of %s", timeout)
	}

	// select above exited which means that aborted or killTimeout
	// was reached. Kill the command and wait for command.Wait()
	// to clean it up (but limit the wait with the cmdWaitTimer)
	if err := KillProcessGroup(command); err != nil {
		return nil, fmt.Errorf("cannot abort: %s", err)
	}
	select {
	case <-time.After(cmdWaitTimeout):
		// cmdWaitTimeout was reached, i.e. command.Wait() did not
		// finish in a reasonable amount of time, we can not use
		// buffer in this case so return without it.
		return nil, fmt.Errorf("%v, but did not stop", abortOrTimeoutError)
	case <-commandCompleted:
		// cmd.Wait came back from waiting the killed process
		break
	}
	fmt.Fprintf(buffer, "\n<%s>", abortOrTimeoutError)

	return buffer.Bytes(), abortOrTimeoutError
}

type waitingReader struct {
	reader io.Reader
	cmd    *exec.Cmd
}

func (r *waitingReader) Close() error {
	if r.cmd.Process != nil {
		r.cmd.Process.Kill()
	}
	return r.cmd.Wait()
}

func (r *waitingReader) Read(b []byte) (int, error) {
	n, err := r.reader.Read(b)
	if n == 0 && err == io.EOF {
		err = r.Close()
		if err == nil {
			return 0, io.EOF
		}
		return 0, err
	}
	return n, err
}

// StreamCommand runs a the named program with the given arguments,
// streaming its standard output over the returned io.ReadCloser.
//
// The program will run until EOF is reached (at which point the
// ReadCloser is closed), or until the ReadCloser is explicitly closed.
func StreamCommand(name string, args ...string) (io.ReadCloser, error) {
	cmd := exec.Command(name, args...)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &waitingReader{reader: pipe, cmd: cmd}, nil
}
