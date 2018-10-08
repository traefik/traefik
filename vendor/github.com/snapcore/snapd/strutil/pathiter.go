// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package strutil

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PathIterator traverses through parts (directories and files) of some
// path. The filesystem is never consulted, traversal is done purely in memory.
//
// The iterator is useful in implementing secure traversal of absolute paths
// using the common idiom of opening the root directory followed by a chain of
// openat calls.
//
// A simple example on how to use the iterator:
// ```
// iter:= NewPathIterator(path)
// for iter.Next() {
//    // Use iter.CurrentName() with openat(2) family of functions.
//    // Use iter.CurrentPath() or iter.CurrentBase() for context.
// }
// ```
type PathIterator struct {
	path        string
	left, right int
	depth       int
}

// NewPathIterator returns an iterator for traversing the given path.
// The path is passed through filepath.Clean automatically.
func NewPathIterator(path string) (*PathIterator, error) {
	cleanPath := filepath.Clean(path)
	if cleanPath != path && cleanPath+"/" != path {
		return nil, fmt.Errorf("cannot iterate over unclean path %q", path)
	}
	return &PathIterator{path: path}, nil
}

// Path returns the path being traversed.
func (iter *PathIterator) Path() string {
	return iter.path
}

// CurrentName returns the name of the current path element.
// The return value may end with '/'. Use CleanName to avoid that.
func (iter *PathIterator) CurrentName() string {
	return iter.path[iter.left:iter.right]
}

// CurrentCleanName returns the same value as Name with right slash trimmed.
func (iter *PathIterator) CurrentCleanName() string {
	if iter.right > 0 && iter.path[iter.right-1:iter.right] == "/" {
		return iter.path[iter.left : iter.right-1]
	}
	return iter.path[iter.left:iter.right]
}

// CurrentPath returns the prefix of path that was traversed, including the current name.
func (iter *PathIterator) CurrentPath() string {
	return iter.path[:iter.right]
}

// CurrentBase returns the prefix of the path that was traversed, excluding the current name.
func (iter *PathIterator) CurrentBase() string {
	return iter.path[:iter.left]
}

// Depth returns the directory depth of the current path.
//
// This is equal to the number of traversed directories, including that of the
// root directory.
func (iter *PathIterator) Depth() int {
	return iter.depth
}

// Next advances the iterator to the next name, returning true if one is found.
//
// If this method returns false then no change is made and all helper methods
// retain their previous return values.
func (iter *PathIterator) Next() bool {
	// Initial state
	// P: "foo/bar"
	// L:  ^
	// R:  ^
	//
	// Next is called
	// P: "foo/bar"
	// L:  ^  |
	// R:     ^
	//
	// Next is called
	// P: "foo/bar"
	// L:     ^   |
	// R:         ^

	// Next is called but returns false
	// P: "foo/bar"
	// L:     ^   |
	// R:         ^
	if iter.right >= len(iter.path) {
		return false
	}
	iter.left = iter.right
	if idx := strings.IndexRune(iter.path[iter.right:], '/'); idx != -1 {
		iter.right += idx + 1
	} else {
		iter.right = len(iter.path)
	}
	iter.depth++
	return true
}

// Rewind returns the iterator the the initial state, allowing the path to be traversed again.
func (iter *PathIterator) Rewind() {
	iter.left = 0
	iter.right = 0
	iter.depth = 0
}
