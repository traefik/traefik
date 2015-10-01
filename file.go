package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

type FileProvider struct {
	Watch    bool
	Filename string
}

func (provider *FileProvider) Provide(configurationChan chan<- configMessage) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Error creating file watcher", err)
		return err
	}
	defer watcher.Close()

	file, err := os.Open(provider.Filename)
	if err != nil {
		log.Error("Error opening file", err)
		return err
	}
	defer file.Close()

	// Process events
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if strings.Contains(event.Name, file.Name()) {
					log.Debug("File event:", event)
					configuration := provider.LoadFileConfig(file.Name())
					if configuration != nil {
						configurationChan <- configMessage{"file", configuration}
					}
				}
			case error := <-watcher.Errors:
				log.Error("Watcher event error", error)
			}
		}
	}()

	if provider.Watch {
		err = watcher.Add(filepath.Dir(file.Name()))
	}

	if err != nil {
		log.Error("Error adding file watcher", err)
		return err
	}

	configuration := provider.LoadFileConfig(file.Name())
	configurationChan <- configMessage{"file", configuration}
	return nil
}

func (provider *FileProvider) LoadFileConfig(filename string) *Configuration {
	configuration := new(Configuration)
	if _, err := toml.DecodeFile(filename, configuration); err != nil {
		log.Error("Error reading file:", err)
		return nil
	}
	return configuration
}
