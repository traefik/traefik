// -*- Mode: Go; indent-tabs-mode: t -*-

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

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/snapcore/snapd/osutil/sys"
)

// XXX: we need to come back and fix this; this is a hack to unblock us.
//      As part of the fixing we should unify with the similar code in
//      cmd/snap-update-ns/utils.(*Secure).MkdirAll
var mu sync.Mutex

// MkdirAllChown is like os.MkdirAll but it calls os.Chown on any
// directories it creates.
func MkdirAllChown(path string, perm os.FileMode, uid sys.UserID, gid sys.GroupID) error {
	mu.Lock()
	defer mu.Unlock()
	return mkdirAllChown(filepath.Clean(path), perm, uid, gid)
}

func mkdirAllChown(path string, perm os.FileMode, uid sys.UserID, gid sys.GroupID) error {
	// split out so filepath.Clean isn't called twice for each inner path
	if s, err := os.Stat(path); err == nil {
		if s.IsDir() {
			return nil
		}

		// emulate os.MkdirAll
		return &os.PathError{
			Op:   "mkdir",
			Path: path,
			Err:  syscall.ENOTDIR,
		}
	}

	dir := filepath.Dir(path)
	if dir != "/" {
		if err := mkdirAllChown(dir, perm, uid, gid); err != nil {
			return err
		}
	}

	cand := path + ".mkdir-new"

	if err := os.Mkdir(cand, perm); err != nil && !os.IsExist(err) {
		return err
	}

	if err := sys.ChownPath(cand, uid, gid); err != nil {
		return err
	}

	if err := os.Rename(cand, path); err != nil {
		return err
	}

	fd, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer fd.Close()

	return fd.Sync()
}
