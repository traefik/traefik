package main

import (
	"log"
	"gopkg.in/fsnotify.v1"
	"github.com/BurntSushi/toml"
)

type FileProvider struct {
}

func (provider *FileProvider) Provide(serviceChan chan<- *Service){
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()


	err = watcher.Add(".")
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	// Process events
	go func() {
		for {
			select {
			case event := <-watcher.Events:
			if(event.Name == "./tortuous.toml"){
				log.Println("event:", event)
				service := provider.LoadFileConfig()
				serviceChan <- service
			}
			case error := <-watcher.Errors:
				log.Println("error:", error)
			}
		}
	}()


	service:= provider.LoadFileConfig()
	serviceChan <- service
	<-done
	log.Println("DONE")
}


func (provider *FileProvider) LoadFileConfig() *Service  {
	service := new(Service)
	if _, err := toml.DecodeFile("tortuous.toml", service); err != nil {
		log.Println(err)
		return nil
	}
	return service
}