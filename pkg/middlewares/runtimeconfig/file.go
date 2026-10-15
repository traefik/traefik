package runtimeconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maxFileSize = 1 << 20

// ReadFile reads an operator-configured regular file, bounded to 1 MiB.
// Symlinks must stay within its configured parent directory.
func ReadFile(path string) ([]byte, error) {
	values, err := ReadFiles(map[string]string{"value": path})
	if err != nil {
		return nil, err
	}
	return values["value"], nil
}

// ReadFiles reads credential files while pinning each projected Secret's
// ..data directory for the whole call, so rotation cannot mix revisions.
func ReadFiles(paths map[string]string) (map[string][]byte, error) {
	roots := make(map[string]*os.Root)
	snapshots := make(map[string]*os.Root)
	result := make(map[string][]byte, len(paths))
	for name, path := range paths {
		if path == "" {
			return nil, errors.New("runtime file path is empty")
		}
		directory, base := filepath.Dir(path), filepath.Base(path)
		root := roots[directory]
		if root == nil {
			var err error
			root, err = os.OpenRoot(directory)
			if err != nil {
				return nil, fmt.Errorf("opening runtime file directory: %w", err)
			}
			defer root.Close()
			roots[directory] = root
		}
		// Only projected keys use the pinned revision. Ordinary files beside a
		// projected volume must still refer to their own configured contents.
		if link, err := root.Readlink(base); err == nil && link == filepath.Join("..data", base) {
			snapshot := snapshots[directory]
			if snapshot == nil {
				snapshot, err = root.OpenRoot("..data")
				if err != nil {
					return nil, fmt.Errorf("opening projected runtime file directory: %w", err)
				}
				defer snapshot.Close()
				snapshots[directory] = snapshot
			}
			root = snapshot
		}
		data, err := readRegularFile(root, base)
		if err != nil {
			return nil, fmt.Errorf("reading runtime file %q: %w", path, err)
		}
		result[name] = data
	}
	return result, nil
}

func readRegularFile(root *os.Root, name string) ([]byte, error) {
	info, err := root.Stat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("runtime file must be a regular file")
	}
	if info.Size() > maxFileSize {
		return nil, errors.New("runtime file exceeds 1 MiB")
	}
	// Non-blocking open prevents a regular-file-to-FIFO replacement from
	// stalling the reload worker between Stat and OpenFile on Unix.
	file, err := root.OpenFile(name, os.O_RDONLY|nonBlockingFileFlag, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err = file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("runtime file must be a regular file")
	}
	if info.Size() > maxFileSize {
		return nil, errors.New("runtime file exceeds 1 MiB")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxFileSize {
		return nil, errors.New("runtime file exceeds 1 MiB")
	}
	return data, nil
}
