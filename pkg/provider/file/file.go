package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/file"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

const providerName = "file"

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	Directory                 string `description:"Load dynamic configuration from one or more .yml or .toml files in a directory." json:"directory,omitempty" toml:"directory,omitempty" yaml:"directory,omitempty" export:"true"`
	Watch                     bool   `description:"Watch provider." json:"watch,omitempty" toml:"watch,omitempty" yaml:"watch,omitempty" export:"true"`
	Filename                  string `description:"Load dynamic configuration from a file." json:"filename,omitempty" toml:"filename,omitempty" yaml:"filename,omitempty" export:"true"`
	DebugLogGeneratedTemplate bool   `description:"Enable debug logging of generated configuration template." json:"debugLogGeneratedTemplate,omitempty" toml:"debugLogGeneratedTemplate,omitempty" yaml:"debugLogGeneratedTemplate,omitempty" export:"true"`

	watcherMu sync.RWMutex
	// watcher is nil when Watch is false.
	watcher *fsnotify.Watcher
	// externalDirs is the set of directories currently watched because they contain an
	// externally referenced file (certificate, key, or CA bundle).
	externalDirs map[string]struct{}
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
	logger := log.With().Str(logs.ProviderName, providerName).Logger()

	if p.Watch {
		var watchItems []string

		switch {
		case len(p.Directory) > 0:
			watchItems = append(watchItems, p.Directory)

			fileList, err := os.ReadDir(p.Directory)
			if err != nil {
				return fmt.Errorf("unable to read directory %s: %w", p.Directory, err)
			}

			for _, entry := range fileList {
				if entry.IsDir() {
					// ignore sub-dir
					continue
				}
				if !isFileSupported(entry.Name()) {
					// ignore unsupported file extension
					continue
				}
				watchItems = append(watchItems, path.Join(p.Directory, entry.Name()))
			}
		case len(p.Filename) > 0:
			if !isFileSupported(p.Filename) {
				return fmt.Errorf("unsupported file extension for file %s", p.Filename)
			}
			watchItems = append(watchItems, filepath.Dir(p.Filename), p.Filename)
		default:
			return errors.New("error using file configuration provider, neither filename nor directory is defined")
		}

		if err := p.addWatcher(pool, watchItems, configurationChan, p.applyConfiguration); err != nil {
			return err
		}
	}

	pool.GoCtx(func(ctx context.Context) {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGHUP)

		for {
			select {
			case <-ctx.Done():
				return
			// signals only receives SIGHUP events.
			case <-signals:
				if err := p.applyConfiguration(configurationChan); err != nil {
					logger.Error().Err(err).Msg("Error while building configuration")
				}
			}
		}
	})

	if err := p.applyConfiguration(configurationChan); err != nil {
		if p.Watch {
			logger.Err(err).Msg("Error while building configuration (for the first time)")
			return nil
		}

		return err
	}

	return nil
}

// CreateConfiguration creates a provider configuration from content using templating.
func (p *Provider) CreateConfiguration(ctx context.Context, filename string, funcMap template.FuncMap, templateObjects any) (*dynamic.Configuration, error) {
	tmplContent, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s - %w", filename, err)
	}

	defaultFuncMap := sprig.TxtFuncMap()
	defaultFuncMap["normalize"] = provider.Normalize
	defaultFuncMap["split"] = strings.Split
	maps.Copy(defaultFuncMap, funcMap)

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

func (p *Provider) addWatcher(pool *safe.Pool, items []string, configurationChan chan<- dynamic.Message, callback func(chan<- dynamic.Message) error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating file watcher: %w", err)
	}

	for _, item := range items {
		log.Debug().Msgf("add watcher on: %s", item)
		err = watcher.Add(item)
		if err != nil {
			return fmt.Errorf("error adding file watcher for %s: %w", item, err)
		}
	}

	p.watcherMu.Lock()
	p.watcher = watcher
	p.watcherMu.Unlock()

	// Process events
	pool.GoCtx(func(ctx context.Context) {
		logger := log.With().Str(logs.ProviderName, providerName).Logger()
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-watcher.Events:
				if !p.isRelevantEvent(evt.Name) {
					continue
				}

				if err := callback(configurationChan); err != nil {
					logger.Error().Err(err).Msg("Error occurred during watcher callback")
				}
			case err := <-watcher.Errors:
				logger.Error().Err(err).Msg("Watcher event error")
			}
		}
	})
	return nil
}

