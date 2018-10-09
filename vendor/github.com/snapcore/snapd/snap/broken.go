// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/snapcore/snapd/dirs"
)

// GuessAppsForBroken guesses what apps and services a broken snap has
// on the system by searching for matches based on the snap name in
// the snap binaries and service file directories. It returns a
// mapping from app names to partial AppInfo.
func GuessAppsForBroken(info *Info) map[string]*AppInfo {
	out := make(map[string]*AppInfo)

	// guess binaries first
	name := info.InstanceName()
	for _, p := range []string{name, fmt.Sprintf("%s.*", name)} {
		matches, _ := filepath.Glob(filepath.Join(dirs.SnapBinariesDir, p))
		for _, m := range matches {
			l := strings.SplitN(filepath.Base(m), ".", 2)
			var appname string
			if len(l) == 1 {
				// when app is named the same as snap, it will
				// be available under '<snap>' name, if the snap
				// was installed with instance key, the app will
				// be named `<snap>_<instance>'
				appname = InstanceSnap(l[0])
			} else {
				appname = l[1]
			}
			out[appname] = &AppInfo{
				Snap: info,
				Name: appname,
			}
		}
	}

	// guess the services next
	matches, _ := filepath.Glob(filepath.Join(dirs.SnapServicesDir, fmt.Sprintf("snap.%s.*.service", name)))
	for _, m := range matches {
		appname := strings.Split(m, ".")[2]
		out[appname] = &AppInfo{
			Snap:   info,
			Name:   appname,
			Daemon: "simple",
		}
	}

	return out
}

// renameClashingCorePlugs renames plugs that clash with slot names on core snap.
//
// Some released core snaps had explicitly defined plugs "network-bind" and
// "core-support" that clashed with implicit slots with the same names but this
// was not validated before.  To avoid a flag day and any potential issues,
// transparently rename the two clashing plugs by appending the "-plug" suffix.
func (info *Info) renameClashingCorePlugs() {
	if info.InstanceName() == "core" && info.Type == TypeOS {
		for _, plugName := range []string{"network-bind", "core-support"} {
			info.forceRenamePlug(plugName, plugName+"-plug")
		}
	}
}

// forceRenamePlug renames the plug from oldName to newName, if present.
func (info *Info) forceRenamePlug(oldName, newName string) {
	if plugInfo, ok := info.Plugs[oldName]; ok {
		delete(info.Plugs, oldName)
		info.Plugs[newName] = plugInfo
		plugInfo.Name = newName
		for _, appInfo := range info.Apps {
			if _, ok := appInfo.Plugs[oldName]; ok {
				delete(appInfo.Plugs, oldName)
				appInfo.Plugs[newName] = plugInfo
			}
		}
		for _, hookInfo := range info.Hooks {
			if _, ok := hookInfo.Plugs[oldName]; ok {
				delete(hookInfo.Plugs, oldName)
				hookInfo.Plugs[newName] = plugInfo
			}
		}
	}
}
