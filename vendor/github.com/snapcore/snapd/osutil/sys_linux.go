// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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
	"syscall"
	"unsafe"
)

// Symlinkat is a direct pass-through to the symlinkat(2) system call.
func Symlinkat(target string, dirfd int, linkpath string) error {
	targetPtr, err := syscall.BytePtrFromString(target)
	if err != nil {
		return err
	}
	linkpathPtr, err := syscall.BytePtrFromString(linkpath)
	if err != nil {
		return err
	}
	_, _, errno := syscall.Syscall(syscall.SYS_SYMLINKAT, uintptr(unsafe.Pointer(targetPtr)), uintptr(dirfd), uintptr(unsafe.Pointer(linkpathPtr)))
	if errno != 0 {
		return errno
	}
	return nil
}

// Readlinkat is a direct pass-through to the readlinkat(2) system call.
func Readlinkat(dirfd int, path string, buf []byte) (n int, err error) {
	var zero uintptr

	pathPtr, err := syscall.BytePtrFromString(path)
	if err != nil {
		return 0, err
	}
	var bufPtr unsafe.Pointer
	if len(buf) > 0 {
		bufPtr = unsafe.Pointer(&buf[0])
	} else {
		bufPtr = unsafe.Pointer(&zero)
	}
	r0, _, errno := syscall.Syscall6(syscall.SYS_READLINKAT, uintptr(dirfd), uintptr(unsafe.Pointer(pathPtr)), uintptr(bufPtr), uintptr(len(buf)), 0, 0)
	n = int(r0)
	if errno != 0 {
		return 0, errno
	}
	return n, nil
}
