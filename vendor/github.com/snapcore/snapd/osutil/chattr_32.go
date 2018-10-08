// -*- Mode: Go; indent-tabs-mode: t -*-
// +build arm 386 ppc

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
	// these are actually _FS_IOC32 (i'm cheating)
	_FS_IOC_GETFLAGS = uintptr(0x80046601)
	_FS_IOC_SETFLAGS = uintptr(0x40046602)
)
