package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/file"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/logs"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/tls"
	"gopkg.in/fsnotify.v1"
)

const providerName = "file"

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	Directory                 string `description:"Load dynamic configuration from one or more .yml or .toml files in a directory." json:"directory,omitempty" toml:"directory,omitempty" yaml:"directory,omitempty" export:"true"`
	Watch                     bool   `description:"Watch provider." json:"watch,omitempty" toml:"watch,omitempty" yaml:"watch,omitempty" export:"true"`
	Filename                  string `description:"Load dynamic configuration from a file." json:"filename,omitempty" toml:"filename,omitempty" yaml:"filename,omitempty" export:"true"`
	DebugLogGeneratedTemplate bool   `description:"Enable debug logging of generated configuration template." json:"debugLogGeneratedTemplate,omitempty" toml:"debugLogGeneratedTemplate,omitempty" yaml:"debugLogGeneratedTemplate,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Watch = true
	p.Filename = ""
}

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Provide allows the file provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
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
			return errors.New("error using file configuration provider, neither filename or directory defined")
		}

		if err := p.addWatcher(pool, watchItem, configurationChan, p.watcherCallback); err != nil {
			return err
		}
	}

	sendConfigToChannel(configurationChan, configuration)
	return nil
}

// BuildConfiguration loads configuration either from file or a directory
// specified by 'Filename'/'Directory' and returns a 'Configuration' object.
func (p *Provider) BuildConfiguration() (*dynamic.Configuration, error) {
	ctx := log.With().Str(logs.ProviderName, providerName).Logger().WithContext(context.Background())

	if len(p.Directory) > 0 {
		return p.loadFileConfigFromDirectory(ctx, p.Directory, nil)
	}

	if len(p.Filename) > 0 {
		return p.loadFileConfig(ctx, p.Filename, true)
	}

	return nil, errors.New("error using file configuration provider, neither filename or directory defined")
}

func (p *Provider) addWatcher(pool *safe.Pool, directory string, configurationChan chan<- dynamic.Message, callback func(chan<- dynamic.Message, fsnotify.Event)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating file watcher: %w", err)
	}

	err = watcher.Add(directory)
	if err != nil {
		return fmt.Errorf("error adding file watcher: %w", err)
	}

	// Process events
	pool.GoCtx(func(ctx context.Context) {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
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
				log.Error().Str(logs.ProviderName, providerName).Err(err).Msg("Watcher event error")
			}
		}
	})
	return nil
}

func (p *Provider) watcherCallback(configurationChan chan<- dynamic.Message, event fsnotify.Event) {
	watchItem := p.Filename
	if len(p.Directory) > 0 {
		watchItem = p.Directory
	}

	logger := log.With().Str(logs.ProviderName, providerName).Logger()

	if _, err := os.Stat(watchItem); err != nil {
		logger.Error().Err(err).Msgf("Unable to watch %s", watchItem)
		return
	}

	configuration, err := p.BuildConfiguration()
	if err != nil {
		logger.Error().Err(err).Msg("Error occurred during watcher callback")
		return
	}

	sendConfigToChannel(configurationChan, configuration)
}

func sendConfigToChannel(configurationChan chan<- dynamic.Message, configuration *dynamic.Configuration) {
	configurationChan <- dynamic.Message{
		ProviderName:  "file",
		Configuration: configuration,
	}
}

func (p *Provider) loadFileConfig(ctx context.Context, filename string, parseTemplate bool) (*dynamic.Configuration, error) {
	var err error
	var configuration *dynamic.Configuration
	if parseTemplate {
		configuration, err = p.CreateConfiguration(ctx, filename, template.FuncMap{}, false)
	} else {
		configuration, err = p.DecodeConfiguration(filename)
	}
	if err != nil {
		return nil, err
	}

	if configuration.TLS != nil {
		configuration.TLS.Certificates = flattenCertificates(ctx, configuration.TLS)

		// TLS Options
		if configuration.TLS.Options != nil {
			for name, options := range configuration.TLS.Options {
				var caCerts []tls.FileOrContent

				for _, caFile := range options.ClientAuth.CAFiles {
					content, err := caFile.Read()
					if err != nil {
						log.Ctx(ctx).Error().Err(err).Send()
						continue
					}

					caCerts = append(caCerts, tls.FileOrContent(content))
				}
				options.ClientAuth.CAFiles = caCerts

				configuration.TLS.Options[name] = options
			}
		}

		// TLS stores
		if len(configuration.TLS.Stores) > 0 {
			for name, store := range configuration.TLS.Stores {
				if store.DefaultCertificate == nil {
					continue
				}

				content, err := store.DefaultCertificate.CertFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				store.DefaultCertificate.CertFile = tls.FileOrContent(content)

				content, err = store.DefaultCertificate.KeyFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				store.DefaultCertificate.KeyFile = tls.FileOrContent(content)

				configuration.TLS.Stores[name] = store
			}
		}
	}

	// ServersTransport
	if configuration.HTTP != nil && len(configuration.HTTP.ServersTransports) > 0 {
		for name, st := range configuration.HTTP.ServersTransports {
			var certificates []tls.Certificate
			for _, cert := range st.Certificates {
				content, err := cert.CertFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				cert.CertFile = tls.FileOrContent(content)

				content, err = cert.KeyFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				cert.KeyFile = tls.FileOrContent(content)

				certificates = append(certificates, cert)
			}

			configuration.HTTP.ServersTransports[name].Certificates = certificates

			var rootCAs []tls.FileOrContent
			for _, rootCA := range st.RootCAs {
				content, err := rootCA.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}

				rootCAs = append(rootCAs, tls.FileOrContent(content))
			}

			st.RootCAs = rootCAs
		}
	}

	return configuration, nil
}

func flattenCertificates(ctx context.Context, tlsConfig *dynamic.TLSConfiguration) []*tls.CertAndStores {
	var certs []*tls.CertAndStores
	for _, cert := range tlsConfig.Certificates {
		content, err := cert.Certificate.CertFile.Read()
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Send()
			continue
		}
		cert.Certificate.CertFile = tls.FileOrContent(string(content))

		content, err = cert.Certificate.KeyFile.Read()
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Send()
			continue
		}
		cert.Certificate.KeyFile = tls.FileOrContent(string(content))

		certs = append(certs, cert)
	}

	return certs
}

func (p *Provider) loadFileConfigFromDirectory(ctx context.Context, directory string, configuration *dynamic.Configuration) (*dynamic.Configuration, error) {
	fileList, err := os.ReadDir(directory)
	if err != nil {
		return configuration, fmt.Errorf("unable to read directory %s: %w", directory, err)
	}

	if configuration == nil {
		configuration = &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Routers:           make(map[string]*dynamic.Router),
				Middlewares:       make(map[string]*dynamic.Middleware),
				Services:          make(map[string]*dynamic.Service),
				ServersTransports: make(map[string]*dynamic.ServersTransport),
			},
			TCP: &dynamic.TCPConfiguration{
				Routers:     make(map[string]*dynamic.TCPRouter),
				Services:    make(map[string]*dynamic.TCPService),
				Middlewares: make(map[string]*dynamic.TCPMiddleware),
			},
			TLS: &dynamic.TLSConfiguration{
				Stores:  make(map[string]tls.Store),
				Options: make(map[string]tls.Options),
			},
			UDP: &dynamic.UDPConfiguration{
				Routers:  make(map[string]*dynamic.UDPRouter),
				Services: make(map[string]*dynamic.UDPService),
			},
		}
	}

	configTLSMaps := make(map[*tls.CertAndStores]struct{})

	for _, item := range fileList {
		logger := log.Ctx(ctx).With().Str("filename", item.Name()).Logger()

		if item.IsDir() {
			configuration, err = p.loadFileConfigFromDirectory(logger.WithContext(ctx), filepath.Join(directory, item.Name()), configuration)
			if err != nil {
				return configuration, fmt.Errorf("unable to load content configuration from subdirectory %s: %w", item, err)
			}
			continue
		}

		switch strings.ToLower(filepath.Ext(item.Name())) {
		case ".toml", ".yaml", ".yml":
			// noop
		default:
			continue
		}

		var c *dynamic.Configuration
		c, err = p.loadFileConfig(logger.WithContext(ctx), filepath.Join(directory, item.Name()), true)
		if err != nil {
			return configuration, fmt.Errorf("%s: %w", filepath.Join(directory, item.Name()), err)
		}

		for name, conf := range c.HTTP.Routers {
			if _, exists := configuration.HTTP.Routers[name]; exists {
				logger.Warn().Str(logs.RouterName, name).Msg("HTTP router already configured, skipping")
			} else {
				configuration.HTTP.Routers[name] = conf
			}
		}

		for name, conf := range c.HTTP.Middlewares {
			if _, exists := configuration.HTTP.Middlewares[name]; exists {
				logger.Warn().Str(logs.MiddlewareName, name).Msg("HTTP middleware already configured, skipping")
			} else {
				configuration.HTTP.Middlewares[name] = conf
			}
		}

		for name, conf := range c.HTTP.Services {
			if _, exists := configuration.HTTP.Services[name]; exists {
				logger.Warn().Str(logs.ServiceName, name).Msg("HTTP service already configured, skipping")
			} else {
				configuration.HTTP.Services[name] = conf
			}
		}

		for name, conf := range c.HTTP.ServersTransports {
			if _, exists := configuration.HTTP.ServersTransports[name]; exists {
				logger.Warn().Str(logs.ServersTransportName, name).Msg("HTTP servers transport already configured, skipping")
			} else {
				configuration.HTTP.ServersTransports[name] = conf
			}
		}

		for name, conf := range c.TCP.Routers {
			if _, exists := configuration.TCP.Routers[name]; exists {
				logger.Warn().Str(logs.RouterName, name).Msg("TCP router already configured, skipping")
			} else {
				configuration.TCP.Routers[name] = conf
			}
		}

		for name, conf := range c.TCP.Middlewares {
			if _, exists := configuration.TCP.Middlewares[name]; exists {
				logger.Warn().Str(logs.MiddlewareName, name).Msg("TCP middleware already configured, skipping")
			} else {
				configuration.TCP.Middlewares[name] = conf
			}
		}

		for name, conf := range c.TCP.Services {
			if _, exists := configuration.TCP.Services[name]; exists {
				logger.Warn().Str(logs.ServiceName, name).Msg("TCP service already configured, skipping")
			} else {
				configuration.TCP.Services[name] = conf
			}
		}

		for name, conf := range c.UDP.Routers {
			if _, exists := configuration.UDP.Routers[name]; exists {
				logger.Warn().Str(logs.RouterName, name).Msg("UDP router already configured, skipping")
			} else {
				configuration.UDP.Routers[name] = conf
			}
		}

		for name, conf := range c.UDP.Services {
			if _, exists := configuration.UDP.Services[name]; exists {
				logger.Warn().Str(logs.ServiceName, name).Msg("UDP service already configured, skipping")
			} else {
				configuration.UDP.Services[name] = conf
			}
		}

		for _, conf := range c.TLS.Certificates {
			if _, exists := configTLSMaps[conf]; exists {
				logger.Warn().Msgf("TLS configuration %v already configured, skipping", conf)
			} else {
				configTLSMaps[conf] = struct{}{}
			}
		}

		for name, conf := range c.TLS.Options {
			if _, exists := configuration.TLS.Options[name]; exists {
				logger.Warn().Msgf("TLS options %v already configured, skipping", name)
			} else {
				if configuration.TLS.Options == nil {
					configuration.TLS.Options = map[string]tls.Options{}
				}
				configuration.TLS.Options[name] = conf
			}
		}

		for name, conf := range c.TLS.Stores {
			if _, exists := configuration.TLS.Stores[name]; exists {
				logger.Warn().Msgf("TLS store %v already configured, skipping", name)
			} else {
				if configuration.TLS.Stores == nil {
					configuration.TLS.Stores = map[string]tls.Store{}
				}
				configuration.TLS.Stores[name] = conf
			}
		}
	}

	if len(configTLSMaps) > 0 && configuration.TLS == nil {
		configuration.TLS = &dynamic.TLSConfiguration{}
	}

	for conf := range configTLSMaps {
		configuration.TLS.Certificates = append(configuration.TLS.Certificates, conf)
	}

	return configuration, nil
}

// CreateConfiguration creates a provider configuration from content using templating.
func (p *Provider) CreateConfiguration(ctx context.Context, filename string, funcMap template.FuncMap, templateObjects interface{}) (*dynamic.Configuration, error) {
	tmplContent, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s - %w", filename, err)
	}

	defaultFuncMap := sprig.TxtFuncMap()
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

	renderedTemplate := buffer.String()
	if p.DebugLogGeneratedTemplate {
		logger := log.Ctx(ctx)
		logger.Debug().Msgf("Template content: %s", tmplContent)
		logger.Debug().Msgf("Rendering results: %s", renderedTemplate)
	}

	return p.decodeConfiguration(filename, renderedTemplate)
}

// DecodeConfiguration Decodes a *types.Configuration from a content.
func (p *Provider) DecodeConfiguration(filename string) (*dynamic.Configuration, error) {
	content, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s - %w", filename, err)
	}

	return p.decodeConfiguration(filename, content)
}

func (p *Provider) decodeConfiguration(filePath, content string) (*dynamic.Configuration, error) {
	configuration := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:     make(map[string]*dynamic.TCPRouter),
			Services:    make(map[string]*dynamic.TCPService),
			Middlewares: make(map[string]*dynamic.TCPMiddleware),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
	}

	err := file.DecodeContent(content, strings.ToLower(filepath.Ext(filePath)), configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

func readFile(filename string) (string, error) {
	if len(filename) > 0 {
		buf, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	return "", fmt.Errorf("invalid filename: %s", filename)
}
