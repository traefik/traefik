package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"gopkg.in/fsnotify.v1"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`
	Directory             string `description:"Load configuration from one or more .toml files in a directory" export:"true"`
}

// Provide allows the file provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	configuration, err := p.BuildConfiguration()

	if err != nil {
		return err
	}

	if p.Watch {
		var watchItem string

		if p.Directory != "" {
			watchItem = p.Directory
		} else {
			watchItem = filepath.Dir(p.Filename)
		}

		if err := p.addWatcher(pool, watchItem, configurationChan, p.watcherCallback); err != nil {
			return err
		}
	}

	sendConfigToChannel(configurationChan, configuration)
	return nil
}

// BuildConfiguration loads configuration either from file or a directory specified by 'Filename'/'Directory'
// and returns a 'Configuration' object
func (p *Provider) BuildConfiguration() (*types.Configuration, error) {
	if p.Directory != "" {
		return p.loadFileConfigFromDirectory(p.Directory, nil)
	}
	return p.loadFileConfig(p.Filename)
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
				if p.Directory == "" {
					_, evtFileName := filepath.Split(evt.Name)
					_, confFileName := filepath.Split(p.Filename)
					if evtFileName == confFileName {
						callback(configurationChan, evt)
					}
				} else {
					callback(configurationChan, evt)
				}
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

func (p *Provider) watcherCallback(configurationChan chan<- types.ConfigMessage, event fsnotify.Event) {
	watchItem := p.Filename
	if p.Directory != "" {
		watchItem = p.Directory
	}

	if _, err := os.Stat(watchItem); err != nil {
		log.Debugf("Unable to watch %s : %v", watchItem, err)
		return
	}

	configuration, err := p.BuildConfiguration()

	if err != nil {
		log.Errorf("Error occurred during watcher callback: %s", err)
		return
	}

	sendConfigToChannel(configurationChan, configuration)
}

func sendConfigToChannel(configurationChan chan<- types.ConfigMessage, configuration *types.Configuration) {
	configurationChan <- types.ConfigMessage{
		ProviderName:  "file",
		Configuration: configuration,
	}
}

func readFile(filename string) (string, error) {
	if len(filename) > 0 {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	return "", fmt.Errorf("invalid filename: %s", filename)
}

func (p *Provider) loadFileConfig(filename string) (*types.Configuration, error) {
	fileContent, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s - %s", filename, err)
	}
	return p.GetConfiguration(fileContent, template.FuncMap{}, false)
}

func (p *Provider) loadFileConfigFromDirectory(directory string, configuration *types.Configuration) (*types.Configuration, error) {
	fileList, err := ioutil.ReadDir(directory)

	if err != nil {
		return configuration, fmt.Errorf("unable to read directory %s: %v", directory, err)
	}

	if configuration == nil {
		configuration = &types.Configuration{
			Frontends: make(map[string]*types.Frontend),
			Backends:  make(map[string]*types.Backend),
		}
	}

	configTLSMaps := make(map[*tls.Configuration]struct{})
	for _, item := range fileList {

		if item.IsDir() {
			configuration, err = p.loadFileConfigFromDirectory(filepath.Join(directory, item.Name()), configuration)
			if err != nil {
				return configuration, fmt.Errorf("unable to load content configuration from subdirectory %s: %v", item, err)
			}
			continue
		} else if !strings.HasSuffix(item.Name(), ".toml") {
			continue
		}

		var c *types.Configuration
		c, err = p.loadFileConfig(path.Join(directory, item.Name()))

		if err != nil {
			return configuration, err
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

		for _, conf := range c.TLS {
			if _, exists := configTLSMaps[conf]; exists {
				log.Warnf("TLS Configuration %v already configured, skipping", conf)
			} else {
				configTLSMaps[conf] = struct{}{}
			}
		}

	}
	for conf := range configTLSMaps {
		configuration.TLS = append(configuration.TLS, conf)
	}
	return configuration, nil
}
