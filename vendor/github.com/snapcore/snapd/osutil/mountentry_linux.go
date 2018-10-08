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
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// MountEntry describes an /etc/fstab-like mount entry.
//
// Fields are named after names in struct returned by getmntent(3).
//
// struct mntent {
//     char *mnt_fsname;   /* name of mounted filesystem */
//     char *mnt_dir;      /* filesystem path prefix */
//     char *mnt_type;     /* mount type (see Mntent.h) */
//     char *mnt_opts;     /* mount options (see Mntent.h) */
//     int   mnt_freq;     /* dump frequency in days */
//     int   mnt_passno;   /* pass number on parallel fsck */
// };
type MountEntry struct {
	Name    string
	Dir     string
	Type    string
	Options []string

	DumpFrequency   int
	CheckPassNumber int
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Equal checks if one entry is equal to another
func (e *MountEntry) Equal(o *MountEntry) bool {
	return (e.Name == o.Name && e.Dir == o.Dir && e.Type == o.Type &&
		equalStrings(e.Options, o.Options) && e.DumpFrequency == o.DumpFrequency &&
		e.CheckPassNumber == o.CheckPassNumber)
}

// escape replaces whitespace characters so that getmntent can parse it correctly.
var escape = strings.NewReplacer(
	" ", `\040`,
	"\t", `\011`,
	"\n", `\012`,
	"\\", `\134`,
).Replace

// unescape replaces escape sequences used by setmnt with whitespace characters.
var unescape = strings.NewReplacer(
	`\040`, " ",
	`\011`, "\t",
	`\012`, "\n",
	`\134`, "\\",
).Replace

// Escape returns the given path with space, tab, newline and forward slash escaped.
func Escape(path string) string {
	return escape(path)
}

// Unescape returns the given path with space, tab, newline and forward slash unescaped.
func Unescape(path string) string {
	return unescape(path)
}

func (e MountEntry) String() string {
	// Name represents name of the device in a mount entry.
	name := "none"
	if e.Name != "" {
		name = escape(e.Name)
	}
	// Dir represents mount directory in a mount entry.
	dir := "none"
	if e.Dir != "" {
		dir = escape(e.Dir)
	}
	// Type represents file system type in a mount entry.
	fsType := "none"
	if e.Type != "" {
		fsType = escape(e.Type)
	}
	// Options represents mount options in a mount entry.
	options := "defaults"
	if len(e.Options) != 0 {
		options = escape(strings.Join(e.Options, ","))
	}
	return fmt.Sprintf("%s %s %s %s %d %d",
		name, dir, fsType, options, e.DumpFrequency, e.CheckPassNumber)
}

// ParseMountEntry parses a fstab-like entry.
func ParseMountEntry(s string) (MountEntry, error) {
	var e MountEntry
	var err error
	var df, cpn int
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' })
	// Look for any inline comments. The first field that starts with '#' is a comment.
	for i, field := range fields {
		if strings.HasPrefix(field, "#") {
			fields = fields[:i]
			break
		}
	}
	// Do all error checks before any assignments to `e'
	if len(fields) < 3 || len(fields) > 6 {
		return e, fmt.Errorf("expected between 3 and 6 fields, found %d", len(fields))
	}
	e.Name = unescape(fields[0])
	e.Dir = unescape(fields[1])
	e.Type = unescape(fields[2])
	// Parse Options if we have at least 4 fields
	if len(fields) > 3 {
		e.Options = strings.Split(unescape(fields[3]), ",")
	}
	// Parse DumpFrequency if we have at least 5 fields
	if len(fields) > 4 {
		df, err = strconv.Atoi(fields[4])
		if err != nil {
			return e, fmt.Errorf("cannot parse dump frequency: %q", fields[4])
		}
	}
	e.DumpFrequency = df
	// Parse CheckPassNumber if we have at least 6 fields
	if len(fields) > 5 {
		cpn, err = strconv.Atoi(fields[5])
		if err != nil {
			return e, fmt.Errorf("cannot parse check pass number: %q", fields[5])
		}
	}
	e.CheckPassNumber = cpn
	return e, nil
}

// MountOptsToCommonFlags converts mount options strings to a mount flag,
// returning unparsed flags. The unparsed flags will not contain any snapd-
// specific mount option, those starting with the string "x-snapd."
func MountOptsToCommonFlags(opts []string) (flags int, unparsed []string) {
	for _, opt := range opts {
		switch opt {
		case "ro":
			flags |= syscall.MS_RDONLY
		case "nosuid":
			flags |= syscall.MS_NOSUID
		case "nodev":
			flags |= syscall.MS_NODEV
		case "noexec":
			flags |= syscall.MS_NOEXEC
		case "sync":
			flags |= syscall.MS_SYNCHRONOUS
		case "remount":
			flags |= syscall.MS_REMOUNT
		case "mand":
			flags |= syscall.MS_MANDLOCK
		case "dirsync":
			flags |= syscall.MS_DIRSYNC
		case "noatime":
			flags |= syscall.MS_NOATIME
		case "nodiratime":
			flags |= syscall.MS_NODIRATIME
		case "bind":
			flags |= syscall.MS_BIND
		case "rbind":
			flags |= syscall.MS_BIND | syscall.MS_REC
		case "move":
			flags |= syscall.MS_MOVE
		case "silent":
			flags |= syscall.MS_SILENT
		case "acl":
			flags |= syscall.MS_POSIXACL
		case "private":
			flags |= syscall.MS_PRIVATE
		case "rprivate":
			flags |= syscall.MS_PRIVATE | syscall.MS_REC
		case "slave":
			flags |= syscall.MS_SLAVE
		case "rslave":
			flags |= syscall.MS_SLAVE | syscall.MS_REC
		case "shared":
			flags |= syscall.MS_SHARED
		case "rshared":
			flags |= syscall.MS_SHARED | syscall.MS_REC
		case "relatime":
			flags |= syscall.MS_RELATIME
		case "strictatime":
			flags |= syscall.MS_STRICTATIME
		default:
			if !strings.HasPrefix(opt, "x-snapd.") {
				unparsed = append(unparsed, opt)
			}
		}
	}
	return flags, unparsed
}

// MountOptsToFlags converts mount options strings to a mount flag.
func MountOptsToFlags(opts []string) (flags int, err error) {
	flags, unparsed := MountOptsToCommonFlags(opts)
	for _, opt := range unparsed {
		if !strings.HasPrefix(opt, "x-snapd.") {
			return 0, fmt.Errorf("unsupported mount option: %q", opt)
		}
	}
	return flags, nil
}

// OptStr returns the value part of a key=value mount option.
// The name of the option must not contain the trailing "=" character.
func (e *MountEntry) OptStr(name string) (string, bool) {
	prefix := name + "="
	for _, opt := range e.Options {
		if strings.HasPrefix(opt, prefix) {
			kv := strings.SplitN(opt, "=", 2)
			return kv[1], true
		}
	}
	return "", false
}

// OptBool returns true if a given mount option is present.
func (e *MountEntry) OptBool(name string) bool {
	for _, opt := range e.Options {
		if opt == name {
			return true
		}
	}
	return false
}

var (
	validModeRe      = regexp.MustCompile("^0[0-7]{3}$")
	validUserGroupRe = regexp.MustCompile("(^[0-9]+$)")
)

// XSnapdMode returns the file mode associated with x-snapd.mode mount option.
// If the mode is not specified explicitly then a default mode of 0755 is assumed.
func (e *MountEntry) XSnapdMode() (os.FileMode, error) {
	if opt, ok := e.OptStr("x-snapd.mode"); ok {
		if !validModeRe.MatchString(opt) {
			return 0, fmt.Errorf("cannot parse octal file mode from %q", opt)
		}
		var mode os.FileMode
		n, err := fmt.Sscanf(opt, "%o", &mode)
		if err != nil || n != 1 {
			return 0, fmt.Errorf("cannot parse octal file mode from %q", opt)
		}
		return mode, nil
	}
	return 0755, nil
}

// XSnapdUID returns the user associated with x-snapd-user mount option.  If
// the mode is not specified explicitly then a default "root" use is
// returned.
func (e *MountEntry) XSnapdUID() (uid uint64, err error) {
	if opt, ok := e.OptStr("x-snapd.uid"); ok {
		if !validUserGroupRe.MatchString(opt) {
			return math.MaxUint64, fmt.Errorf("cannot parse user name %q", opt)
		}
		// Try to parse a numeric ID first.
		if n, err := fmt.Sscanf(opt, "%d", &uid); n == 1 && err == nil {
			return uid, nil
		}
		return uid, nil
	}
	return 0, nil
}

// XSnapdGID returns the user associated with x-snapd-user mount option.  If
// the mode is not specified explicitly then a default "root" use is
// returned.
func (e *MountEntry) XSnapdGID() (gid uint64, err error) {
	if opt, ok := e.OptStr("x-snapd.gid"); ok {
		if !validUserGroupRe.MatchString(opt) {
			return math.MaxUint64, fmt.Errorf("cannot parse group name %q", opt)
		}
		// Try to parse a numeric ID first.
		if n, err := fmt.Sscanf(opt, "%d", &gid); n == 1 && err == nil {
			return gid, nil
		}
		return gid, nil
	}
	return 0, nil
}

// XSnapdEntryID returns the identifier of a given mount enrty.
//
// Identifiers are kept in the x-snapd.id mount option. The value is a string
// that identifies a mount entry and is stable across invocations of snapd. In
// absence of that identifier the entry mount point is returned.
func (e *MountEntry) XSnapdEntryID() string {
	if val, ok := e.OptStr("x-snapd.id"); ok {
		return val
	}
	return e.Dir
}

// XSnapdNeededBy the identifier of an entry which needs this entry to function.
//
// The "needed by" identifiers are kept in the x-snapd.needed-by mount option.
// The value is a string that identifies another mount entry which, in order to
// be feasible, has spawned one or more additional support entries. Each such
// entry contains the needed-by attribute.
func (e *MountEntry) XSnapdNeededBy() string {
	val, _ := e.OptStr("x-snapd.needed-by")
	return val
}

// XSnapdOrigin returns the origin of a given mount entry.
//
// Currently only "layout" entries are identified with a unique origin string.
func (e *MountEntry) XSnapdOrigin() string {
	val, _ := e.OptStr("x-snapd.origin")
	return val
}

// XSnapdSynthetic returns true of a given mount entry is synthetic.
//
// Synthetic mount entries are created by snap-update-ns itself, separately
// from what snapd instructed. Such entries are needed to make other things
// possible.  They are identified by having the "x-snapd.synthetic" mount
// option.
func (e *MountEntry) XSnapdSynthetic() bool {
	return e.OptBool("x-snapd.synthetic")
}

// XSnapdKind returns the kind of a given mount entry.
//
// There are three kinds of mount entries today: one for directories, one for
// files and one for symlinks. The values are "", "file" and "symlink" respectively.
//
// Directories use the empty string (in fact they don't need the option at
// all) as this was the default and is retained for backwards compatibility.
func (e *MountEntry) XSnapdKind() string {
	val, _ := e.OptStr("x-snapd.kind")
	return val
}

// XSnapdDetach returns true if a mount entry should be detached rather than unmounted.
//
// Whenever we create a recursive bind mount we don't want to just unmount it
// as it may have replicated additional mount entries. For simplicity and
// race-free behavior we just detach such mount entries and let the kernel do
// the rest.
func (e *MountEntry) XSnapdDetach() bool {
	return e.OptBool("x-snapd.detach")
}

// XSnapdSymlink returns the target for a symlink mount entry.
//
// For non-symlinks an empty string is returned.
func (e *MountEntry) XSnapdSymlink() string {
	val, _ := e.OptStr("x-snapd.symlink")
	return val
}

// XSnapdIgnoreMissing returns true if a mount entry should be ignored
// if the source or target are missing.
//
// By default, snap-update-ns will try to create missing source and
// target paths when processing a mount entry.  In some cases, this
// behaviour is not desired and it would be better to ignore the mount
// entry when the source or target are missing.
func (e *MountEntry) XSnapdIgnoreMissing() bool {
	return e.OptBool("x-snapd.ignore-missing")
}

// XSnapdNeededBy returns the string "x-snapd.needed-by=..." with the given path appended.
func XSnapdNeededBy(path string) string {
	return fmt.Sprintf("x-snapd.needed-by=%s", path)
}

// XSnapdSynthetic returns the string "x-snapd.synthetic".
func XSnapdSynthetic() string {
	return "x-snapd.synthetic"
}

// XSnapdDetach returns the string "x-snapd.detach".
func XSnapdDetach() string {
	return "x-snapd.detach"
}

// XSnapdKindSymlink returns the string "x-snapd.kind=symlink".
func XSnapdKindSymlink() string {
	return "x-snapd.kind=symlink"
}

// XSnapdKindFile returns the string "x-snapd.kind=file".
func XSnapdKindFile() string {
	return "x-snapd.kind=file"
}

// XSnapdOriginLayout returns the string "x-snapd.origin=layout"
func XSnapdOriginLayout() string {
	return "x-snapd.origin=layout"
}

// XSnapdOriginOvername returns the string "x-snapd.origin=overname"
func XSnapdOriginOvername() string {
	return "x-snapd.origin=overname"
}

// XSnapdUser returns the string "x-snapd.user=%d".
func XSnapdUser(uid uint32) string {
	return fmt.Sprintf("x-snapd.user=%d", uid)
}

// XSnapdGroup returns the string "x-snapd.group=%d".
func XSnapdGroup(gid uint32) string {
	return fmt.Sprintf("x-snapd.group=%d", gid)
}

// XSnapdMode returns the string "x-snapd.mode=%#o".
func XSnapdMode(mode uint32) string {
	return fmt.Sprintf("x-snapd.mode=%#o", mode)
}

// XSnapdSymlink returns the string "x-snapd.symlink=%s".
func XSnapdSymlink(oldname string) string {
	return fmt.Sprintf("x-snapd.symlink=%s", oldname)
}

// XSnapdIgnoreMissing returns the string "x-snapd.ignore-missing".
func XSnapdIgnoreMissing() string {
	return "x-snapd.ignore-missing"
}
