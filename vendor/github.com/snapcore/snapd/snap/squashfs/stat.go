// -*- Mode: Go; indent-tabs-mode: t -*-

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

package squashfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type SnapFileOwner struct {
	UID uint32
	GID uint32
}

type stat struct {
	path  string
	size  int64
	mode  os.FileMode
	mtime time.Time
	user  string
	group string
}

func (s stat) Name() string       { return filepath.Base(s.path) }
func (s stat) Size() int64        { return s.size }
func (s stat) Mode() os.FileMode  { return s.mode }
func (s stat) ModTime() time.Time { return s.mtime }
func (s stat) IsDir() bool        { return s.mode.IsDir() }
func (s stat) Sys() interface{}   { return nil }
func (s stat) Path() string       { return s.path } // not path of os.FileInfo

const minLen = len("drwxrwxr-x u/g             53595 2017-12-08 11:19 .")

func fromRaw(raw []byte) (*stat, error) {
	if len(raw) < minLen {
		return nil, errBadLine(raw)
	}

	st := &stat{}

	parsers := []func([]byte) (int, error){
		// first, the file mode, e.g. "-rwxr-xr-x"
		st.parseMode,
		// next, user/group info
		st.parseOwner,
		// next'll come the size or the node type
		st.parseSize,
		// and then the time
		st.parseTimeUTC,
		// and finally the path
		st.parsePath,
	}
	p := 0
	for _, parser := range parsers {
		n, err := parser(raw[p:])
		if err != nil {
			return nil, err
		}
		p += n
		if p < len(raw) && raw[p] != ' ' {
			return nil, errBadLine(raw)
		}
		p++
	}

	if st.mode&os.ModeSymlink != 0 {
		// the symlink *could* be from a file called "foo -> bar" to
		// another called "baz -> quux" in which case the following
		// would be wrong, but so be it.

		idx := strings.Index(st.path, " -> ")
		if idx < 0 {
			return nil, errBadPath(raw)
		}
		st.path = st.path[:idx]
	}

	return st, nil
}

type statError struct {
	part string
	raw  []byte
}

func (e statError) Error() string {
	return fmt.Sprintf("cannot parse %s: %q", e.part, e.raw)
}

func errBadLine(raw []byte) statError {
	return statError{
		part: "line",
		raw:  raw,
	}
}

func errBadMode(raw []byte) statError {
	return statError{
		part: "mode",
		raw:  raw,
	}
}

func errBadOwner(raw []byte) statError {
	return statError{
		part: "owner",
		raw:  raw,
	}
}

func errBadNode(raw []byte) statError {
	return statError{
		part: "node",
		raw:  raw,
	}
}

func errBadSize(raw []byte) statError {
	return statError{
		part: "size",
		raw:  raw,
	}
}

func errBadTime(raw []byte) statError {
	return statError{
		part: "time",
		raw:  raw,
	}
}

func errBadPath(raw []byte) statError {
	return statError{
		part: "path",
		raw:  raw,
	}
}

func (st *stat) parseTimeUTC(raw []byte) (int, error) {
	const timelen = 16
	t, err := time.Parse("2006-01-02 15:04", string(raw[:timelen]))
	if err != nil {
		return 0, errBadTime(raw)
	}

	st.mtime = t

	return timelen, nil
}

func (st *stat) parseMode(raw []byte) (int, error) {
	switch raw[0] {
	case '-':
		// 0
	case 'd':
		st.mode |= os.ModeDir
	case 's':
		st.mode |= os.ModeSocket
	case 'c':
		st.mode |= os.ModeCharDevice
	case 'b':
		st.mode |= os.ModeDevice
	case 'p':
		st.mode |= os.ModeNamedPipe
	case 'l':
		st.mode |= os.ModeSymlink
	default:
		return 0, errBadMode(raw)
	}

	for i := 0; i < 3; i++ {
		m, err := modeFromTriplet(raw[1+3*i:4+3*i], uint(2-i))
		if err != nil {
			return 0, err
		}
		st.mode |= m
	}

	// always this length (1+3*3==10)
	return 10, nil
}

