// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CopyFlag is used to tweak the behaviour of CopyFile
type CopyFlag uint8

const (
	// CopyFlagDefault is the default behaviour
	CopyFlagDefault CopyFlag = 0
	// CopyFlagSync does a sync after copying the files
	CopyFlagSync CopyFlag = 1 << iota
	// CopyFlagOverwrite overwrites the target if it exists
	CopyFlagOverwrite
	// CopyFlagPreserveAll preserves mode,owner,time attributes
	CopyFlagPreserveAll
)

var (
	openfile = doOpenFile
	copyfile = doCopyFile
)

type fileish interface {
	Close() error
	Sync() error
	Fd() uintptr
	Stat() (os.FileInfo, error)
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

func doOpenFile(name string, flag int, perm os.FileMode) (fileish, error) {
	return os.OpenFile(name, flag, perm)
}

// CopyFile copies src to dst
func CopyFile(src, dst string, flags CopyFlag) (err error) {
	if flags&CopyFlagPreserveAll != 0 {
		// Our native copy code does not preserve all attributes
		// (yet). If the user needs this functionatlity we just
		// fallback to use the system's "cp" binary to do the copy.
		if err := runCpPreserveAll(src, dst, "copy all"); err != nil {
			return err
		}
		if flags&CopyFlagSync != 0 {
			return runSync()
		}
		return nil
	}

	fin, err := openfile(src, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("unable to open %s: %v", src, err)
	}
	defer func() {
		if cerr := fin.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("when closing %s: %v", src, cerr)
		}
	}()

	fi, err := fin.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat %s: %v", src, err)
	}

	outflags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if flags&CopyFlagOverwrite == 0 {
		outflags |= os.O_EXCL
	}

	fout, err := openfile(dst, outflags, fi.Mode())
	if err != nil {
		return fmt.Errorf("unable to create %s: %v", dst, err)
	}
	defer func() {
		if cerr := fout.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("when closing %s: %v", dst, cerr)
		}
	}()

	if err := copyfile(fin, fout, fi); err != nil {
		return fmt.Errorf("unable to copy %s to %s: %v", src, dst, err)
	}

	if flags&CopyFlagSync != 0 {
		if err = fout.Sync(); err != nil {
			return fmt.Errorf("unable to sync %s: %v", dst, err)
		}
	}

	return nil
}

func runCmd(cmd *exec.Cmd, errdesc string) error {
	if output, err := cmd.CombinedOutput(); err != nil {
		output = bytes.TrimSpace(output)
		if exitCode, err := ExitCode(err); err == nil {
			return &CopySpecialFileError{
				desc:     errdesc,
				exitCode: exitCode,
				output:   output,
			}
		}
		return &CopySpecialFileError{
			desc:   errdesc,
			err:    err,
			output: output,
		}
	}

	return nil
}

func runSync(args ...string) error {
	return runCmd(exec.Command("sync", args...), "sync")
}

func runCpPreserveAll(path, dest, errdesc string) error {
	return runCmd(exec.Command("cp", "-av", path, dest), errdesc)
}

// CopySpecialFile is used to copy all the things that are not files
// (like device nodes, named pipes etc)
func CopySpecialFile(path, dest string) error {
	if err := runCpPreserveAll(path, dest, "copy device node"); err != nil {
		return err
	}
	return runSync(filepath.Dir(dest))
}

// CopySpecialFileError is returned if a special file copy fails
type CopySpecialFileError struct {
	desc     string
	exitCode int
	output   []byte
	err      error
}

func (e CopySpecialFileError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("failed to %s: %q (%v)", e.desc, e.output, e.exitCode)
	}

	return fmt.Sprintf("failed to %s: %q (%v)", e.desc, e.output, e.err)
}
