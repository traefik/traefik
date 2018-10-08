// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
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
	"strings"
)

// IsRootWritableOverlay detects if the current '/' is a writable overlay
// (fstype is 'overlay' and 'upperdir' is specified) and returns upperdir or
// the empty string if not used.
//
// Debian-based LiveCD systems use 'casper' to setup the mounts, and part of
// this setup involves running mount commands to mount / on /cow as overlay and
// results in AppArmor seeing '/upper' as the upperdir rather than '/cow/upper'
// as seen in mountinfo. By the time snapd is run, we don't have enough
// information to discover /cow through mount parent ID or st_dev (maj:min).
// While overlay doesn't use the mount source for anything itself, casper sets
// the mount source ('/cow' with the above) for its own purposes and we can
// leverage this by stripping the mount source from the beginning of upperdir.
//
// https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt
// man 5 proc
//
// Currently uses variables and Mock functions from nfs.go
func IsRootWritableOverlay() (string, error) {
	mountinfo, err := LoadMountInfo(procSelfMountInfo)
	if err != nil {
		return "", fmt.Errorf("cannot parse %s: %s", procSelfMountInfo, err)
	}
	for _, entry := range mountinfo {
		if entry.FsType == "overlay" && entry.MountDir == "/" {
			if dir, ok := entry.SuperOptions["upperdir"]; ok {
				// upperdir must be an absolute path without
				// any AppArmor regular expression (AARE)
				// characters or double quotes to be considered
				if !strings.HasPrefix(dir, "/") || strings.ContainsAny(dir, `?*[]{}^"`) {
					continue
				}
				// if mount source is path, strip it from dir
				// (for casper)
				if strings.HasPrefix(entry.MountSource, "/") {
					dir = strings.TrimPrefix(dir, strings.TrimRight(entry.MountSource, "/"))
				}

				dir = strings.TrimRight(dir, "/")

				// The resulting trimmed dir must be an
				// absolute path that is not '/'
				if len(dir) < 2 || !strings.HasPrefix(dir, "/") {
					continue
				}

				// Make sure trailing slashes are predicatably
				// missing
				return dir, nil
			}
		}
	}
	return "", nil
}
