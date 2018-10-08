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

package snap

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/snap/snapdir"
	"github.com/snapcore/snapd/snap/squashfs"
)

// Container is the interface to interact with the low-level snap files.
type Container interface {
	// Size returns the size of the snap in bytes.
	Size() (int64, error)

	// ReadFile returns the content of a single file from the snap.
	ReadFile(relative string) ([]byte, error)

	// Walk is like filepath.Walk, without the ordering guarantee.
	Walk(relative string, walkFn filepath.WalkFunc) error

	// ListDir returns the content of a single directory inside the snap.
	ListDir(path string) ([]string, error)

	// Install copies the snap file to targetPath (and possibly unpacks it to mountDir)
	Install(targetPath, mountDir string) error

	// Unpack unpacks the src parts to the dst directory
	Unpack(src, dst string) error
}

// backend implements a specific snap format
type snapFormat struct {
	magic []byte
	open  func(fn string) (Container, error)
}

// formatHandlers is the registry of known formats, squashfs is the only one atm.
var formatHandlers = []snapFormat{
	{squashfs.Magic, func(p string) (Container, error) {
		return squashfs.New(p), nil
	}},
}

// Open opens a given snap file with the right backend.
func Open(path string) (Container, error) {

	if osutil.IsDirectory(path) {
		if osutil.FileExists(filepath.Join(path, "meta", "snap.yaml")) {
			return snapdir.New(path), nil
		}

		return nil, NotSnapError{Path: path}
	}

	// open the file and check magic
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open snap: %v", err)
	}
	defer f.Close()

	header := make([]byte, 20)
	if _, err := f.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("cannot read snap: %v", err)
	}

	for _, h := range formatHandlers {
		if bytes.HasPrefix(header, h.magic) {
			return h.open(path)
		}
	}

	return nil, fmt.Errorf("cannot open snap: unknown header: %q", header)
}

var (
	// ErrBadModes is returned by ValidateContainer when the container has files with the wrong file modes for their role
	ErrBadModes = errors.New("snap is unusable due to bad permissions")
	// ErrMissingPaths is returned by ValidateContainer when the container is missing required files or directories
	ErrMissingPaths = errors.New("snap is unusable due to missing files")
)

