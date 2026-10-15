package runtimeconfig

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadRuntimeFile(t *testing.T) {
	for _, size := range []int{0, 32, maxFileSize, maxFileSize + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "credential")
			data := []byte(strings.Repeat("x", size))
			require.NoError(t, os.WriteFile(path, data, 0o600))
			actual, err := ReadFile(path)
			if size > maxFileSize {
				require.ErrorContains(t, err, "exceeds 1 MiB")
				assert.Nil(t, actual)
			} else {
				require.NoError(t, err)
				assert.Equal(t, data, actual)
			}
		})
	}
}

func TestReadRuntimeFileRejectsInvalidPaths(t *testing.T) {
	directory := t.TempDir()
	for _, path := range []string{"", directory, filepath.Join(directory, "missing")} {
		_, err := ReadFile(path)
		require.Error(t, err)
	}
}

func TestReadRuntimeFileSymlinkBoundaries(t *testing.T) {
	directory := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "credential"), []byte("must-not-leak"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "regular"), []byte("inside"), 0o600))
	require.NoError(t, os.Symlink("regular", filepath.Join(directory, "safe")))
	data, err := ReadFile(filepath.Join(directory, "safe"))
	require.NoError(t, err)
	assert.Equal(t, "inside", string(data))
	relative, err := filepath.Rel(directory, outside)
	require.NoError(t, err)
	for name, target := range map[string]string{
		"absolute": filepath.Join(outside, "credential"),
		"relative": filepath.Join(relative, "credential"),
		"broken":   "missing",
		"loop":     "loop",
	} {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, os.Symlink(target, filepath.Join(directory, name)))
			data, err := ReadFile(filepath.Join(directory, name))
			require.Error(t, err)
			assert.Nil(t, data)
			assert.NotContains(t, err.Error(), "must-not-leak")
		})
	}
	require.NoError(t, os.Symlink(relative, filepath.Join(directory, "..data")))
	require.NoError(t, os.Symlink(filepath.Join("..data", "credential"), filepath.Join(directory, "credential")))
	_, err = ReadFile(filepath.Join(directory, "credential"))
	require.Error(t, err)
}

func TestProjectedRuntimeFilesStayOnOneRevision(t *testing.T) {
	directory := t.TempDir()
	for _, version := range []string{"one", "two"} {
		require.NoError(t, os.Mkdir(filepath.Join(directory, version), 0o700))
		for _, name := range []string{"token", "project"} {
			require.NoError(t, os.WriteFile(filepath.Join(directory, version, name), []byte(version), 0o600))
		}
	}
	for _, name := range []string{"token", "project"} {
		require.NoError(t, os.Symlink(filepath.Join("..data", name), filepath.Join(directory, name)))
	}
	require.NoError(t, os.WriteFile(filepath.Join(directory, "ordinary"), []byte("literal"), 0o600))
	require.NoError(t, os.Symlink("one", filepath.Join(directory, "..data")))
	paths := map[string]string{"token": filepath.Join(directory, "token"), "project": filepath.Join(directory, "project"), "ordinary": filepath.Join(directory, "ordinary")}
	initial, err := ReadFiles(paths)
	require.NoError(t, err)
	assert.Equal(t, "one", string(initial["token"]))
	rotate := func(version string) error {
		if err := os.Symlink(version, filepath.Join(directory, "..next")); err != nil {
			return err
		}
		return os.Rename(filepath.Join(directory, "..next"), filepath.Join(directory, "..data"))
	}
	require.NoError(t, rotate("two"))
	changed, err := ReadFiles(paths)
	require.NoError(t, err)
	assert.Equal(t, "two", string(changed["token"]))

	stop := make(chan struct{})
	errors := make(chan error, 1)
	var worker sync.WaitGroup
	worker.Go(func() {
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			if err := rotate([]string{"one", "two"}[i%2]); err != nil {
				errors <- err
				return
			}
		}
	})
	defer func() { close(stop); worker.Wait() }()
	for range 100 {
		values, err := ReadFiles(paths)
		require.NoError(t, err)
		assert.Equal(t, values["token"], values["project"])
		assert.Equal(t, "literal", string(values["ordinary"]))
	}
	select {
	case err := <-errors:
		require.NoError(t, err)
	default:
	}
}
