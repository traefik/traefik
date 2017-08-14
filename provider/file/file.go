package file

import (
	"fmt"
	"io/ioutil"
	"path"
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
	Directory             string `description:"Load configuration from one or more .toml files in a directory"`
}

// Provide allows the file provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	configuration, err := p.loadConfig()

	if err != nil {
		return err
	}

	if p.Watch {
		var watchItem string

		if p.Directory != "" {
			watchItem = p.Directory
		} else {
			watchItem = p.Filename
		}

		if err := p.addWatcher(pool, watchItem, configurationChan, p.watcherCallback); err != nil {
			return err
		}
	}

	sendConfigToChannel(configurationChan, configuration)
	return nil
}

func (p *Provider) addWatcher(pool *safe.Pool, directory string, configurationChan chan<- types.ConfigMessage, callback func(chan<- types.ConfigMessage, fsnotify.Event)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating file watcher: %s", err)
	}

	// Process events
	pool.Go(func(stop chan bool) {
		defer watcher.Close()
		for {
			select {
			case <-stop:
				return
			case evt := <-watcher.Events:
				callback(configurationChan, evt)
			case err := <-watcher.Errors:
				log.Errorf("Watcher event error: %s", err)
			}
		}
	})
	err = watcher.Add(directory)
	if err != nil {
		return fmt.Errorf("error adding file watcher: %s", err)
	}

	return nil
}

func sendConfigToChannel(configurationChan chan<- types.ConfigMessage, configuration *types.Configuration) {
	configurationChan <- types.ConfigMessage{
		ProviderName:  "file",
		Configuration: configuration,
	}
}

func loadFileConfig(filename string) (*types.Configuration, error) {
	configuration := new(types.Configuration)
	if _, err := toml.DecodeFile(filename, configuration); err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s", err)
	}
	return configuration, nil
}

func loadFileConfigFromDirectory(directory string) (*types.Configuration, error) {
	fileList, err := ioutil.ReadDir(directory)

	if err != nil {
		return nil, fmt.Errorf("unable to read directory %s: %v", directory, err)
	}

	configuration := &types.Configuration{
		Frontends: make(map[string]*types.Frontend),
		Backends:  make(map[string]*types.Backend),
	}

	for _, file := range fileList {
		if !strings.HasSuffix(file.Name(), ".toml") {
			continue
		}

		var c *types.Configuration
		c, err = loadFileConfig(path.Join(directory, file.Name()))

		if err != nil {
			return nil, err
		}

		for backendName, backend := range c.Backends {
			if _, exists := configuration.Backends[backendName]; exists {
				log.Warnf("Backend %s already configured, skipping", backendName)
			} else {
				configuration.Backends[backendName] = backend
			}
		}

		for frontendName, frontend := range c.Frontends {
			if _, exists := configuration.Frontends[frontendName]; exists {
				log.Warnf("Frontend %s already configured, skipping", frontendName)
			} else {
				configuration.Frontends[frontendName] = frontend
			}
		}
	}

	return configuration, nil
}

func (p *Provider) watcherCallback(configurationChan chan<- types.ConfigMessage, event fsnotify.Event) {
	configuration, err := p.loadConfig()

	if err != nil {
		log.Errorf("Error occurred during watcher callback: %s", err)
		return
	}

	sendConfigToChannel(configurationChan, configuration)
}

func (p *Provider) loadConfig() (*types.Configuration, error) {
	if p.Directory != "" {
		return loadFileConfigFromDirectory(p.Directory)
	}

	return loadFileConfig(p.Filename)
}