// ValidateContainer does a minimal sanity check on the container.
func ValidateContainer(c Container, s *Info, logf func(format string, v ...interface{})) error {
	// needsrx keeps track of things that need to have at least 0555 perms
	needsrx := map[string]bool{
		".":    true,
		"meta": true,
	}
	// needsx keeps track of things that need to have at least 0111 perms
	needsx := map[string]bool{}
	// needsr keeps track of things that need to have at least 0444 perms
	needsr := map[string]bool{
		"meta/snap.yaml": true,
	}
	// needsf keeps track of things that need to be regular files (or symlinks to regular files)
	needsf := map[string]bool{}
	// noskipd tracks directories we want to descend into despite not being in needs*
	noskipd := map[string]bool{}

	for _, app := range s.Apps {
		// for non-services, paths go into the needsrx bag because users
		// need rx perms to execute it
		bag := needsrx
		paths := []string{app.Command}
		if app.IsService() {
			// services' paths just need to not be skipped by the validator
			bag = noskipd
			// additional paths to check for services:
			// XXX maybe have a method on app to keep this in sync
			paths = append(paths, app.StopCommand, app.ReloadCommand, app.PostStopCommand)
		}

		for _, path := range paths {
			path = normPath(path)
			if path == "" {
				continue
			}

			needsf[path] = true
			if app.IsService() {
				needsx[path] = true
			}
			for ; path != "."; path = filepath.Dir(path) {
				bag[path] = true
			}
		}

		// completer is special :-/
		if path := normPath(app.Completer); path != "" {
			needsr[path] = true
			for path = filepath.Dir(path); path != "."; path = filepath.Dir(path) {
				needsrx[path] = true
			}
		}
	}
	// note all needsr so far need to be regular files (or symlinks)
	for k := range needsr {
		needsf[k] = true
	}
	// thing can get jumbled up
	for path := range needsrx {
		delete(needsx, path)
		delete(needsr, path)
	}
	for path := range needsx {
		if needsr[path] {
			delete(needsx, path)
			delete(needsr, path)
			needsrx[path] = true
		}
	}
	seen := make(map[string]bool, len(needsx)+len(needsrx)+len(needsr))

	// bad modes are logged instead of being returned because the end user
	// can do nothing with the info (and the developer can read the logs)
	hasBadModes := false
	err := c.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		mode := info.Mode()
		if needsrx[path] || needsx[path] || needsr[path] {
			seen[path] = true
		}
		if !needsrx[path] && !needsx[path] && !needsr[path] && !strings.HasPrefix(path, "meta/") {
			if mode.IsDir() {
				if noskipd[path] {
					return nil
				}
				return filepath.SkipDir
			}
			return nil
		}

		if needsrx[path] || mode.IsDir() {
			if mode.Perm()&0555 != 0555 {
				// TODO parallel-install: use of proper instance/store name
				logf("in snap %q: %q should be world-readable and executable, and isn't: %s", s.InstanceName(), path, mode)
				hasBadModes = true
			}
		} else {
			if needsf[path] {
				// this assumes that if it's a symlink it's OK. Arguably we
				// should instead follow the symlink.  We'd have to expose
				// Lstat(), and guard against loops, and ...  huge can of
				// worms, and as this validator is meant as a developer aid
				// more than anything else, not worth it IMHO (as I can't
				// imagine this happening by accident).
				if mode&(os.ModeDir|os.ModeNamedPipe|os.ModeSocket|os.ModeDevice) != 0 {
					// TODO parallel-install: use of proper instance/store name
					logf("in snap %q: %q should be a regular file (or a symlink) and isn't", s.InstanceName(), path)
					hasBadModes = true
				}
			}
			if needsx[path] || strings.HasPrefix(path, "meta/hooks/") {
				if mode.Perm()&0111 == 0 {
					// TODO parallel-install: use of proper instance/store name
					logf("in snap %q: %q should be executable, and isn't: %s", s.InstanceName(), path, mode)
					hasBadModes = true
				}
			} else {
				// in needsr, or under meta but not a hook
				if mode.Perm()&0444 != 0444 {
					// TODO parallel-install: use of proper instance/store name
					logf("in snap %q: %q should be world-readable, and isn't: %s", s.InstanceName(), path, mode)
					hasBadModes = true
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(needsx)+len(needsrx)+len(needsr) {
		for _, needs := range []map[string]bool{needsx, needsrx, needsr} {
			for path := range needs {
				if !seen[path] {
					// TODO parallel-install: use of proper instance/store name
					logf("in snap %q: path %q does not exist", s.InstanceName(), path)
				}
			}
		}
		return ErrMissingPaths
	}

	if hasBadModes {
		return ErrBadModes
	}
	return nil
}

// normPath is a helper for validateContainer. It takes a relative path (e.g. an
// app's RestartCommand, which might be empty to mean there is no such thing),
// and cleans it.
//
// * empty paths are returned as is
// * if the path is not relative, it's initial / is dropped
// * if the path goes "outside" (ie starts with ../), the empty string is
//   returned (i.e. "ignore")
// * if there's a space in the command, ignore the rest of the string
//   (see also cmd/snap-exec/main.go's comment about strings.Split)
func normPath(path string) string {
	if path == "" {
		return ""
	}

	path = strings.TrimPrefix(filepath.Clean(path), "/")
	if strings.HasPrefix(path, "../") {
		// not something inside the snap
		return ""
	}
	if idx := strings.IndexByte(path, ' '); idx > -1 {
		return path[:idx]
	}

	return path
}
