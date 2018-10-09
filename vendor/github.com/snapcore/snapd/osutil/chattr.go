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

package osutil

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	// from /usr/include/linux/fs.h
	FS_SECRM_FL        = 0x00000001 /* Secure deletion */
	FS_UNRM_FL         = 0x00000002 /* Undelete */
	FS_COMPR_FL        = 0x00000004 /* Compress file */
	FS_SYNC_FL         = 0x00000008 /* Synchronous updates */
	FS_IMMUTABLE_FL    = 0x00000010 /* Immutable file */
	FS_APPEND_FL       = 0x00000020 /* writes to file may only append */
	FS_NODUMP_FL       = 0x00000040 /* do not dump file */
	FS_NOATIME_FL      = 0x00000080 /* do not update atime */
	FS_DIRTY_FL        = 0x00000100
	FS_COMPRBLK_FL     = 0x00000200 /* One or more compressed clusters */
	FS_NOCOMP_FL       = 0x00000400 /* Don't compress */
	FS_ECOMPR_FL       = 0x00000800 /* Compression error */
	FS_BTREE_FL        = 0x00001000 /* btree format dir */
	FS_INDEX_FL        = 0x00001000 /* hash-indexed directory */
	FS_IMAGIC_FL       = 0x00002000 /* AFS directory */
	FS_JOURNAL_DATA_FL = 0x00004000 /* Reserved for ext3 */
	FS_NOTAIL_FL       = 0x00008000 /* file tail should not be merged */
	FS_DIRSYNC_FL      = 0x00010000 /* dirsync behaviour (directories only) */
	FS_TOPDIR_FL       = 0x00020000 /* Top of directory hierarchies*/
	FS_EXTENT_FL       = 0x00080000 /* Extents */
	FS_DIRECTIO_FL     = 0x00100000 /* Use direct i/o */
	FS_NOCOW_FL        = 0x00800000 /* Do not cow file */
	FS_PROJINHERIT_FL  = 0x20000000 /* Create with parents projid */
	FS_RESERVED_FL     = 0x80000000 /* reserved for ext2 lib */
)

func ioctl(f *os.File, request uintptr, attrp *int32) error {
	argp := uintptr(unsafe.Pointer(attrp))
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), request, argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}

	return nil
}

// GetAttr retrieves the attributes of a file on a linux filesystem
func GetAttr(f *os.File) (int32, error) {
	attr := int32(-1)
	err := ioctl(f, _FS_IOC_GETFLAGS, &attr)
	return attr, err
}

// SetAttr sets the attributes of a file on a linux filesystem to the given value
func SetAttr(f *os.File, attr int32) error {
	return ioctl(f, _FS_IOC_SETFLAGS, &attr)
}
