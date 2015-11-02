package provider

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/emilevauge/traefik/types"
	"gopkg.in/fsnotify.v1"
)

type File struct {
	Watch    bool
	Filename string
}

func (provider *File) Provide(configurationChan chan<- types.ConfigMessage) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Error creating file watcher", err)
		return err
	}

	file, err := os.Open(provider.Filename)
	if err != nil {
		log.Error("Error opening file", err)
		return err
	}
	defer file.Close()

	if provider.Watch {
		// Process events
		go func() {
			defer watcher.Close()
			for {
				select {
				case event := <-watcher.Events:
					if strings.Contains(event.Name, file.Name()) {
						log.Debug("File event:", event)
						configuration := provider.LoadFileConfig(file.Name())
						if configuration != nil {
							configurationChan <- types.ConfigMessage{"file", configuration}
						}
					}
				case error := <-watcher.Errors:
					log.Error("Watcher event error", error)
				}
			}
		}()
		err = watcher.Add(filepath.Dir(file.Name()))
		if err != nil {
			log.Error("Error adding file watcher", err)
			return err
		}
	}

	configuration := provider.LoadFileConfig(file.Name())
	configurationChan <- types.ConfigMessage{"file", configuration}
	return nil
}

func (provider *File) LoadFileConfig(filename string) *types.Configuration {
	configuration := new(types.Configuration)
	if _, err := toml.DecodeFile(filename, configuration); err != nil {
		log.Error("Error reading file:", err)
		return nil
	}
	return configuration
}
