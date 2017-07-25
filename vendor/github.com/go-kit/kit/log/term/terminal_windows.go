// Based on ssh/terminal:
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package term

import (
	"io"
	"syscall"
	"unsafe"
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var (
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
)

// IsTerminal returns true if w writes to a terminal.
func IsTerminal(w io.Writer) bool {
	fw, ok := w.(fder)
	if !ok {
		return false
	}
	var st uint32
	r, _, e := syscall.Syscall(procGetConsoleMode.Addr(), 2, fw.Fd(), uintptr(unsafe.Pointer(&st)), 0)
	return r != 0 && e == 0
}
