package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/provider"
	"github.com/containous/traefik/pkg/safe"
	"github.com/containous/traefik/pkg/tls"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/yaml.v2"
)

const providerName = "file"

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	Directory                 string `description:"Load configuration from one or more .toml files in a directory." json:"directory,omitempty" toml:"directory,omitempty" yaml:"directory,omitempty" export:"true"`
	Watch                     bool   `description:"Watch provider." json:"watch,omitempty" toml:"watch,omitempty" yaml:"watch,omitempty" export:"true"`
	Filename                  string `description:"Override default configuration template. For advanced users :)" json:"filename,omitempty" toml:"filename,omitempty" yaml:"filename,omitempty" export:"true"`
	DebugLogGeneratedTemplate bool   `description:"Enable debug logging of generated configuration template." json:"debugLogGeneratedTemplate,omitempty" toml:"debugLogGeneratedTemplate,omitempty" yaml:"debugLogGeneratedTemplate,omitempty" export:"true"`
	TraefikFile               string `description:"-" json:"traefikFile,omitempty" toml:"-" yaml:"-"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Watch = true
	p.Filename = ""
}

// Init the provider
func (p *Provider) Init() error {
	return nil
}

// Provide allows the file provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- config.Message, pool *safe.Pool) error {
	configuration, err := p.BuildConfiguration()

	if err != nil {
		return err
	}

	if p.Watch {
		var watchItem string

		switch {
		case len(p.Directory) > 0:
			watchItem = p.Directory
		case len(p.Filename) > 0:
			watchItem = filepath.Dir(p.Filename)
		default:
			watchItem = filepath.Dir(p.TraefikFile)
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
func (p *Provider) BuildConfiguration() (*config.Configuration, error) {
	ctx := log.With(context.Background(), log.Str(log.ProviderName, providerName))

	if len(p.Directory) > 0 {
		return p.loadFileConfigFromDirectory(ctx, p.Directory, nil)
	}

	if len(p.Filename) > 0 {
		return p.loadFileConfig(p.Filename, true)
	}

	if len(p.TraefikFile) > 0 {
		return p.loadFileConfig(p.TraefikFile, false)
	}

	return nil, errors.New("error using file configuration backend, no filename defined")
}

func (p *Provider) addWatcher(pool *safe.Pool, directory string, configurationChan chan<- config.Message, callback func(chan<- config.Message, fsnotify.Event)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating file watcher: %s", err)
	}

	err = watcher.Add(directory)
	if err != nil {
		return fmt.Errorf("error adding file watcher: %s", err)
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
					var filename string
					if len(p.Filename) > 0 {
						filename = p.Filename
					} else {
						filename = p.TraefikFile
					}

					_, evtFileName := filepath.Split(evt.Name)
					_, confFileName := filepath.Split(filename)
					if evtFileName == confFileName {
						callback(configurationChan, evt)
					}
				} else {
					callback(configurationChan, evt)
				}
			case err := <-watcher.Errors:
				log.WithoutContext().WithField(log.ProviderName, providerName).Errorf("Watcher event error: %s", err)
			}
		}
	})
	return nil
}

func (p *Provider) watcherCallback(configurationChan chan<- config.Message, event fsnotify.Event) {
	watchItem := p.TraefikFile
	if len(p.Directory) > 0 {
		watchItem = p.Directory
	} else if len(p.Filename) > 0 {
		watchItem = p.Filename
	}

	logger := log.WithoutContext().WithField(log.ProviderName, providerName)

	if _, err := os.Stat(watchItem); err != nil {
		logger.Errorf("Unable to watch %s : %v", watchItem, err)
		return
	}

	configuration, err := p.BuildConfiguration()
	if err != nil {
		logger.Errorf("Error occurred during watcher callback: %s", err)
		return
	}

	sendConfigToChannel(configurationChan, configuration)
}

func sendConfigToChannel(configurationChan chan<- config.Message, configuration *config.Configuration) {
	configurationChan <- config.Message{
		ProviderName:  "file",
		Configuration: configuration,
	}
}

func (p *Provider) loadFileConfig(filename string, parseTemplate bool) (*config.Configuration, error) {
	var err error
	var configuration *config.Configuration
	if parseTemplate {
		configuration, err = p.CreateConfiguration(filename, template.FuncMap{}, false)
	} else {
		configuration, err = p.DecodeConfiguration(filename)
	}
	if err != nil {
		return nil, err
	}

	if configuration.TLS != nil {
		configuration.TLS.Certificates = flattenCertificates(configuration.TLS)
	}

	return configuration, nil
}

func flattenCertificates(tlsConfig *config.TLSConfiguration) []*tls.CertAndStores {
	var certs []*tls.CertAndStores
	for _, cert := range tlsConfig.Certificates {
		content, err := cert.Certificate.CertFile.Read()
		if err != nil {
			log.Error(err)
			continue
		}
		cert.Certificate.CertFile = tls.FileOrContent(string(content))

		content, err = cert.Certificate.KeyFile.Read()
		if err != nil {
			log.Error(err)
			continue
		}
		cert.Certificate.KeyFile = tls.FileOrContent(string(content))

		certs = append(certs, cert)
	}

	return certs
}

func (p *Provider) loadFileConfigFromDirectory(ctx context.Context, directory string, configuration *config.Configuration) (*config.Configuration, error) {
	logger := log.FromContext(ctx)

	fileList, err := ioutil.ReadDir(directory)
	if err != nil {
		return configuration, fmt.Errorf("unable to read directory %s: %v", directory, err)
	}

	if configuration == nil {
		configuration = &config.Configuration{
			HTTP: &config.HTTPConfiguration{
				Routers:     make(map[string]*config.Router),
				Middlewares: make(map[string]*config.Middleware),
				Services:    make(map[string]*config.Service),
			},
			TCP: &config.TCPConfiguration{
				Routers:  make(map[string]*config.TCPRouter),
				Services: make(map[string]*config.TCPService),
			},
			TLS: &config.TLSConfiguration{
				Stores:  make(map[string]tls.Store),
				Options: make(map[string]tls.Options),
			},
		}
	}

	configTLSMaps := make(map[*tls.CertAndStores]struct{})

	for _, item := range fileList {
		if item.IsDir() {
			configuration, err = p.loadFileConfigFromDirectory(ctx, filepath.Join(directory, item.Name()), configuration)
			if err != nil {
				return configuration, fmt.Errorf("unable to load content configuration from subdirectory %s: %v", item, err)
			}
			continue
		}

		switch strings.ToLower(filepath.Ext(item.Name())) {
		case ".toml", ".yaml", ".yml":
			// noop
		default:
			continue
		}

		var c *config.Configuration
		c, err = p.loadFileConfig(filepath.Join(directory, item.Name()), true)
		if err != nil {
			return configuration, err
		}

		for name, conf := range c.HTTP.Routers {
			if _, exists := configuration.HTTP.Routers[name]; exists {
				logger.WithField(log.RouterName, name).Warn("HTTP router already configured, skipping")
			} else {
				configuration.HTTP.Routers[name] = conf
			}
		}

		for name, conf := range c.HTTP.Middlewares {
			if _, exists := configuration.HTTP.Middlewares[name]; exists {
				logger.WithField(log.MiddlewareName, name).Warn("HTTP middleware already configured, skipping")
			} else {
				configuration.HTTP.Middlewares[name] = conf
			}
		}

		for name, conf := range c.HTTP.Services {
			if _, exists := configuration.HTTP.Services[name]; exists {
				logger.WithField(log.ServiceName, name).Warn("HTTP service already configured, skipping")
			} else {
				configuration.HTTP.Services[name] = conf
			}
		}

		for name, conf := range c.TCP.Routers {
			if _, exists := configuration.TCP.Routers[name]; exists {
				logger.WithField(log.RouterName, name).Warn("TCP router already configured, skipping")
			} else {
				configuration.TCP.Routers[name] = conf
			}
		}

		for name, conf := range c.TCP.Services {
			if _, exists := configuration.TCP.Services[name]; exists {
				logger.WithField(log.ServiceName, name).Warn("TCP service already configured, skipping")
			} else {
				configuration.TCP.Services[name] = conf
			}
		}

		for _, conf := range c.TLS.Certificates {
			if _, exists := configTLSMaps[conf]; exists {
				logger.Warnf("TLS configuration %v already configured, skipping", conf)
			} else {
				configTLSMaps[conf] = struct{}{}
			}
		}
	}

	if len(configTLSMaps) > 0 {
		configuration.TLS = &config.TLSConfiguration{}
	}

	for conf := range configTLSMaps {
		configuration.TLS.Certificates = append(configuration.TLS.Certificates, conf)
	}

	return configuration, nil
}

// CreateConfiguration creates a provider configuration from content using templating.
func (p *Provider) CreateConfiguration(filename string, funcMap template.FuncMap, templateObjects interface{}) (*config.Configuration, error) {
	tmplContent, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s - %s", filename, err)
	}

	var defaultFuncMap = sprig.TxtFuncMap()
	defaultFuncMap["normalize"] = provider.Normalize
	defaultFuncMap["split"] = strings.Split
	for funcID, funcElement := range funcMap {
		defaultFuncMap[funcID] = funcElement
	}

	tmpl := template.New(p.Filename).Funcs(defaultFuncMap)

	_, err = tmpl.Parse(tmplContent)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		return nil, err
	}

	var renderedTemplate = buffer.String()
	if p.DebugLogGeneratedTemplate {
		logger := log.WithoutContext().WithField(log.ProviderName, providerName)
		logger.Debugf("Template content: %s", tmplContent)
		logger.Debugf("Rendering results: %s", renderedTemplate)
	}

	return p.decodeConfiguration(filename, renderedTemplate)
}

// DecodeConfiguration Decodes a *types.Configuration from a content.
func (p *Provider) DecodeConfiguration(filename string) (*config.Configuration, error) {
	content, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s - %s", filename, err)
	}

	return p.decodeConfiguration(filename, content)
}

func (p *Provider) decodeConfiguration(filePath string, content string) (*config.Configuration, error) {
	configuration := &config.Configuration{
		HTTP: &config.HTTPConfiguration{
			Routers:     make(map[string]*config.Router),
			Middlewares: make(map[string]*config.Middleware),
			Services:    make(map[string]*config.Service),
		},
		TCP: &config.TCPConfiguration{
			Routers:  make(map[string]*config.TCPRouter),
			Services: make(map[string]*config.TCPService),
		},
		TLS: &config.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
	}

	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".toml":
		_, err := toml.Decode(content, configuration)
		if err != nil {
			return nil, err
		}

	case ".yml", ".yaml":
		var err error
		err = yaml.Unmarshal([]byte(content), configuration)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported file extension: %s", filePath)
	}

	return configuration, nil
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
