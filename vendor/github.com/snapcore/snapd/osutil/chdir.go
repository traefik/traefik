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
	"os"
)

// ChDir runs runs "f" inside the given directory
// Note that this will only work reliable in a single-threaded context.
func ChDir(newDir string, f func() error) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(newDir); err != nil {
		return err
	}
	defer os.Chdir(cwd)
	return f()
}
