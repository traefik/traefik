// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
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

package snap

import "fmt"

type AlreadyInstalledError struct {
	Snap string
}

func (e AlreadyInstalledError) Error() string {
	return fmt.Sprintf("snap %q is already installed", e.Snap)
}

type NotInstalledError struct {
	Snap string
	Rev  Revision
}

func (e NotInstalledError) Error() string {
	if e.Rev.Unset() {
		return fmt.Sprintf("snap %q is not installed", e.Snap)
	}
	return fmt.Sprintf("revision %s of snap %q is not installed", e.Rev, e.Snap)
}

type NotSnapError struct {
	Path string
}

func (e NotSnapError) Error() string {
	return fmt.Sprintf("%q is not a snap or snapdir", e.Path)
}
