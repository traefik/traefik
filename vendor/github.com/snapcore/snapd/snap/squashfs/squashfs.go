// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2018 Canonical Ltd
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

package squashfs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"time"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/strutil"
)

// Magic is the magic prefix of squashfs snap files.
var Magic = []byte{'h', 's', 'q', 's'}

// Snap is the squashfs based snap.
type Snap struct {
	path string
}

// Path returns the path of the backing file.
func (s *Snap) Path() string {
	return s.path
}

// New returns a new Squashfs snap.
func New(snapPath string) *Snap {
	return &Snap{path: snapPath}
}

var osLink = os.Link

func (s *Snap) Install(targetPath, mountDir string) error {

	// ensure mount-point and blob target dir.
	for _, dir := range []string{mountDir, filepath.Dir(targetPath)} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// This is required so that the tests can simulate a mounted
	// snap when we "install" a squashfs snap in the tests.
	// We can not mount it for real in the tests, so we just unpack
	// it to the location which is good enough for the tests.
	if osutil.GetenvBool("SNAPPY_SQUASHFS_UNPACK_FOR_TESTS") {
		if err := s.Unpack("*", mountDir); err != nil {
			return err
		}
	}

	// nothing to do, happens on e.g. first-boot when we already
	// booted with the OS snap but its also in the seed.yaml
	if s.path == targetPath || osutil.FilesAreEqual(s.path, targetPath) {
		return nil
	}

	// try to (hard)link the file, but go on to trying to copy it
	// if it fails for whatever reason
	//
	// link(2) returns EPERM on filesystems that don't support
	// hard links (like vfat), so checking the error here doesn't
	// make sense vs just trying to copy it.
	if err := osLink(s.path, targetPath); err == nil {
		return nil
	}

	// if the file is a seed, but the hardlink failed, symlinking it
	// saves the copy (which in livecd is expensive) so try that next
	if filepath.Dir(s.path) == dirs.SnapSeedDir && os.Symlink(s.path, targetPath) == nil {
		return nil
	}

	return osutil.CopyFile(s.path, targetPath, osutil.CopyFlagPreserveAll|osutil.CopyFlagSync)
}

// unsquashfsStderrWriter is a helper that captures errors from
// unsquashfs on stderr. Because unsquashfs will potentially
// (e.g. on out-of-diskspace) report an error on every single
// file we limit the reported error lines to 4.
//
// unsquashfs does not exit with an exit code for write errors
// (e.g. no space left on device). There is an upstream PR
// to fix this https://github.com/plougher/squashfs-tools/pull/46
//
// However in the meantime we can detect errors by looking
// on stderr for "failed" which is pretty consistently used in
// the unsquashfs.c source in case of errors.
type unsquashfsStderrWriter struct {
	strutil.MatchCounter
}

func newUnsquashfsStderrWriter() *unsquashfsStderrWriter {
	return &unsquashfsStderrWriter{strutil.MatchCounter{
		Regexp: regexp.MustCompile(`(?m).*\b[Ff]ailed\b.*`),
		N:      4, // note Err below uses this value
	}}
}

func (u *unsquashfsStderrWriter) Err() error {
	// here we use that our N is 4.
	errors, count := u.Matches()
	switch count {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("failed: %q", errors[0])
	case 2, 3, 4:
		return fmt.Errorf("failed: %s, and %q", strutil.Quoted(errors[:len(errors)-1]), errors[len(errors)-1])
	default:
		// count > len(matches)
		extra := count - len(errors)
		return fmt.Errorf("failed: %s, and %d more", strutil.Quoted(errors), extra)
	}
}

func (s *Snap) Unpack(src, dstDir string) error {
	usw := newUnsquashfsStderrWriter()

	cmd := exec.Command("unsquashfs", "-n", "-f", "-d", dstDir, s.path, src)
	cmd.Stderr = usw
	if err := cmd.Run(); err != nil {
		return err
	}
	if usw.Err() != nil {
		return fmt.Errorf("cannot extract %q to %q: %v", src, dstDir, usw.Err())
	}
	return nil
}

// Size returns the size of a squashfs snap.
func (s *Snap) Size() (size int64, err error) {
	st, err := os.Stat(s.path)
	if err != nil {
		return 0, err
	}

	return st.Size(), nil
}

// ReadFile returns the content of a single file inside a squashfs snap.
func (s *Snap) ReadFile(filePath string) (content []byte, err error) {
	tmpdir, err := ioutil.TempDir("", "read-file")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)

	unpackDir := filepath.Join(tmpdir, "unpack")
	if err := exec.Command("unsquashfs", "-n", "-i", "-d", unpackDir, s.path, filePath).Run(); err != nil {
		return nil, err
	}

	return ioutil.ReadFile(filepath.Join(unpackDir, filePath))
}

