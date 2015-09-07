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
	Filename string
}

func (provider *FileProvider) Provide(serviceChan chan<- *Service){
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
					log.Println("event:", event)
					service := provider.LoadFileConfig(file.Name())
					serviceChan <- service
				}
			case error := <-watcher.Errors:
				log.Println("error:", error)
			}
		}
	}()

	err = watcher.Add(filepath.Dir(file.Name()))

	if err != nil {
		log.Println(err)
		return
	}


	service:= provider.LoadFileConfig(file.Name())
	serviceChan <- service
	<-done
}


func (provider *FileProvider) LoadFileConfig(filename string) *Service  {
	service := new(Service)
	if _, err := toml.DecodeFile(filename, service); err != nil {
		log.Println("Error reading file:", err)
		return nil
	}
	return service
}