// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
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
	"io/ioutil"

	"github.com/snapcore/snapd/osutil"
)

// addImplicitHooks adds hooks from the installed snap's hookdir to the snap info.
//
// Existing hooks (i.e. ones defined in the YAML) are not changed; only missing
// hooks are added.
func addImplicitHooks(snapInfo *Info) error {
	// First of all, check to ensure the hooks directory exists. If it doesn't,
	// it's not an error-- there's just nothing to do.
	hooksDir := snapInfo.HooksDir()
	if !osutil.IsDirectory(hooksDir) {
		return nil
	}

	fileInfos, err := ioutil.ReadDir(hooksDir)
	if err != nil {
		return fmt.Errorf("unable to read hooks directory: %s", err)
	}

	for _, fileInfo := range fileInfos {
		addHookIfValid(snapInfo, fileInfo.Name())
	}

	return nil
}

// addImplicitHooksFromContainer adds hooks from the snap file's hookdir to the snap info.
//
// Existing hooks (i.e. ones defined in the YAML) are not changed; only missing
// hooks are added.
func addImplicitHooksFromContainer(snapInfo *Info, snapf Container) error {
	// Read the hooks directory. If this fails we assume the hooks directory
	// doesn't exist, which means there are no implicit hooks to load (not an
	// error).
	fileNames, err := snapf.ListDir("meta/hooks")
	if err != nil {
		return nil
	}

	for _, fileName := range fileNames {
		addHookIfValid(snapInfo, fileName)
	}

	return nil
}

func addHookIfValid(snapInfo *Info, hookName string) {
	// Verify that the hook name is actually supported. If not, ignore it.
	if !IsHookSupported(hookName) {
		return
	}

	// Don't overwrite a hook that has already been loaded from the YAML
	if _, ok := snapInfo.Hooks[hookName]; !ok {
		snapInfo.Hooks[hookName] = &HookInfo{Snap: snapInfo, Name: hookName}
	}
}
