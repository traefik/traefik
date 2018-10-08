// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2016 Canonical Ltd
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

package asserts

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/snapcore/snapd/osutil"
)

// utilities to read/write fs entries

func ensureTop(path string) error {
	err := os.MkdirAll(path, 0775)
	if err != nil {
		return fmt.Errorf("cannot create assert storage root: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot create assert storage root: %v", err)
	}
	if info.Mode().Perm()&0002 != 0 {
		return fmt.Errorf("assert storage root unexpectedly world-writable: %v", path)
	}
	return nil
}

func atomicWriteEntry(data []byte, secret bool, top string, subpath ...string) error {
	fpath := filepath.Join(top, filepath.Join(subpath...))
	dir := filepath.Dir(fpath)
	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return err
	}
	fperm := 0664
	if secret {
		fperm = 0600
	}
	return osutil.AtomicWriteFile(fpath, data, os.FileMode(fperm), 0)
}

func entryExists(top string, subpath ...string) bool {
	fpath := filepath.Join(top, filepath.Join(subpath...))
	return osutil.FileExists(fpath)
}

func readEntry(top string, subpath ...string) ([]byte, error) {
	fpath := filepath.Join(top, filepath.Join(subpath...))
	return ioutil.ReadFile(fpath)
}
