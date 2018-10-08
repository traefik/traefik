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
	"syscall"
	"unsafe"
)

// Winsize is from tty_ioctl(4)
type Winsize struct {
	Row    uint16
	Col    uint16
	xpixel uint16 // unused
	Ypixel uint16 // unused
}

// GetTermWinsize performs the TIOCGWINSZ ioctl on stdout
func GetTermWinsize() (*Winsize, error) {
	ws := &Winsize{}
	x, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdout), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))

	if int(x) == -1 {
		// returning ws on error lets people that don't care
		// about the error get on with querying the struct
		// (which will be empty on error).
		return ws, errno
	}

	return ws, nil
}