// isBaseDir reports whether dir is the directory permanently watched as part of the
// provider's own Directory/Filename item, i.e. one that syncExternalFileWatches must
// never Add or Remove: an externally referenced file happening to live alongside the
// dynamic config file(s) must not cause that permanent watch to be torn down once the
// file stops being referenced.
func (p *Provider) isBaseDir(dir string) bool {
	if p.Directory != "" && dir == p.Directory {
		return true
	}
	return p.Filename != "" && dir == filepath.Dir(p.Filename)
}

// isRelevantEvent reports whether an fsnotify event for filename should trigger a
// configuration reload: any event under the watched Directory, a change to the watched
// Filename itself, or a change under a directory currently tracked because it contains an
// externally referenced file (see externalDirs).
func (p *Provider) isRelevantEvent(filename string) bool {
	if p.Directory != "" && (filename == p.Directory || strings.HasPrefix(filename, p.Directory+string(filepath.Separator))) {
		return true
	}

	if p.Directory == "" {
		_, evtFileName := filepath.Split(filename)
		_, confFileName := filepath.Split(p.Filename)
		if evtFileName == confFileName {
			return true
		}
	}

	p.watcherMu.RLock()
	defer p.watcherMu.RUnlock()

	_, ok := p.externalDirs[filepath.Dir(filename)]
	return ok
}

// syncExternalFileWatches updates the fsnotify watcher so that it watches the parent
// directories of exactly the external files (certificates, keys, CA bundles) currently
// referenced by the dynamic configuration
func (p *Provider) syncExternalFileWatches(refFiles []string) {
	p.watcherMu.Lock()
	defer p.watcherMu.Unlock()

	if p.watcher == nil {
		// Watch is disabled.
		return
	}

	wantedDirs := make(map[string]struct{}, len(refFiles))
	for _, f := range refFiles {
		dir := filepath.Dir(f)
		if p.isBaseDir(dir) {
			// Already permanently watched; nothing to add or, later, to remove.
			continue
		}
		wantedDirs[dir] = struct{}{}
	}

	for dir := range wantedDirs {
		if _, ok := p.externalDirs[dir]; ok {
			continue
		}

		if err := p.watcher.Add(dir); err != nil {
			log.Error().Err(err).Str("directory", dir).Msg("Error adding watcher for externally referenced file")
			continue
		}

		log.Debug().Msgf("add watcher on: %s", dir)
	}

	for dir := range p.externalDirs {
		if _, ok := wantedDirs[dir]; ok {
			continue
		}

		if err := p.watcher.Remove(dir); err != nil {
			log.Debug().Err(err).Str("directory", dir).Msg("Error removing watcher for externally referenced file")
			continue
		}

		log.Debug().Msgf("remove watcher on: %s", dir)
	}

	p.externalDirs = wantedDirs
}

// applyConfiguration builds the configuration and sends it to the given configurationChan.
func (p *Provider) applyConfiguration(configurationChan chan<- dynamic.Message) error {
	configuration, refFiles, err := p.buildConfiguration()
	if err != nil {
		return err
	}

	p.syncExternalFileWatches(refFiles)

	sendConfigToChannel(configurationChan, configuration)
	return nil
}

// buildConfiguration loads configuration either from file or a directory specified by
// 'Filename'/'Directory' and returns a 'Configuration' object, along with the paths of
// every external file (certificates, keys, CA bundles) it referenced.
func (p *Provider) buildConfiguration() (*dynamic.Configuration, []string, error) {
	ctx := log.With().Str(logs.ProviderName, providerName).Logger().WithContext(context.Background())

	if len(p.Directory) > 0 {
		return p.loadFileConfigFromDirectory(ctx, p.Directory, nil)
	}

	if len(p.Filename) > 0 {
		return p.loadFileConfig(ctx, p.Filename)
	}

	return nil, nil, errors.New("error using file configuration provider, neither filename nor directory is defined")
}

