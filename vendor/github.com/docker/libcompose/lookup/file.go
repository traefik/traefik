package lookup

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
)

// relativePath returns the proper relative path for the given file path. If
// the relativeTo string equals "-", then it means that it's from the stdin,
// and the returned path will be the current working directory. Otherwise, if
// file is really an absolute path, then it will be returned without any
// changes. Otherwise, the returned path will be a combination of relativeTo
// and file.
func relativePath(file, relativeTo string) string {
	// stdin: return the current working directory if possible.
	if relativeTo == "-" {
		if cwd, err := os.Getwd(); err == nil {
			return filepath.Join(cwd, file)
		}
	}

	// If the given file is already an absolute path, just return it.
	// Otherwise, the returned path will be relative to the given relativeTo
	// path.
	if filepath.IsAbs(file) {
		return file
	}

	abs, err := filepath.Abs(filepath.Join(path.Dir(relativeTo), file))
	if err != nil {
		logrus.Errorf("Failed to get absolute directory: %s", err)
		return file
	}
	return abs
}

// FileResourceLookup is a "bare" structure that implements the project.ResourceLookup interface
type FileResourceLookup struct {
}

// Lookup returns the content and the actual filename of the file that is "built" using the
// specified file and relativeTo string. file and relativeTo are supposed to be file path.
// If file starts with a slash ('/'), it tries to load it, otherwise it will build a
// filename using the folder part of relativeTo joined with file.
func (f *FileResourceLookup) Lookup(file, relativeTo string) ([]byte, string, error) {
	file = relativePath(file, relativeTo)
	logrus.Debugf("Reading file %s", file)
	bytes, err := ioutil.ReadFile(file)
	return bytes, file, err
}

// ResolvePath returns the path to be used for the given path volume. This
// function already takes care of relative paths.
func (f *FileResourceLookup) ResolvePath(path, relativeTo string) string {
	vs := strings.SplitN(path, ":", 2)
	if len(vs) != 2 || filepath.IsAbs(vs[0]) {
		return path
	}
	vs[0] = relativePath(vs[0], relativeTo)
	return strings.Join(vs, ":")
}
