package file

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"gopkg.in/fsnotify.v1"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
}

// Provide allows the file provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Error creating file watcher", err)
		return err
	}

	file, err := os.Open(p.Filename)
	if err != nil {
		log.Error("Error opening file", err)
		return err
	}
	defer file.Close()

	if p.Watch {
		// Process events
		pool.Go(func(stop chan bool) {
			defer watcher.Close()
			for {
				select {
				case <-stop:
					return
				case event := <-watcher.Events:
					if strings.Contains(event.Name, file.Name()) {
						log.Debug("Provider event:", event)
						configuration := p.loadFileConfig(file.Name())
						if configuration != nil {
							configurationChan <- types.ConfigMessage{
								ProviderName:  "file",
								Configuration: configuration,
							}
						}
					}
				case error := <-watcher.Errors:
					log.Error("Watcher event error", error)
				}
			}
		})
		err = watcher.Add(filepath.Dir(file.Name()))
		if err != nil {
			log.Error("Error adding file watcher", err)
			return err
		}
	}

	configuration := p.loadFileConfig(file.Name())
	configurationChan <- types.ConfigMessage{
		ProviderName:  "file",
		Configuration: configuration,
	}
	return nil
}

func (p *Provider) loadFileConfig(filename string) *types.Configuration {
	configuration := new(types.Configuration)
	if _, err := toml.DecodeFile(filename, configuration); err != nil {
		log.Error("Error reading file:", err)
		return nil
	}
	return configuration
}
