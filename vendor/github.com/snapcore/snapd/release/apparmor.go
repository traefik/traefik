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

package release

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ApparmorLevelType encodes the kind of support for apparmor
// found on this system.
type AppArmorLevelType int

const (
	// NoAppArmor indicates that apparmor is not enabled.
	NoAppArmor AppArmorLevelType = iota
	// PartialAppArmor indicates that apparmor is enabled but some
	// features are missing.
	PartialAppArmor
	// FullAppArmor indicates that all features are supported.
	FullAppArmor
)

var (
	appArmorLevel   AppArmorLevelType
	appArmorSummary string
)

func init() {
	appArmorLevel, appArmorSummary = probeAppArmor()
}

// AppArmorLevel quantifies how well apparmor is supported on the
// current kernel.
func AppArmorLevel() AppArmorLevelType {
	return appArmorLevel
}

// AppArmorSummary describes how well apparmor is supported on the
// current kernel.
func AppArmorSummary() string {
	return appArmorSummary
}

// MockAppArmorSupportLevel makes the system believe it has certain
// level of apparmor support.
func MockAppArmorLevel(level AppArmorLevelType) (restore func()) {
	oldAppArmorLevel := appArmorLevel
	oldAppArmorSummary := appArmorSummary
	appArmorLevel = level
	appArmorSummary = fmt.Sprintf("mocked apparmor level: %v", level)
	return func() {
		appArmorLevel = oldAppArmorLevel
		appArmorSummary = oldAppArmorSummary
	}
}

// probe related code
var (
	appArmorFeaturesSysPath  = "/sys/kernel/security/apparmor/features"
	requiredAppArmorFeatures = []string{
		"caps",
		"dbus",
		"domain",
		"file",
		"mount",
		"namespaces",
		"network",
		"ptrace",
		"signal",
	}
)

// isDirectoy is like osutil.IsDirectory but we cannot import this
// because of import cycles
func isDirectory(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func probeAppArmor() (AppArmorLevelType, string) {
	if !isDirectory(appArmorFeaturesSysPath) {
		return NoAppArmor, "apparmor not enabled"
	}
	var missing []string
	for _, feature := range requiredAppArmorFeatures {
		if !isDirectory(filepath.Join(appArmorFeaturesSysPath, feature)) {
			missing = append(missing, feature)
		}
	}
	if len(missing) > 0 {
		return PartialAppArmor, fmt.Sprintf("apparmor is enabled but some features are missing: %s", strings.Join(missing, ", "))
	}
	return FullAppArmor, "apparmor is enabled and all features are available"
}

// AppArmorFeatures returns a sorted list of apparmor features like
// []string{"dbus", "network"}.
func AppArmorFeatures() []string {
	// note that ioutil.ReadDir() is already sorted
	dentries, err := ioutil.ReadDir(appArmorFeaturesSysPath)
	if err != nil {
		return nil
	}
	appArmorFeatures := make([]string, 0, len(dentries))
	for _, f := range dentries {
		if isDirectory(filepath.Join(appArmorFeaturesSysPath, f.Name())) {
			appArmorFeatures = append(appArmorFeatures, f.Name())
		}
	}
	return appArmorFeatures
}
