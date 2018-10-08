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

package dirs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snapcore/snapd/release"
)

// the various file paths
var (
	GlobalRootDir string

	SnapMountDir string

	DistroLibExecDir string

	SnapBlobDir               string
	SnapDataDir               string
	SnapDataHomeGlob          string
	SnapDownloadCacheDir      string
	SnapAppArmorDir           string
	AppArmorCacheDir          string
	SnapAppArmorAdditionalDir string
	SnapConfineAppArmorDir    string
	SnapSeccompDir            string
	SnapMountPolicyDir        string
	SnapUdevRulesDir          string
	SnapKModModulesDir        string
	LocaleDir                 string
	SnapMetaDir               string
	SnapdSocket               string
	SnapSocket                string
	SnapRunDir                string
	SnapRunNsDir              string
	SnapRunLockDir            string

	SnapSeedDir   string
	SnapDeviceDir string

	SnapAssertsDBDir      string
	SnapCookieDir         string
	SnapTrustedAccountKey string
	SnapAssertsSpoolDir   string
	SnapSeqDir            string

	SnapStateFile     string
	SnapSystemKeyFile string

	SnapRepairDir        string
	SnapRepairStateFile  string
	SnapRepairRunDir     string
	SnapRepairAssertsDir string
	SnapRunRepairDir     string

	SnapCacheDir     string
	SnapNamesFile    string
	SnapSectionsFile string
	SnapCommandsDB   string

	SnapBinariesDir     string
	SnapServicesDir     string
	SnapSystemdConfDir  string
	SnapDesktopFilesDir string
	SnapBusPolicyDir    string

	SystemApparmorDir      string
	SystemApparmorCacheDir string

	CloudMetaDataFile     string
	CloudInstanceDataFile string

	ClassicDir string

	XdgRuntimeDirBase string
	XdgRuntimeDirGlob string

	CompletionHelper string
	CompletersDir    string
	CompleteSh       string

	SystemFontsDir           string
	SystemLocalFontsDir      string
	SystemFontconfigCacheDir string

	FreezerCgroupDir string
	SnapshotsDir     string

	ErrtrackerDbDir string
	SysfsDir        string
)

const (
	defaultSnapMountDir = "/snap"

	// These are directories which are static inside the core snap and
	// can never be prefixed as they will be always absolute once we
	// are in the snap confinement environment.
	CoreLibExecDir   = "/usr/lib/snapd"
	CoreSnapMountDir = "/snap"

	// Directory with snap data inside user's home
	UserHomeSnapDir = "snap"
)

var (
	// not exported because it does not honor the global rootdir
	snappyDir = filepath.Join("var", "lib", "snapd")
)

func init() {
	// init the global directories at startup
	root := os.Getenv("SNAPPY_GLOBAL_ROOT")

	SetRootDir(root)
}

// StripRootDir strips the custom global root directory from the specified argument.
func StripRootDir(dir string) string {
	if !filepath.IsAbs(dir) {
		panic(fmt.Sprintf("supplied path is not absolute %q", dir))
	}
	if !strings.HasPrefix(dir, GlobalRootDir) {
		panic(fmt.Sprintf("supplied path is not related to global root %q", dir))
	}
	result, err := filepath.Rel(GlobalRootDir, dir)
	if err != nil {
		panic(err)
	}
	return "/" + result
}

// SupportsClassicConfinement returns true if the current directory layout supports classic confinement.
func SupportsClassicConfinement() bool {
	smd := filepath.Join(GlobalRootDir, defaultSnapMountDir)
	if SnapMountDir == smd {
		return true
	}

	// distros with a non-default /snap location may still be good
	// if there is a symlink in place that links from the
	// defaultSnapMountDir (/snap) to the distro specific mount dir
	fi, err := os.Lstat(smd)
	if err == nil && fi.Mode()&os.ModeSymlink != 0 {
		if target, err := filepath.EvalSymlinks(smd); err == nil {
			if target == SnapMountDir {
				return true
			}
		}
	}

	return false
}

var metaSnapPath = "/meta/snap.yaml"

