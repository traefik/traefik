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

package osutil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// MountInfoEntry contains data from /proc/$PID/mountinfo
//
// For details please refer to mountinfo documentation at
// https://www.kernel.org/doc/Documentation/filesystems/proc.txt
type MountInfoEntry struct {
	MountID        int
	ParentID       int
	DevMajor       int
	DevMinor       int
	Root           string
	MountDir       string
	MountOptions   map[string]string
	OptionalFields []string
	FsType         string
	MountSource    string
	SuperOptions   map[string]string
}

func flattenMap(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for i, key := range keys {
		if i > 0 {
			buf.WriteRune(',')
		}
		if m[key] != "" {
			fmt.Fprintf(&buf, "%s=%s", escape(key), escape(m[key]))
		} else {
			buf.WriteString(escape(key))
		}
	}
	return buf.String()
}

func flattenList(l []string) string {
	var buf bytes.Buffer
	for i, item := range l {
		if i > 0 {
			buf.WriteRune(',')
		}
		buf.WriteString(escape(item))
	}
	return buf.String()
}

func (mi *MountInfoEntry) String() string {
	maybeSpace := " "
	if len(mi.OptionalFields) == 0 {
		maybeSpace = ""
	}
	return fmt.Sprintf("%d %d %d:%d %s %s %s %s%s- %s %s %s",
		mi.MountID, mi.ParentID, mi.DevMajor, mi.DevMinor, escape(mi.Root),
		escape(mi.MountDir), flattenMap(mi.MountOptions), flattenList(mi.OptionalFields),
		maybeSpace, escape(mi.FsType), escape(mi.MountSource),
		flattenMap(mi.SuperOptions))
}

// LoadMountInfo loads list of mounted entries from a given file.
//
// The file is typically ProcSelfMountInfo but any other process mount table
// can be read the same way.
func LoadMountInfo(fname string) ([]*MountInfoEntry, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadMountInfo(f)
}

// ReadMountInfo reads and parses a mountinfo file.
func ReadMountInfo(reader io.Reader) ([]*MountInfoEntry, error) {
	scanner := bufio.NewScanner(reader)
	var entries []*MountInfoEntry
	for scanner.Scan() {
		s := scanner.Text()
		entry, err := ParseMountInfoEntry(s)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// ParseMountInfoEntry parses a single line of /proc/$PID/mountinfo file.
func ParseMountInfoEntry(s string) (*MountInfoEntry, error) {
	var e MountInfoEntry
	var err error
	fields := strings.Fields(s)
	// The format is variable-length, but at least 10 fields are mandatory.
	// The (7) below is a list of optional field which is terminated with (8).
	// 36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
	// (1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)
	if len(fields) < 10 {
		return nil, fmt.Errorf("incorrect number of fields, expected at least 10 but found %d", len(fields))
	}
	// Parse MountID (decimal number).
	e.MountID, err = strconv.Atoi(fields[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse mount ID: %q", fields[0])
	}
	// Parse ParentID (decimal number).
	e.ParentID, err = strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("cannot parse parent mount ID: %q", fields[1])
	}
	// Parses DevMajor:DevMinor pair (decimal numbers separated by colon).
	subFields := strings.FieldsFunc(fields[2], func(r rune) bool { return r == ':' })
	if len(subFields) != 2 {
		return nil, fmt.Errorf("cannot parse device major:minor number pair: %q", fields[2])
	}
	e.DevMajor, err = strconv.Atoi(subFields[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse device major number: %q", subFields[0])
	}
	e.DevMinor, err = strconv.Atoi(subFields[1])
	if err != nil {
		return nil, fmt.Errorf("cannot parse device minor number: %q", subFields[1])
	}
	// NOTE: All string fields use the same escape/unescape logic as fstab files.
	// Parse Root, MountDir and MountOptions fields.
	e.Root = unescape(fields[3])
	e.MountDir = unescape(fields[4])
	e.MountOptions = parseMountOpts(unescape(fields[5]))
	// Optional fields are terminated with a "-" value and start
	// after the mount options field. Skip ahead until we see the "-"
	// marker.
	var i int
	for i = 6; i < len(fields) && fields[i] != "-"; i++ {
	}
	if i == len(fields) {
		return nil, fmt.Errorf("list of optional fields is not terminated properly")
	}
	e.OptionalFields = fields[6:i]
	for j := range e.OptionalFields {
		e.OptionalFields[j] = unescape(e.OptionalFields[j])
	}
	// Parse the last three fixed fields.
	tailFields := fields[i+1:]
	if len(tailFields) != 3 {
		return nil, fmt.Errorf("incorrect number of tail fields, expected 3 but found %d", len(tailFields))
	}
	e.FsType = unescape(tailFields[0])
	e.MountSource = unescape(tailFields[1])
	e.SuperOptions = parseMountOpts(unescape(tailFields[2]))
	return &e, nil
}

func parseMountOpts(opts string) map[string]string {
	result := make(map[string]string)
	for _, opt := range strings.Split(opts, ",") {
		keyValue := strings.SplitN(opt, "=", 2)
		key := keyValue[0]
		if len(keyValue) == 2 {
			value := keyValue[1]
			result[key] = value
		} else {
			result[key] = ""
		}
	}
	return result
}
