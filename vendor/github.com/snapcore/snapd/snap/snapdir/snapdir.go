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

package snapdir

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// SnapDir is the snapdir based snap.
type SnapDir struct {
	path string
}

// Path returns the path of the backing container.
func (s *SnapDir) Path() string {
	return s.path
}

// New returns a new snap directory container.
func New(path string) *SnapDir {
	return &SnapDir{path: path}
}

func (s *SnapDir) Size() (size int64, err error) {
	totalSize := int64(0)
	f := func(_ string, info os.FileInfo, err error) error {
		totalSize += info.Size()
		return err
	}
	filepath.Walk(s.path, f)

	return totalSize, nil
}

func (s *SnapDir) Install(targetPath, mountDir string) error {
	return os.Symlink(s.path, targetPath)
}

func (s *SnapDir) ReadFile(file string) (content []byte, err error) {
	return ioutil.ReadFile(filepath.Join(s.path, file))
}

func littleWalk(dirPath string, dirHandle *os.File, dirstack *[]string, walkFn filepath.WalkFunc) error {
	const numSt = 100

	sts, err := dirHandle.Readdir(numSt)
	if err != nil {
		return err
	}
	for _, st := range sts {
		path := filepath.Join(dirPath, st.Name())
		if err := walkFn(path, st, nil); err != nil {
			if st.IsDir() && err == filepath.SkipDir {
				// caller wants to skip this directory
				continue
			}
			return err
		} else if st.IsDir() {
			*dirstack = append(*dirstack, path)
		}
	}

	return nil
}

// Walk (part of snap.Container) is like filepath.Walk, without the ordering guarantee.
func (s *SnapDir) Walk(relative string, walkFn filepath.WalkFunc) error {
	relative = filepath.Clean(relative)
	if relative == "" || relative == "/" {
		relative = "."
	} else if relative[0] == '/' {
		// I said relative, darn it :-)
		relative = relative[1:]
	}
	root := filepath.Join(s.path, relative)
	// we could just filepath.Walk(root, walkFn), but that doesn't scale
	// well to insanely big directories as it reads the whole directory,
	// in order to sort it. This Walk doesn't do that.
	//
	// Also the directory is always relative to the top of the container
	// for us, which would make it a little more messy to get right.
	f, err := os.Open(root)
	if err != nil {
		return walkFn(relative, nil, err)
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	st, err := f.Stat()
	if err != nil {
		return walkFn(relative, nil, err)
	}

	err = walkFn(relative, st, nil)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return nil
	}

	var dirstack []string
	for {
		if err := littleWalk(relative, f, &dirstack, walkFn); err != nil {
			if err != io.EOF {
				err = walkFn(relative, nil, err)
				if err != nil {
					return err
				}
			}
			if len(dirstack) == 0 {
				// finished
				break
			}
			f.Close()
			f = nil
			for f == nil && len(dirstack) > 0 {
				relative = dirstack[0]
				f, err = os.Open(filepath.Join(s.path, relative))
				if err != nil {
					err = walkFn(relative, nil, err)
					if err != nil {
						return err
					}
				}
				dirstack = dirstack[1:]
			}
			if f == nil {
				break
			}
			continue
		}
	}

	return nil
}

func (s *SnapDir) ListDir(path string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(filepath.Join(s.path, path))
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, fileInfo := range fileInfos {
		fileNames = append(fileNames, fileInfo.Name())
	}

	return fileNames, nil
}

func (s *SnapDir) Unpack(src, dstDir string) error {
	return fmt.Errorf("unpack is not supported with snaps of type snapdir")
}