// skipper is used to track directories that should be skipped
//
// Given sk := make(skipper), if you sk.Add("foo/bar"), then
// sk.Has("foo/bar") is true, but also sk.Has("foo/bar/baz")
//
// It could also be a map[string]bool, but because it's only supposed
// to be checked through its Has method as above, the small added
// complexity of it being a map[string]struct{} lose to the associated
// space savings.
type skipper map[string]struct{}

func (sk skipper) Add(path string) {
	sk[filepath.Clean(path)] = struct{}{}
}

func (sk skipper) Has(path string) bool {
	for p := filepath.Clean(path); p != "." && p != "/"; p = filepath.Dir(p) {
		if _, ok := sk[p]; ok {
			return true
		}
	}

	return false
}

// Walk (part of snap.Container) is like filepath.Walk, without the ordering guarantee.
func (s *Snap) Walk(relative string, walkFn filepath.WalkFunc) error {
	relative = filepath.Clean(relative)
	if relative == "" || relative == "/" {
		relative = "."
	} else if relative[0] == '/' {
		// I said relative, darn it :-)
		relative = relative[1:]
	}

	var cmd *exec.Cmd
	if relative == "." {
		cmd = exec.Command("unsquashfs", "-no-progress", "-dest", ".", "-ll", s.path)
	} else {
		cmd = exec.Command("unsquashfs", "-no-progress", "-dest", ".", "-ll", s.path, relative)
	}
	cmd.Env = []string{"TZ=UTC"}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return walkFn(relative, nil, err)
	}
	if err := cmd.Start(); err != nil {
		return walkFn(relative, nil, err)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	// skip the header
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			break
		}
	}

	skipper := make(skipper)
	for scanner.Scan() {
		st, err := fromRaw(scanner.Bytes())
		if err != nil {
			err = walkFn(relative, nil, err)
			if err != nil {
				return err
			}
		} else {
			path := filepath.Join(relative, st.Path())
			if skipper.Has(path) {
				continue
			}
			err = walkFn(path, st, nil)
			if err != nil {
				if err == filepath.SkipDir && st.IsDir() {
					skipper.Add(path)
				} else {
					return err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return walkFn(relative, nil, err)
	}

	if err := cmd.Wait(); err != nil {
		return walkFn(relative, nil, err)
	}
	return nil
}

// ListDir returns the content of a single directory inside a squashfs snap.
func (s *Snap) ListDir(dirPath string) ([]string, error) {
	output, err := exec.Command(
		"unsquashfs", "-no-progress", "-dest", "_", "-l", s.path, dirPath).CombinedOutput()
	if err != nil {
		return nil, osutil.OutputErr(output, err)
	}

	prefixPath := path.Join("_", dirPath)
	pattern, err := regexp.Compile("(?m)^" + regexp.QuoteMeta(prefixPath) + "/([^/\r\n]+)$")
	if err != nil {
		return nil, fmt.Errorf("internal error: cannot compile squashfs list dir regexp for %q: %s", dirPath, err)
	}

	var directoryContents []string
	for _, groups := range pattern.FindAllSubmatch(output, -1) {
		if len(groups) > 1 {
			directoryContents = append(directoryContents, string(groups[1]))
		}
	}

	return directoryContents, nil
}

// Build builds the snap.
func (s *Snap) Build(buildDir string) error {
	fullSnapPath, err := filepath.Abs(s.path)
	if err != nil {
		return err
	}

	return osutil.ChDir(buildDir, func() error {
		output, err := exec.Command(
			"mksquashfs",
			".", fullSnapPath,
			"-noappend",
			"-comp", "xz",
			"-no-xattrs",
			"-no-fragments",
			"-no-progress",
		).CombinedOutput()
		if err != nil {
			return fmt.Errorf("mksquashfs call failed: %s", osutil.OutputErr(output, err))
		}

		return nil
	})
}

// BuildDate returns the "Creation or last append time" as reported by unsquashfs.
func (s *Snap) BuildDate() time.Time {
	return BuildDate(s.path)
}

// BuildDate returns the "Creation or last append time" as reported by unsquashfs.
func BuildDate(path string) time.Time {
	var t0 time.Time

	const prefix = "Creation or last append time "
	m := &strutil.MatchCounter{
		Regexp: regexp.MustCompile("(?m)^" + prefix + ".*$"),
		N:      1,
	}

	cmd := exec.Command("unsquashfs", "-n", "-s", path)
	cmd.Env = []string{"TZ=UTC"}
	cmd.Stdout = m
	cmd.Stderr = m
	if err := cmd.Run(); err != nil {
		return t0
	}
	matches, count := m.Matches()
	if count != 1 {
		return t0
	}
	t0, _ = time.Parse(time.ANSIC, matches[0][len(prefix):])
	return t0
}
