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
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// the default filesystem based backstore for assertions

const (
	assertionsLayoutVersion = "v0"
	assertionsRoot          = "asserts-" + assertionsLayoutVersion
)

type filesystemBackstore struct {
	top string
	mu  sync.RWMutex
}

// OpenFSBackstore opens a filesystem backed assertions backstore under path.
func OpenFSBackstore(path string) (Backstore, error) {
	top := filepath.Join(path, assertionsRoot)
	err := ensureTop(top)
	if err != nil {
		return nil, err
	}
	return &filesystemBackstore{top: top}, nil
}

// guarantees that result assertion is of the expected type (both in the AssertionType and go type sense)
func (fsbs *filesystemBackstore) readAssertion(assertType *AssertionType, diskPrimaryPath string) (Assertion, error) {
	encoded, err := readEntry(fsbs.top, assertType.Name, diskPrimaryPath)
	if os.IsNotExist(err) {
		return nil, errNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("broken assertion storage, cannot read assertion: %v", err)
	}
	assert, err := Decode(encoded)
	if err != nil {
		return nil, fmt.Errorf("broken assertion storage, cannot decode assertion: %v", err)
	}
	if assert.Type() != assertType {
		return nil, fmt.Errorf("assertion that is not of type %q under their storage tree", assertType.Name)
	}
	// because of Decode() construction assert has also the expected go type
	return assert, nil
}

func (fsbs *filesystemBackstore) pickLatestAssertion(assertType *AssertionType, diskPrimaryPaths []string, maxFormat int) (a Assertion, er error) {
	for _, diskPrimaryPath := range diskPrimaryPaths {
		fn := filepath.Base(diskPrimaryPath)
		parts := strings.SplitN(fn, ".", 2)
		formatnum := 0
		if len(parts) == 2 {
			var err error
			formatnum, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid active assertion filename: %q", fn)
			}
		}
		if formatnum <= maxFormat {
			a1, err := fsbs.readAssertion(assertType, diskPrimaryPath)
			if err != nil {
				return nil, err
			}
			if a == nil || a1.Revision() > a.Revision() {
				a = a1
			}
		}
	}
	if a == nil {
		return nil, errNotFound
	}
	return a, nil
}

func diskPrimaryPathComps(primaryPath []string, active string) []string {
	n := len(primaryPath)
	comps := make([]string, n+1)
	// safety against '/' etc
	for i, comp := range primaryPath {
		comps[i] = url.QueryEscape(comp)
	}
	comps[n] = active
	return comps
}

func (fsbs *filesystemBackstore) currentAssertion(assertType *AssertionType, primaryPath []string, maxFormat int) (Assertion, error) {
	var a Assertion
	namesCb := func(relpaths []string) error {
		var err error
		a, err = fsbs.pickLatestAssertion(assertType, relpaths, maxFormat)
		if err == errNotFound {
			return nil
		}
		return err
	}

	comps := diskPrimaryPathComps(primaryPath, "active*")
	assertTypeTop := filepath.Join(fsbs.top, assertType.Name)
	err := findWildcard(assertTypeTop, comps, namesCb)
	if err != nil {
		return nil, fmt.Errorf("broken assertion storage, looking for %s: %v", assertType.Name, err)
	}

	if a == nil {
		return nil, errNotFound
	}

	return a, nil
}

func (fsbs *filesystemBackstore) Put(assertType *AssertionType, assert Assertion) error {
	fsbs.mu.Lock()
	defer fsbs.mu.Unlock()

	primaryPath := assert.Ref().PrimaryKey

	curAssert, err := fsbs.currentAssertion(assertType, primaryPath, assertType.MaxSupportedFormat())
	if err == nil {
		curRev := curAssert.Revision()
		rev := assert.Revision()
		if curRev >= rev {
			return &RevisionError{Current: curRev, Used: rev}
		}
	} else if err != errNotFound {
		return err
	}

	formatnum := assert.Format()
	activeFn := "active"
	if formatnum > 0 {
		activeFn = fmt.Sprintf("active.%d", formatnum)
	}
	diskPrimaryPath := filepath.Join(diskPrimaryPathComps(primaryPath, activeFn)...)
	err = atomicWriteEntry(Encode(assert), false, fsbs.top, assertType.Name, diskPrimaryPath)
	if err != nil {
		return fmt.Errorf("broken assertion storage, cannot write assertion: %v", err)
	}
	return nil
}

func (fsbs *filesystemBackstore) Get(assertType *AssertionType, key []string, maxFormat int) (Assertion, error) {
	fsbs.mu.RLock()
	defer fsbs.mu.RUnlock()

	a, err := fsbs.currentAssertion(assertType, key, maxFormat)
	if err == errNotFound {
		return nil, &NotFoundError{Type: assertType}
	}
	return a, err
}

func (fsbs *filesystemBackstore) search(assertType *AssertionType, diskPattern []string, foundCb func(Assertion), maxFormat int) error {
	assertTypeTop := filepath.Join(fsbs.top, assertType.Name)
	candCb := func(diskPrimaryPaths []string) error {
		a, err := fsbs.pickLatestAssertion(assertType, diskPrimaryPaths, maxFormat)
		if err == errNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		foundCb(a)
		return nil
	}
	err := findWildcard(assertTypeTop, diskPattern, candCb)
	if err != nil {
		return fmt.Errorf("broken assertion storage, searching for %s: %v", assertType.Name, err)
	}
	return nil
}

func (fsbs *filesystemBackstore) Search(assertType *AssertionType, headers map[string]string, foundCb func(Assertion), maxFormat int) error {
	fsbs.mu.RLock()
	defer fsbs.mu.RUnlock()

	n := len(assertType.PrimaryKey)
	diskPattern := make([]string, n+1)
	for i, k := range assertType.PrimaryKey {
		keyVal := headers[k]
		if keyVal == "" {
			diskPattern[i] = "*"
		} else {
			diskPattern[i] = url.QueryEscape(keyVal)
		}
	}
	diskPattern[n] = "active*"

	candCb := func(a Assertion) {
		if searchMatch(a, headers) {
			foundCb(a)
		}
	}
	return fsbs.search(assertType, diskPattern, candCb, maxFormat)
}