// loadFileConfig loads and decodes the configuration in filename, and returns the paths of
// every external file (certificates, keys, CA bundles) it referenced, so that the caller can
// keep watching them for changes (e.g. certificate renewal).
func (p *Provider) loadFileConfig(ctx context.Context, filename string) (*dynamic.Configuration, []string, error) {
	configuration, err := p.CreateConfiguration(ctx, filename, template.FuncMap{}, false)

	if err != nil {
		return nil, nil, err
	}

	// Collect every referenced external file path before any of the fields below get
	// overwritten with the file's content.
	refFiles := collectExternalFiles(configuration)

	if configuration.TLS != nil {
		configuration.TLS.Certificates = flattenCertificates(ctx, configuration.TLS)

		// TLS Options
		if configuration.TLS.Options != nil {
			for name, options := range configuration.TLS.Options {
				var caCerts []types.FileOrContent

				for _, caFile := range options.ClientAuth.CAFiles {
					content, err := caFile.Read()
					if err != nil {
						log.Ctx(ctx).Error().Err(err).Send()
						continue
					}

					caCerts = append(caCerts, types.FileOrContent(content))
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
				store.DefaultCertificate.CertFile = types.FileOrContent(content)

				content, err = store.DefaultCertificate.KeyFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				store.DefaultCertificate.KeyFile = types.FileOrContent(content)

				configuration.TLS.Stores[name] = store
			}
		}
	}

	// HTTP ServersTransport
	if configuration.HTTP != nil && len(configuration.HTTP.ServersTransports) > 0 {
		for name, st := range configuration.HTTP.ServersTransports {
			var certificates []tls.Certificate
			for _, cert := range st.Certificates {
				content, err := cert.CertFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				cert.CertFile = types.FileOrContent(content)

				content, err = cert.KeyFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				cert.KeyFile = types.FileOrContent(content)

				certificates = append(certificates, cert)
			}

			configuration.HTTP.ServersTransports[name].Certificates = certificates

			var rootCAs []types.FileOrContent
			for _, rootCA := range st.RootCAs {
				content, err := rootCA.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}

				rootCAs = append(rootCAs, types.FileOrContent(content))
			}

			st.RootCAs = rootCAs
		}
	}

	// TCP ServersTransport
	if configuration.TCP != nil && len(configuration.TCP.ServersTransports) > 0 {
		for name, st := range configuration.TCP.ServersTransports {
			var certificates []tls.Certificate
			if st.TLS == nil {
				continue
			}
			for _, cert := range st.TLS.Certificates {
				content, err := cert.CertFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				cert.CertFile = types.FileOrContent(content)

				content, err = cert.KeyFile.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}
				cert.KeyFile = types.FileOrContent(content)

				certificates = append(certificates, cert)
			}

			configuration.TCP.ServersTransports[name].TLS.Certificates = certificates

			var rootCAs []types.FileOrContent
			for _, rootCA := range st.TLS.RootCAs {
				content, err := rootCA.Read()
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Send()
					continue
				}

				rootCAs = append(rootCAs, types.FileOrContent(content))
			}

			st.TLS.RootCAs = rootCAs
		}
	}

	return configuration, refFiles, nil
}

