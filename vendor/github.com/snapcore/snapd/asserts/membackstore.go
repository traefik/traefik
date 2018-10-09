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

package asserts

import (
	"errors"
	"sync"
)

type memoryBackstore struct {
	top memBSBranch
	mu  sync.RWMutex
}

type memBSNode interface {
	put(assertType *AssertionType, key []string, assert Assertion) error
	get(key []string, maxFormat int) (Assertion, error)
	search(hint []string, found func(Assertion), maxFormat int)
}

type memBSBranch map[string]memBSNode

type memBSLeaf map[string]map[int]Assertion

func (br memBSBranch) put(assertType *AssertionType, key []string, assert Assertion) error {
	key0 := key[0]
	down := br[key0]
	if down == nil {
		if len(key) > 2 {
			down = make(memBSBranch)
		} else {
			down = make(memBSLeaf)
		}
		br[key0] = down
	}
	return down.put(assertType, key[1:], assert)
}

func (leaf memBSLeaf) cur(key0 string, maxFormat int) (a Assertion) {
	for formatnum, a1 := range leaf[key0] {
		if formatnum <= maxFormat {
			if a == nil || a1.Revision() > a.Revision() {
				a = a1
			}
		}
	}
	return a
}

func (leaf memBSLeaf) put(assertType *AssertionType, key []string, assert Assertion) error {
	key0 := key[0]
	cur := leaf.cur(key0, assertType.MaxSupportedFormat())
	if cur != nil {
		rev := assert.Revision()
		curRev := cur.Revision()
		if curRev >= rev {
			return &RevisionError{Current: curRev, Used: rev}
		}
	}
	if _, ok := leaf[key0]; !ok {
		leaf[key0] = make(map[int]Assertion)
	}
	leaf[key0][assert.Format()] = assert
	return nil
}

// errNotFound is used internally by backends, it is converted to the richer
// NotFoundError only at their public interface boundary
var errNotFound = errors.New("assertion not found")

func (br memBSBranch) get(key []string, maxFormat int) (Assertion, error) {
	key0 := key[0]
	down := br[key0]
	if down == nil {
		return nil, errNotFound
	}
	return down.get(key[1:], maxFormat)
}

func (leaf memBSLeaf) get(key []string, maxFormat int) (Assertion, error) {
	key0 := key[0]
	cur := leaf.cur(key0, maxFormat)
	if cur == nil {
		return nil, errNotFound
	}
	return cur, nil
}

func (br memBSBranch) search(hint []string, found func(Assertion), maxFormat int) {
	hint0 := hint[0]
	if hint0 == "" {
		for _, down := range br {
			down.search(hint[1:], found, maxFormat)
		}
		return
	}
	down := br[hint0]
	if down != nil {
		down.search(hint[1:], found, maxFormat)
	}
	return
}

func (leaf memBSLeaf) search(hint []string, found func(Assertion), maxFormat int) {
	hint0 := hint[0]
	if hint0 == "" {
		for key := range leaf {
			cand := leaf.cur(key, maxFormat)
			if cand != nil {
				found(cand)
			}
		}
		return
	}

	cur := leaf.cur(hint0, maxFormat)
	if cur != nil {
		found(cur)
	}
}

// NewMemoryBackstore creates a memory backed assertions backstore.
func NewMemoryBackstore() Backstore {
	return &memoryBackstore{
		top: make(memBSBranch),
	}
}

func (mbs *memoryBackstore) Put(assertType *AssertionType, assert Assertion) error {
	mbs.mu.Lock()
	defer mbs.mu.Unlock()

	internalKey := make([]string, 1, 1+len(assertType.PrimaryKey))
	internalKey[0] = assertType.Name
	internalKey = append(internalKey, assert.Ref().PrimaryKey...)

	err := mbs.top.put(assertType, internalKey, assert)
	return err
}

func (mbs *memoryBackstore) Get(assertType *AssertionType, key []string, maxFormat int) (Assertion, error) {
	mbs.mu.RLock()
	defer mbs.mu.RUnlock()

	internalKey := make([]string, 1+len(assertType.PrimaryKey))
	internalKey[0] = assertType.Name
	copy(internalKey[1:], key)

	a, err := mbs.top.get(internalKey, maxFormat)
	if err == errNotFound {
		return nil, &NotFoundError{Type: assertType}
	}
	return a, err
}

func (mbs *memoryBackstore) Search(assertType *AssertionType, headers map[string]string, foundCb func(Assertion), maxFormat int) error {
	mbs.mu.RLock()
	defer mbs.mu.RUnlock()

	hint := make([]string, 1+len(assertType.PrimaryKey))
	hint[0] = assertType.Name
	for i, name := range assertType.PrimaryKey {
		hint[1+i] = headers[name]
	}

	candCb := func(a Assertion) {
		if searchMatch(a, headers) {
			foundCb(a)
		}
	}

	mbs.top.search(hint, candCb, maxFormat)
	return nil
}
