package main

import (
	"log"
	"gopkg.in/fsnotify.v1"
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"strings"
)

type FileProvider struct {
	Watch bool
	Filename string
}

func (provider *FileProvider) Provide(configurationChan chan<- *Configuration){
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return
	}
	defer watcher.Close()

	file, err := os.Open(provider.Filename)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	done := make(chan bool)
	// Process events
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if(strings.Contains(event.Name,file.Name())){
					log.Println("File event:", event)
					configuration := provider.LoadFileConfig(file.Name())
					if(configuration != nil) {
						configurationChan <- configuration
					}
				}
			case error := <-watcher.Errors:
				log.Println("error:", error)
			}
		}
	}()

	if(provider.Watch){
		err = watcher.Add(filepath.Dir(file.Name()))
	}

	if err != nil {
		log.Println(err)
		return
	}


	configuration := provider.LoadFileConfig(file.Name())
	configurationChan <- configuration
	<-done
}


func (provider *FileProvider) LoadFileConfig(filename string) *Configuration {
	configuration := new(Configuration)
	if _, err := toml.DecodeFile(filename, configuration); err != nil {
		log.Println("Error reading file:", err)
		return nil
	}
	return configuration
}