// isInsideBaseSnap returns true if the process is inside a base snap environment.
//
// The things that count as a base snap are:
// - any base snap mounted at /
// - any os snap mounted at /
func isInsideBaseSnap() (bool, error) {
	_, err := os.Stat(metaSnapPath)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// SetRootDir allows settings a new global root directory, this is useful
// for e.g. chroot operations
func SetRootDir(rootdir string) {
	if rootdir == "" {
		rootdir = "/"
	}
	GlobalRootDir = rootdir

	isInsideBase, _ := isInsideBaseSnap()
	if !isInsideBase && release.DistroLike("fedora", "arch", "manjaro", "antergos") {
		SnapMountDir = filepath.Join(rootdir, "/var/lib/snapd/snap")
	} else {
		SnapMountDir = filepath.Join(rootdir, defaultSnapMountDir)
	}

	SnapDataDir = filepath.Join(rootdir, "/var/snap")
	SnapDataHomeGlob = filepath.Join(rootdir, "/home/*/", UserHomeSnapDir)
	SnapAppArmorDir = filepath.Join(rootdir, snappyDir, "apparmor", "profiles")
	SnapConfineAppArmorDir = filepath.Join(rootdir, snappyDir, "apparmor", "snap-confine")
	AppArmorCacheDir = filepath.Join(rootdir, "/var/cache/apparmor")
	SnapAppArmorAdditionalDir = filepath.Join(rootdir, snappyDir, "apparmor", "additional")
	SnapDownloadCacheDir = filepath.Join(rootdir, snappyDir, "cache")
	SnapSeccompDir = filepath.Join(rootdir, snappyDir, "seccomp", "bpf")
	SnapMountPolicyDir = filepath.Join(rootdir, snappyDir, "mount")
	SnapMetaDir = filepath.Join(rootdir, snappyDir, "meta")
	SnapBlobDir = filepath.Join(rootdir, snappyDir, "snaps")
	SnapDesktopFilesDir = filepath.Join(rootdir, snappyDir, "desktop", "applications")
	SnapRunDir = filepath.Join(rootdir, "/run/snapd")
	SnapRunNsDir = filepath.Join(SnapRunDir, "/ns")
	SnapRunLockDir = filepath.Join(SnapRunDir, "/lock")

	// keep in sync with the debian/snapd.socket file:
	SnapdSocket = filepath.Join(rootdir, "/run/snapd.socket")
	SnapSocket = filepath.Join(rootdir, "/run/snapd-snap.socket")

	SnapAssertsDBDir = filepath.Join(rootdir, snappyDir, "assertions")
	SnapCookieDir = filepath.Join(rootdir, snappyDir, "cookie")
	SnapAssertsSpoolDir = filepath.Join(rootdir, "run/snapd/auto-import")
	SnapSeqDir = filepath.Join(rootdir, snappyDir, "sequence")

	SnapStateFile = filepath.Join(rootdir, snappyDir, "state.json")
	SnapSystemKeyFile = filepath.Join(rootdir, snappyDir, "system-key")

	SnapCacheDir = filepath.Join(rootdir, "/var/cache/snapd")
	SnapNamesFile = filepath.Join(SnapCacheDir, "names")
	SnapSectionsFile = filepath.Join(SnapCacheDir, "sections")
	SnapCommandsDB = filepath.Join(SnapCacheDir, "commands.db")

	SnapSeedDir = filepath.Join(rootdir, snappyDir, "seed")
	SnapDeviceDir = filepath.Join(rootdir, snappyDir, "device")

	SnapRepairDir = filepath.Join(rootdir, snappyDir, "repair")
	SnapRepairStateFile = filepath.Join(SnapRepairDir, "repair.json")
	SnapRepairRunDir = filepath.Join(SnapRepairDir, "run")
	SnapRepairAssertsDir = filepath.Join(SnapRepairDir, "assertions")
	SnapRunRepairDir = filepath.Join(SnapRunDir, "repair")

	SnapBinariesDir = filepath.Join(SnapMountDir, "bin")
	SnapServicesDir = filepath.Join(rootdir, "/etc/systemd/system")
	SnapSystemdConfDir = filepath.Join(rootdir, "/etc/systemd/system.conf.d")
	SnapBusPolicyDir = filepath.Join(rootdir, "/etc/dbus-1/system.d")

	SystemApparmorDir = filepath.Join(rootdir, "/etc/apparmor.d")
	SystemApparmorCacheDir = filepath.Join(rootdir, "/etc/apparmor.d/cache")

	CloudMetaDataFile = filepath.Join(rootdir, "/var/lib/cloud/seed/nocloud-net/meta-data")
	CloudInstanceDataFile = filepath.Join(rootdir, "/run/cloud-init/instance-data.json")

	SnapUdevRulesDir = filepath.Join(rootdir, "/etc/udev/rules.d")

	SnapKModModulesDir = filepath.Join(rootdir, "/etc/modules-load.d/")

	LocaleDir = filepath.Join(rootdir, "/usr/share/locale")
	ClassicDir = filepath.Join(rootdir, "/writable/classic")

	if release.DistroLike("fedora") {
		// rhel, centos, fedora and derivatives
		// both rhel and centos list "fedora" in ID_LIKE
		DistroLibExecDir = filepath.Join(rootdir, "/usr/libexec/snapd")
	} else {
		DistroLibExecDir = filepath.Join(rootdir, "/usr/lib/snapd")
	}

	XdgRuntimeDirBase = filepath.Join(rootdir, "/run/user")
	XdgRuntimeDirGlob = filepath.Join(rootdir, XdgRuntimeDirBase, "*/")

	CompletionHelper = filepath.Join(CoreLibExecDir, "etelpmoc.sh")
	CompletersDir = filepath.Join(rootdir, "/usr/share/bash-completion/completions/")
	CompleteSh = filepath.Join(SnapMountDir, "core/current/usr/lib/snapd/complete.sh")

	SystemFontsDir = filepath.Join(rootdir, "/usr/share/fonts")
	SystemLocalFontsDir = filepath.Join(rootdir, "/usr/local/share/fonts")
	SystemFontconfigCacheDir = filepath.Join(rootdir, "/var/cache/fontconfig")

	FreezerCgroupDir = filepath.Join(rootdir, "/sys/fs/cgroup/freezer/")
	SnapshotsDir = filepath.Join(rootdir, snappyDir, "snapshots")

	ErrtrackerDbDir = filepath.Join(rootdir, snappyDir, "errtracker.db")
	SysfsDir = filepath.Join(rootdir, "/sys")
}