func (st *stat) parseOwner(raw []byte) (int, error) {
	var p, ui, uj, gi, gj int

	// first check it's sane (at least two non-space chars)
	if raw[0] == ' ' || raw[1] == ' ' {
		return 0, errBadLine(raw)
	}

	ui = 0
	// from useradd(8): Usernames may only be up to 32 characters long.
	// from groupadd(8): Groupnames may only be up to 32 characters long.
	// +1 for the separator, +1 for the ending space
	maxL := 66
	if len(raw) < maxL {
		maxL = len(raw)
	}
out:
	for p = ui; p < maxL; p++ {
		switch raw[p] {
		case '/':
			uj = p
			gi = p + 1
		case ' ':
			gj = p
			break out
		}
	}

	if uj == 0 || gj == 0 || gi == gj {
		return 0, errBadOwner(raw)
	}
	st.user, st.group = string(raw[ui:uj]), string(raw[gi:gj])

	return p, nil
}

func modeFromTriplet(trip []byte, shift uint) (os.FileMode, error) {
	var mode os.FileMode
	high := false
	if len(trip) != 3 {
		panic("bad triplet length")
	}
	switch trip[0] {
	case '-':
		// 0
	case 'r':
		mode |= 4
	default:
		return 0, errBadMode(trip)
	}
	switch trip[1] {
	case '-':
		// 0
	case 'w':
		mode |= 2
	default:
		return 0, errBadMode(trip)
	}
	switch trip[2] {
	case '-':
		// 0
	case 'x':
		mode |= 1
	case 'S', 'T':
		high = true
	case 's', 't':
		mode |= 1
		high = true
	default:
		return 0, errBadMode(trip)
	}

	mode <<= 3 * shift
	if high {
		mode |= (01000 << shift)
	}
	return mode, nil
}

func (st *stat) parseSize(raw []byte) (int, error) {
	// the "size" column, for regular files, is the file size in bytes:
	//   -rwxr-xr-x user/group            53595 2017-12-08 11:19 ./yadda
	//                                    ^^^^^ like this
	// for devices, though, it's the major, minor of the node:
	//   crw-rw---- root/audio           14,  3 2017-12-05 10:29 ./dev/dsp
	//                                   ^^^^^^ like so
	// (for other things it is a size, although what the size is
	// _of_ is left as an exercise for the reader)
	isNode := st.mode&(os.ModeDevice|os.ModeCharDevice) != 0
	p := 0
	maxP := len(raw) - len("2006-01-02 15:04 .")

	for raw[p] == ' ' {
		if p >= maxP {
			return 0, errBadLine(raw)
		}
		p++
	}

	ni := p

	for raw[p] >= '0' && raw[p] <= '9' {
		if p >= maxP {
			return 0, errBadLine(raw)
		}
		p++
	}

	if p == ni {
		if isNode {
			return 0, errBadNode(raw)
		}
		return 0, errBadSize(raw)
	}

	if isNode {
		if raw[p] != ',' {
			return 0, errBadNode(raw)
		}

		p++

		// drop the space before the minor mode
		for raw[p] == ' ' {
			p++
		}
		// drop the minor mode
		for raw[p] >= '0' && raw[p] <= '9' {
			p++
		}

		if raw[p] != ' ' {
			return 0, errBadNode(raw)
		}
	} else {
		if raw[p] != ' ' {
			return 0, errBadSize(raw)
		}
		// note that, much as it makes very little sense, the arch-
		// dependent st_size is never an unsigned 64 bit quantity.
		// It's one of unsigned long, long long, or just off_t.
		//
		// Also note os.FileInfo's Size needs to return an int64, and
		// squashfs's inode->data (where it stores sizes for regular
		// files) is a long long.
		sz, err := strconv.ParseInt(string(raw[ni:p]), 10, 64)
		if err != nil {
			return 0, errBadSize(raw)
		}
		st.size = sz
	}

	return p, nil
}

func (st *stat) parsePath(raw []byte) (int, error) {
	if raw[0] != '.' {
		return 0, errBadPath(raw)
	}
	if len(raw[1:]) == 0 {
		st.path = "/"
	} else {
		st.path = string(raw[1:])
	}

	return len(raw), nil
}
