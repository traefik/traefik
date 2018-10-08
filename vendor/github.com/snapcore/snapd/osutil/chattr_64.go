// -*- Mode: Go; indent-tabs-mode: t -*-
// +build arm64 amd64 ppc64le s390x

/*
 * Copyright (C) 2016 Canonical Ltd
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

const (
	// There is a logic to these but I don't care to implement it all.
	// If you do, chase them from linux/fs.h
	_FS_IOC_GETFLAGS = uintptr(0x80086601)
	_FS_IOC_SETFLAGS = uintptr(0x40086602)
)
