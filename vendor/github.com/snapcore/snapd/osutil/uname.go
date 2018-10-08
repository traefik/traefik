// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2018 Canonical Ltd
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

var KernelVersion = kernelVersion

// We have to implement separate functions for the kernel version and the
// machine name at the moment as the utsname struct is either using int8
// or uint8 depending on the architecture the code is built for. As there
// is no easy way to generalise this implements the same code twice (other
// than going via unsafe.Pointer, or interfaces, which are overkill). The
// way to get this solved is by using []byte inside the utsname struct
// instead of []int8/[]uint8. See https://github.com/golang/go/issues/20753
// for details.

func kernelVersion() string {
	u, err := uname()
	if err != nil {
		return "unknown"
	}

	// Release is more informative than Version.
	buf := make([]byte, len(u.Release))
	for i, c := range u.Release {
		if c == 0 {
			buf = buf[:i]
			break
		}
		// c can be uint8 or int8 depending on arch (see comment above)
		buf[i] = byte(c)
	}

	return string(buf)
}

func MachineName() string {
	u, err := uname()
	if err != nil {
		return "unknown"
	}

	buf := make([]byte, len(u.Machine))
	for i, c := range u.Machine {
		if c == 0 {
			buf = buf[:i]
			break
		}
		// c can be uint8 or int8 depending on arch (see comment above)
		buf[i] = byte(c)
	}

	return string(buf)
}

// MockKernelVersion replaces the function that returns the kernel version string.
func MockKernelVersion(version string) (restore func()) {
	old := KernelVersion
	KernelVersion = func() string { return version }
	return func() {
		KernelVersion = old
	}
}
