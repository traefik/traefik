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

// #include <stdlib.h>
// #include <sys/types.h>
// #include <grp.h>
// #include <unistd.h>
import "C"

import (
	"fmt"
	"os/user"
	"strconv"
	"syscall"
	"unsafe"
)

// hrm, user.LookupGroup() doesn't exist yet:
// https://github.com/golang/go/issues/2617
//
// Use implementation from upcoming releases:
// https://golang.org/src/os/user/lookup_unix.go
func lookupGroup(groupname string) (string, error) {
	var grp C.struct_group
	var result *C.struct_group

	buf := alloc(groupBuffer)
	defer buf.free()
	cname := C.CString(groupname)
	defer C.free(unsafe.Pointer(cname))

	err := retryWithBuffer(buf, func() syscall.Errno {
		return syscall.Errno(C.getgrnam_r(cname,
			&grp,
			(*C.char)(buf.ptr),
			C.size_t(buf.size),
			&result))
	})
	if err != nil {
		return "", fmt.Errorf("group: lookup groupname %s: %v", groupname, err)
	}
	if result == nil {
		return "", fmt.Errorf("group: unknown group %s", groupname)
	}
	return strconv.Itoa(int(grp.gr_gid)), nil
}

type bufferKind C.int

const (
	groupBuffer = bufferKind(C._SC_GETGR_R_SIZE_MAX)
)

func (k bufferKind) initialSize() C.size_t {
	sz := C.sysconf(C.int(k))
	if sz == -1 {
		// DragonFly and FreeBSD do not have _SC_GETPW_R_SIZE_MAX.
		// Additionally, not all Linux systems have it, either. For
		// example, the musl libc returns -1.
		return 1024
	}
	if !isSizeReasonable(int64(sz)) {
		// Truncate.  If this truly isn't enough, retryWithBuffer will error on the first run.
		return maxBufferSize
	}
	return C.size_t(sz)
}

type memBuffer struct {
	ptr  unsafe.Pointer
	size C.size_t
}

func alloc(kind bufferKind) *memBuffer {
	sz := kind.initialSize()
	return &memBuffer{
		ptr:  C.malloc(sz),
		size: sz,
	}
}

func (mb *memBuffer) resize(newSize C.size_t) {
	mb.ptr = C.realloc(mb.ptr, newSize)
	mb.size = newSize
}

func (mb *memBuffer) free() {
	C.free(mb.ptr)
}

// retryWithBuffer repeatedly calls f(), increasing the size of the
// buffer each time, until f succeeds, fails with a non-ERANGE error,
// or the buffer exceeds a reasonable limit.
func retryWithBuffer(buf *memBuffer, f func() syscall.Errno) error {
	for {
		errno := f()
		if errno == 0 {
			return nil
		} else if errno != syscall.ERANGE {
			return errno
		}
		newSize := buf.size * 2
		if !isSizeReasonable(int64(newSize)) {
			return fmt.Errorf("internal buffer exceeds %d bytes", maxBufferSize)
		}
		buf.resize(newSize)
	}
}

const maxBufferSize = 1 << 20

func isSizeReasonable(sz int64) bool {
	return sz > 0 && sz <= maxBufferSize
}

// end code from https://golang.org/src/os/user/lookup_unix.go

// FindUid returns the identifier of the given UNIX user name.
func FindUid(username string) (uint64, error) {
	user, err := user.Lookup(username)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(user.Uid, 10, 64)
}

// FindGid returns the identifier of the given UNIX group name.
func FindGid(group string) (uint64, error) {
	// In golang 1.8 we can use the built-in function like this:
	//group, err := user.LookupGroup(group)
	group, err := lookupGroup(group)
	if err != nil {
		return 0, err
	}

	// In golang 1.8 we can parse the group.Gid string instead.
	//return strconv.ParseUint(group.Gid, 10, 64)
	return strconv.ParseUint(group, 10, 64)
}
