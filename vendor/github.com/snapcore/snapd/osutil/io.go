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
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/snapcore/snapd/osutil/sys"
	"github.com/snapcore/snapd/strutil"
)

// AtomicWriteFlags are a bitfield of flags for AtomicWriteFile
type AtomicWriteFlags uint

const (
	// AtomicWriteFollow makes AtomicWriteFile follow symlinks
	AtomicWriteFollow AtomicWriteFlags = 1 << iota
)

// Allow disabling sync for testing. This brings massive improvements on
// certain filesystems (like btrfs) and very much noticeable improvements in
// all unit tests in genreal.
var snapdUnsafeIO bool = len(os.Args) > 0 && strings.HasSuffix(os.Args[0], ".test") && GetenvBool("SNAPD_UNSAFE_IO", true)

// An AtomicFile is similar to an os.File but it has an additional
// Commit() method that does whatever needs to be done so the
// modification is "atomic": an AtomicFile will do its best to leave
// either the previous content or the new content in permanent
// storage. It also has a Cancel() method to abort and clean up.
type AtomicFile struct {
	*os.File

	target  string
	tmpname string
	uid     sys.UserID
	gid     sys.GroupID
	closed  bool
	renamed bool
}

// NewAtomicFile builds an AtomicFile backed by an *os.File that will have
// the given filename, permissions and uid/gid when Committed.
//
//   It _might_ be implemented using O_TMPFILE (see open(2)).
//
// Note that it won't follow symlinks and will replace existing symlinks with
// the real file, unless the AtomicWriteFollow flag is specified.
//
// It is the caller's responsibility to clean up on error, by calling Cancel().
//
// It is also the caller's responsibility to coordinate access to this, if it
// is used from different goroutines.
//
// Also note that there are a number of scenarios where Commit fails and then
// Cancel also fails. In all these scenarios your filesystem was probably in a
// rather poor state. Good luck.
func NewAtomicFile(filename string, perm os.FileMode, flags AtomicWriteFlags, uid sys.UserID, gid sys.GroupID) (aw *AtomicFile, err error) {
	if flags&AtomicWriteFollow != 0 {
		if fn, err := os.Readlink(filename); err == nil || (fn != "" && os.IsNotExist(err)) {
			if filepath.IsAbs(fn) {
				filename = fn
			} else {
				filename = filepath.Join(filepath.Dir(filename), fn)
			}
		}
	}
	// The tilde is appended so that programs that inspect all files in some
	// directory are more likely to ignore this file as an editor backup file.
	//
	// This fixes an issue in apparmor-utils package, specifically in
	// aa-enforce. Tools from this package enumerate all profiles by loading
	// parsing any file found in /etc/apparmor.d/, skipping only very specific
	// suffixes, such as the one we selected below.
	tmp := filename + "." + strutil.MakeRandomString(12) + "~"

	fd, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_EXCL, perm)
	if err != nil {
		return nil, err
	}

	return &AtomicFile{
		File:    fd,
		target:  filename,
		tmpname: tmp,
		uid:     uid,
		gid:     gid,
	}, nil
}

// ErrCannotCancel means the Commit operation failed at the last step, and
// your luck has run out.
var ErrCannotCancel = errors.New("cannot cancel: file has already been renamed")

func (aw *AtomicFile) Close() error {
	aw.closed = true
	return aw.File.Close()
}

// Cancel closes the AtomicWriter, and cleans up any artifacts. Cancel
// can fail if Commit() was (even partially) successful, but calling
// Cancel after a successful Commit does nothing beyond returning
// error--so it's always safe to defer a Cancel().
func (aw *AtomicFile) Cancel() error {
	if aw.renamed {
		return ErrCannotCancel
	}

	var e1, e2 error
	if aw.tmpname != "" {
		e1 = os.Remove(aw.tmpname)
	}
	if !aw.closed {
		e2 = aw.Close()
	}
	if e1 != nil {
		return e1
	}
	return e2
}

var chown = sys.Chown

const NoChown = sys.FlagID

// Commit the modification; make it permanent.
//
// If Commit succeeds, the writer is closed and further attempts to
// write will fail. If Commit fails, the writer _might_ be closed;
// Cancel() needs to be called to clean up.
func (aw *AtomicFile) Commit() error {
	if aw.uid != NoChown || aw.gid != NoChown {
		if err := chown(aw.File, aw.uid, aw.gid); err != nil {
			return err
		}
	}

	var dir *os.File
	if !snapdUnsafeIO {
		// XXX: if go switches to use aio_fsync, we need to open the dir for writing
		d, err := os.Open(filepath.Dir(aw.target))
		if err != nil {
			return err
		}
		dir = d
		defer dir.Close()

		if err := aw.Sync(); err != nil {
			return err
		}
	}

	if err := aw.Close(); err != nil {
		return err
	}

	if err := os.Rename(aw.tmpname, aw.target); err != nil {
		return err
	}
	aw.renamed = true // it is now too late to Cancel()

	if !snapdUnsafeIO {
		return dir.Sync()
	}

	return nil
}

// The AtomicWrite* family of functions work like ioutil.WriteFile(), but the
// file created is an AtomicWriter, which is Committed before returning.
//
// AtomicWriteChown and AtomicWriteFileChown take an uid and a gid that can be
// used to specify the ownership of the created file. A special value of
// 0xffffffff (math.MaxUint32, or NoChown for convenience) can be used to
// request no change to that attribute.
//
// AtomicWriteFile and AtomicWriteFileChown take the content to be written as a
// []byte, and so work exactly like io.WriteFile(); AtomicWrite and
// AtomicWriteChown take an io.Reader which is copied into the file instead,
// and so are more amenable to streaming.
func AtomicWrite(filename string, reader io.Reader, perm os.FileMode, flags AtomicWriteFlags) (err error) {
	return AtomicWriteChown(filename, reader, perm, flags, NoChown, NoChown)
}

func AtomicWriteFile(filename string, data []byte, perm os.FileMode, flags AtomicWriteFlags) (err error) {
	return AtomicWriteChown(filename, bytes.NewReader(data), perm, flags, NoChown, NoChown)
}

func AtomicWriteFileChown(filename string, data []byte, perm os.FileMode, flags AtomicWriteFlags, uid sys.UserID, gid sys.GroupID) (err error) {
	return AtomicWriteChown(filename, bytes.NewReader(data), perm, flags, uid, gid)
}

func AtomicWriteChown(filename string, reader io.Reader, perm os.FileMode, flags AtomicWriteFlags, uid sys.UserID, gid sys.GroupID) (err error) {
	aw, err := NewAtomicFile(filename, perm, flags, uid, gid)
	if err != nil {
		return err
	}

	// Cancel once Committed is a NOP :-)
	defer aw.Cancel()

	if _, err := io.Copy(aw, reader); err != nil {
		return err
	}

	return aw.Commit()
}