// loadFileConfigFromDirectory recursively loads and merges configuration from every supported
// file in directory, and returns the paths of every external file (certificates, keys, CA
// bundles) referenced anywhere within it, so the caller can keep watching them for changes.
func (p *Provider) loadFileConfigFromDirectory(ctx context.Context, directory string, configuration *dynamic.Configuration) (*dynamic.Configuration, []string, error) {
	fileList, err := os.ReadDir(directory)
	if err != nil {
		return configuration, nil, fmt.Errorf("unable to read directory %s: %w", directory, err)
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
				Routers:           make(map[string]*dynamic.TCPRouter),
				Services:          make(map[string]*dynamic.TCPService),
				Middlewares:       make(map[string]*dynamic.TCPMiddleware),
				ServersTransports: make(map[string]*dynamic.TCPServersTransport),
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
	var refFiles []string

	for _, item := range fileList {
		logger := log.Ctx(ctx).With().Str("filename", item.Name()).Logger()

		if item.IsDir() {
			var subRefFiles []string
			configuration, subRefFiles, err = p.loadFileConfigFromDirectory(logger.WithContext(ctx), filepath.Join(directory, item.Name()), configuration)
			if err != nil {
				return configuration, refFiles, fmt.Errorf("unable to load content configuration from subdirectory %s: %w", item, err)
			}
			refFiles = append(refFiles, subRefFiles...)
			continue
		}

		if !isFileSupported(item.Name()) {
			logger.Debug().Msg("Skipping file, unsupported extension")
			continue
		}

		var c *dynamic.Configuration
		var cRefFiles []string
		c, cRefFiles, err = p.loadFileConfig(logger.WithContext(ctx), filepath.Join(directory, item.Name()))
		if err != nil {
			return configuration, refFiles, fmt.Errorf("%s: %w", filepath.Join(directory, item.Name()), err)
		}
		refFiles = append(refFiles, cRefFiles...)

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

		for name, conf := range c.TCP.ServersTransports {
			if _, exists := configuration.TCP.ServersTransports[name]; exists {
				logger.Warn().Str(logs.ServersTransportName, name).Msg("TCP servers transport already configured, skipping")
			} else {
				configuration.TCP.ServersTransports[name] = conf
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

	return configuration, refFiles, nil
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
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			Middlewares:       make(map[string]*dynamic.TCPMiddleware),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
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

func sendConfigToChannel(configurationChan chan<- dynamic.Message, configuration *dynamic.Configuration) {
	configurationChan <- dynamic.Message{
		ProviderName:  "file",
		Configuration: configuration,
	}
}

var fileOrContentType = reflect.TypeFor[types.FileOrContent]()

// collectExternalFiles walks configuration and returns the path of every types.FileOrContent
// field that refers to a file path, so the caller can keep watching those files (and, when they
// are deleted or renamed, their containing directories) for external changes such as certificate
// renewal. It is called before any of those fields get overwritten with the file's content.
func collectExternalFiles(configuration *dynamic.Configuration) []string {
	var paths []string
	walkFileOrContent(reflect.ValueOf(configuration), &paths)
	return paths
}

func walkFileOrContent(v reflect.Value, paths *[]string) {
	if !v.IsValid() {
		return
	}

	if v.Type() == fileOrContentType {
		if f, ok := v.Interface().(types.FileOrContent); ok && (f.IsPath() || looksLikeFilePath(f.String())) {
			*paths = append(*paths, f.String())
		}
		return
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return
		}
		walkFileOrContent(v.Elem(), paths)
	case reflect.Struct:
		for i := range v.NumField() {
			if v.Type().Field(i).PkgPath != "" {
				continue // unexported field.
			}
			walkFileOrContent(v.Field(i), paths)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			walkFileOrContent(v.Index(i), paths)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			walkFileOrContent(v.MapIndex(k), paths)
		}
	}
}

// looksLikeFilePath reports whether value looks like it was meant to be a file path, as opposed
// to inline certificate/CA content, regardless of whether that file currently exists on disk.
// Inline PEM content always spans multiple lines and starts with a "-----BEGIN" header; a real
// path never does. This lets us keep watching a certificate's directory when the file has been
// deleted mid-renewal, or has not been written yet by an external process (e.g. an ACME client
// that has not finished issuing it), instead of losing track of it the moment os.Stat fails.
func looksLikeFilePath(value string) bool {
	if value == "" {
		return false
	}
	return !strings.Contains(value, "\n") && !strings.HasPrefix(strings.TrimSpace(value), "-----BEGIN")
}

func flattenCertificates(ctx context.Context, tlsConfig *dynamic.TLSConfiguration) []*tls.CertAndStores {
	var certs []*tls.CertAndStores
	for _, cert := range tlsConfig.Certificates {
		content, err := cert.Certificate.CertFile.Read()
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Send()
			continue
		}
		cert.Certificate.CertFile = types.FileOrContent(string(content))

		content, err = cert.Certificate.KeyFile.Read()
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Send()
			continue
		}
		cert.Certificate.KeyFile = types.FileOrContent(string(content))

		certs = append(certs, cert)
	}

	return certs
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

func isFileSupported(filename string) bool {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".toml", ".yaml", ".yml":
		return true
	default:
		return false
	}
}
