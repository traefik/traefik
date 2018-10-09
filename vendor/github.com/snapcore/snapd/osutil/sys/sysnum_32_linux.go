// -*- Mode: Go; indent-tabs-mode: t -*-
// +build arm64 amd64 ppc64le s390x ppc

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

package sys

import "syscall"

// these are the constants for where getuid et al are already 32-bit

const (
	_SYS_GETUID  = syscall.SYS_GETUID
	_SYS_GETGID  = syscall.SYS_GETGID
	_SYS_GETEUID = syscall.SYS_GETEUID
	_SYS_GETEGID = syscall.SYS_GETEGID
)
