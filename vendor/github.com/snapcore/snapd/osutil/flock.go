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
	"errors"
	"os"
	"syscall"
)

// FileLock describes a file system lock
type FileLock struct {
	file *os.File
}

var ErrAlreadyLocked = errors.New("cannot acquire lock, already locked")

// NewFileLock creates and opens the lock file given by "path"
func NewFileLock(path string) (*FileLock, error) {
	mode := syscall.O_RDWR | syscall.O_CREAT | syscall.O_NOFOLLOW | syscall.O_CLOEXEC
	file, err := os.OpenFile(path, mode, os.FileMode(0600))
	if err != nil {
		return nil, err
	}
	l := &FileLock{file: file}
	return l, nil
}

// Path returns the path of the lock file.
func (l *FileLock) Path() string {
	return l.file.Name()
}

// Close closes the lock, unlocking it automatically if needed.
func (l *FileLock) Close() error {
	return l.file.Close()
}

// Lock acquires an exclusive lock and blocks until the lock is free.
func (l *FileLock) Lock() error {
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX)
}

// TryLock acquires an exclusive lock and errors if the lock cannot be acquired.
func (l *FileLock) TryLock() error {
	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == syscall.EWOULDBLOCK {
		return ErrAlreadyLocked
	}
	return err
}

// Unlock releases an acquired lock.
func (l *FileLock) Unlock() error {
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
}
