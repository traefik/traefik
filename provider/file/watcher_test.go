package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/fsnotify.v1"
)

// createRandomFile Helper
func createRandomFile(t *testing.T, tempDir string, contents ...string) *os.File {
	return createFile(t, tempDir, fmt.Sprintf("temp%d.toml", time.Now().UnixNano()), contents...)
}

// createFile Helper
func createFile(t *testing.T, tempDir string, name string, contents ...string) *os.File {
	t.Helper()
	fileName := path.Join(tempDir, name)

	tempFile, err := os.Create(fileName)
	if err != nil {
		t.Fatal(err)
	}

	for _, content := range contents {
		_, err := tempFile.WriteString(content)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = tempFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	return tempFile
}

// createTempDir Helper
func createTempDir(t *testing.T, dir string) string {
	t.Helper()
	d, err := ioutil.TempDir("", dir)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// createDir Helper
func createDir(t *testing.T, parentDir string, name string) string {
	t.Helper()
	dirPath := path.Join(parentDir, name)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	return dirPath
}

// I want test create/remove/rename file
func TestWatchFileInWatchedDir(t *testing.T) {
	watcher, err := newWatcher()
	assert.Nil(t, err)
	defer watcher.Close()

	dirToWatch := createTempDir(t, "watcheddir")
	watcher.addRecursive(dirToWatch)

	name := "foo.toml"
	originalPath := path.Join(dirToWatch, name)
	renamedPath := path.Join(dirToWatch, "bar.toml")

	createFile(t, dirToWatch, name)
	evt := <-watcher.Events
	assert.Equal(t, evt.Name, originalPath)
	assert.Equal(t, evt.Op, fsnotify.Create)
	_, exists := watcher.directories[originalPath]
	assert.False(t, exists)

	os.Rename(originalPath, renamedPath)
	evt = <-watcher.Events
	assert.Equal(t, evt.Name, originalPath)
	assert.Equal(t, evt.Op, fsnotify.Rename)

	evt = <-watcher.Events
	assert.Equal(t, evt.Name, renamedPath)
	assert.Equal(t, evt.Op, fsnotify.Create)

	os.Remove(renamedPath)
	evt = <-watcher.Events
	assert.Equal(t, evt.Name, renamedPath)
	assert.Equal(t, evt.Op, fsnotify.Remove)
}

func TestWatchSubdirInWatchedDir(t *testing.T) {
	watcher, err := newWatcher()
	assert.Nil(t, err)
	defer watcher.Close()

	dirToWatch := createTempDir(t, "watcheddir")
	watcher.addRecursive(dirToWatch)

	renamedPath := path.Join(dirToWatch, "bar")

	originalPath := createDir(t, dirToWatch, "foo")
	evt := <-watcher.Events
	assert.Equal(t, evt.Name, originalPath)
	assert.Equal(t, evt.Op, fsnotify.Create)
	_, exists := watcher.directories[originalPath]
	assert.True(t, exists)

	err = os.Rename(originalPath, renamedPath)
	assert.Nil(t, err)

	evt = <-watcher.Events
	assert.Equal(t, evt.Name, originalPath)
	assert.Equal(t, evt.Op, fsnotify.Rename)
	_, exists = watcher.directories[originalPath]
	assert.False(t, exists)

	evt = <-watcher.Events
	assert.Equal(t, evt.Name, renamedPath)
	assert.Equal(t, evt.Op, fsnotify.Create)
	_, exists = watcher.directories[renamedPath]
	assert.True(t, exists)

	os.Remove(renamedPath)
	evt = <-watcher.Events
	assert.Equal(t, evt.Name, renamedPath)
	assert.Equal(t, evt.Op, fsnotify.Remove)
	_, exists = watcher.directories[renamedPath]
	assert.False(t, exists)
}

// I want test create subdir + create/remove/rename file in subdir

// I want test removerecursive but not delete real subdir modify its contents

// I want test rename dir

// I want test create sub-subdir + file + modify
