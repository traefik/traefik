package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"gopkg.in/fsnotify.v1"
)

type watcher struct {
	*fsnotify.Watcher
	Events      chan fsnotify.Event
	done        chan struct{}
	directories map[string]struct{}
}

func newWatcher() (*watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating file watcher: %v", err)
	}

	newWatcher := &watcher{
		Watcher:     fsWatcher,
		done:        make(chan struct{}),
		Events:      make(chan fsnotify.Event),
		directories: make(map[string]struct{}),
	}

	newWatcher.handleRecursiveEvent()

	return newWatcher, nil
}

func (w *watcher) handleRecursiveEvent() {
	safe.Go(func() {
		defer func() {
			err := w.Watcher.Close()
			if err != nil {
				log.Debugf("Unable to close watcher: %v", err)
			}
		}()

		for {
			select {
			case evt := <-w.Watcher.Events:
				if evt.Op == fsnotify.Remove || evt.Op == fsnotify.Rename {
					w.removeRecursive(evt.Name)
				} else if evt.Op == fsnotify.Create {
					w.addRecursive(evt.Name)
				}
				w.Events <- evt
			case <-w.done:
				return
			}
		}
	})
}

func (w *watcher) Close() error {
	close(w.done)
	return nil
}

func (w *watcher) removeRecursive(rootDir string) {
	for dir := range w.directories {
		if strings.HasPrefix(dir, rootDir) {
			err := w.Remove(dir)
			if err != nil {
				log.Errorf("Unable to remove %s from watcher: %v", dir, err)
				continue
			}
			delete(w.directories, dir)
		}
	}
}

func (w *watcher) addRecursive(rootDir string) error {
	directories := getDirectoriesRecursive(rootDir)

	for dir := range directories {
		err := w.Add(dir)
		if err != nil {
			log.Errorf("Unable to add file watcher on directory %q: %v", dir, err)
		} else {
			w.directories[dir] = struct{}{}
		}
	}

	return nil
}

func getDirectoriesRecursive(rootDir string) map[string]struct{} {
	rootDirInfo, err := os.Stat(rootDir)
	if err != nil {
		log.Errorf("unable to stat %q: %v", rootDir, err)
		return nil
	}

	if !rootDirInfo.IsDir() {
		return nil
	}

	directories := map[string]struct{}{
		rootDir: {},
	}

	fileList, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Errorf("unable to initialize sub-directories list from directory %s: %v", rootDir, err)
		return nil
	}

	for _, item := range fileList {
		for dir := range getDirectoriesRecursive(strings.TrimSuffix(rootDir, "/") + "/" + item.Name()) {
			directories[dir] = struct{}{}
		}
	}

	return directories
}
