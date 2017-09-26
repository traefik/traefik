package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"gopkg.in/fsnotify.v1"
)

var _ provider.Provider = (*Provider)(nil)

// ProviderHTTPSFile is the HTTPS file provider name
const ProviderHTTPSFile = "httpsFile"

// Provider holds configurations of the provider.
type Provider struct {
	provider.RootProvider `mapstructure:",squash"`
	ConfigurationFile     string `description:"TOML file with all certificates to manage"`
}

// Provide allows the HTTPS file provider to provide SSL configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	if p.Watch {
		if p.ConfigurationFile == "" {
			return errors.New("none file to watch provided")
		}

		if err := p.addWatcher(pool, p.ConfigurationFile, configurationChan, p.watcherCallback); err != nil {
			return err
		}
	}
	configuration, err := p.loadConfig()

	if err != nil {
		return err
	}
	sendHTTPSConfigToChannel(configurationChan, configuration)
	return nil
}

func (p *Provider) addWatcher(pool *safe.Pool, configurationFilePath string, configurationChan chan<- types.ConfigMessage, callback func(chan<- types.ConfigMessage, fsnotify.Event)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating HTTPS file watcher: %s", err)
	}

	// Process events
	pool.Go(func(stop chan bool) {
		defer watcher.Close()
		for {
			select {
			case <-stop:
				return
			case evt := <-watcher.Events:
				_, evtFileName := filepath.Split(evt.Name)
				_, confFileName := filepath.Split(configurationFilePath)
				if evtFileName == confFileName {
					callback(configurationChan, evt)
				}
			case err := <-watcher.Errors:
				log.Errorf("HTTPS file watcher event error: %s", err)
			}
		}
	})
	configurationDirPath := filepath.Dir(configurationFilePath)
	err = watcher.Add(configurationDirPath)
	if err != nil {
		return fmt.Errorf("error adding HTTPS file watcher: %s", err)
	}

	return nil
}

func sendHTTPSConfigToChannel(configurationChan chan<- types.ConfigMessage, configuration *tls.EntrypointsCertificates) {
	configurationChan <- types.ConfigMessage{
		ProviderName:     ProviderHTTPSFile,
		TLSConfiguration: configuration,
	}
}

func (p *Provider) watcherCallback(configurationChan chan<- types.ConfigMessage, event fsnotify.Event) {
	if _, err := os.Stat(p.ConfigurationFile); err != nil {
		log.Debugf("Impossible to watch file %s : %s", p.ConfigurationFile, err.Error())
		return
	}
	configuration, err := p.loadConfig()
	if err != nil {
		log.Errorf("Error occurred during watcher callback: %s", err)
		return
	}
	sendHTTPSConfigToChannel(configurationChan, configuration)
}

func (p *Provider) loadConfig() (*tls.EntrypointsCertificates, error) {

	configFromFile := new(tls.DynamicConfigurations)

	if _, err := toml.DecodeFile(p.ConfigurationFile, configFromFile); err != nil {
		return nil, fmt.Errorf("error reading HTTPS configuration file: %s", err)
	}

	if len(configFromFile.TLS) == 0 {
		log.Debugf("No certificate defined in the HTTPS configuration file.")
		return nil, nil
	}
	return configFromFile.ConvertTLSDynamicsToTLSConfiguration(), nil
}
