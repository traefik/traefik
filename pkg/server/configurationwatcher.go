package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// ConfigurationWatcher watches configuration changes.
type ConfigurationWatcher struct {
	providerAggregator provider.Provider

	defaultEntryPoints []string

	strictTLSOptions bool

	allProvidersConfigs chan dynamic.Message

	newConfigs chan dynamic.Configurations

	requiredProvider       string
	configurationListeners []func(dynamic.Configuration) error
	initialProviders       []string
	initialDone            chan struct{}
	initialOnce            sync.Once
	initialError           error

	configurationTransformers []func(context.Context, dynamic.Configurations) dynamic.Configurations

	routinesPool *safe.Pool
}

// NewConfigurationWatcher creates a new ConfigurationWatcher.
func NewConfigurationWatcher(
	routinesPool *safe.Pool,
	pvd provider.Provider,
	defaultEntryPoints []string,
	requiredProvider string,
	strictTLSOptions bool,
) *ConfigurationWatcher {
	return &ConfigurationWatcher{
		providerAggregator:  pvd,
		allProvidersConfigs: make(chan dynamic.Message, 100),
		newConfigs:          make(chan dynamic.Configurations),
		routinesPool:        routinesPool,
		defaultEntryPoints:  defaultEntryPoints,
		requiredProvider:    requiredProvider,
		strictTLSOptions:    strictTLSOptions,
		initialDone:         make(chan struct{}),
		initialProviders:    []string{requiredProvider},
	}
}

// Start the configuration watcher.
func (c *ConfigurationWatcher) Start() {
	c.routinesPool.GoCtx(c.receiveConfigurations)
	c.routinesPool.GoCtx(c.applyConfigurations)
	c.startProviderAggregator()
}

// Stop the configuration watcher.
func (c *ConfigurationWatcher) Stop() {
	close(c.allProvidersConfigs)
	close(c.newConfigs)
}

// AddListener adds a new listener function used when new configuration is provided.
func (c *ConfigurationWatcher) AddListener(listener func(dynamic.Configuration)) {
	c.AddCriticalListener(func(conf dynamic.Configuration) error { listener(conf); return nil })
}

// AddCriticalListener registers an application step whose failure prevents initial readiness.
func (c *ConfigurationWatcher) AddCriticalListener(listener func(dynamic.Configuration) error) {
	c.configurationListeners = append(c.configurationListeners, listener)
}

// SetRequiredProviders selects initial snapshots required before accepting traffic.
// Partial configurations still reach listeners, including certificate-provider feedback.
func (c *ConfigurationWatcher) SetRequiredProviders(names ...string) {
	c.initialProviders = slices.Clone(names)
	slices.Sort(c.initialProviders)
	c.initialProviders = slices.Compact(c.initialProviders)
}

// WaitInitialConfiguration waits for the first complete configuration to finish applying.
func (c *ConfigurationWatcher) WaitInitialConfiguration(ctx context.Context) error {
	log.Info().Strs("providers", c.initialProviders).Msg("Waiting for initial configuration before accepting traffic")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.initialDone:
		return c.initialError
	}
}

func (c *ConfigurationWatcher) finishInitialConfiguration(err error) {
	c.initialOnce.Do(func() { c.initialError = err; close(c.initialDone) })
}

func (c *ConfigurationWatcher) missingInitialProviders(configs dynamic.Configurations) []string {
	var missing []string
	for _, name := range c.initialProviders {
		if name == "" {
			continue
		}
		if _, ok := configs[name]; !ok {
			missing = append(missing, name)
		}
	}
	return missing
}

// AddTransformer registers a function to modify configurations before they are applied.
func (c *ConfigurationWatcher) AddTransformer(transformer func(context.Context, dynamic.Configurations) dynamic.Configurations) {
	c.configurationTransformers = append(c.configurationTransformers, transformer)
}

func (c *ConfigurationWatcher) startProviderAggregator() {
	log.Info().Msgf("Starting provider aggregator %T", c.providerAggregator)

	safe.Go(func() {
		err := c.providerAggregator.Provide(c.allProvidersConfigs, c.routinesPool)
		if err != nil {
			log.Error().Err(err).Msgf("Error starting provider aggregator %T", c.providerAggregator)
		}
	})
}

// receiveConfigurations receives configuration changes from the providers.
// The configuration message then gets passed along a series of checks, notably
// to verify that, for a given provider, the configuration that was just received
// is at least different from the previously received one.
// The full set of configurations is then sent to applyConfigurations
// via a channel in a non-blocking manner, ensuring the latest global state
// is always available for processing.
func (c *ConfigurationWatcher) receiveConfigurations(ctx context.Context) {
	newConfigurations := make(dynamic.Configurations)

	var output chan dynamic.Configurations

	var pending dynamic.Configurations

	for {
		select {
		case <-ctx.Done():
			return
		case output <- pending:
			output = nil
			pending = nil

		default:
			select {
			case <-ctx.Done():
				return
			case configMsg, ok := <-c.allProvidersConfigs:
				if !ok {
					return
				}

				logger := log.Ctx(ctx).With().Str(logs.ProviderName, configMsg.ProviderName).Logger()

				if configMsg.Configuration == nil {
					logger.Debug().Msg("Skipping nil configuration")
					continue
				}

				if isEmptyConfiguration(configMsg.Configuration) && !slices.Contains(c.initialProviders, configMsg.ProviderName) {
					logger.Debug().Msg("Skipping empty configuration")
					continue
				}

				logConfiguration(logger, configMsg)

				if reflect.DeepEqual(newConfigurations[configMsg.ProviderName], configMsg.Configuration) {
					// no change, do nothing.
					logger.Debug().Msg("Skipping unchanged configuration")
					continue
				}

				newConfigurations[configMsg.ProviderName] = configMsg.Configuration.DeepCopy()

				// DeepCopy is necessary because newConfigurations gets modified later by the consumer of c.newConfigs.
				transformedConfigurations := newConfigurations.DeepCopy()
				for _, transform := range c.configurationTransformers {
					// Each transformer gets its own copy because a transformer could keep a reference to the one it received.
					transformedConfigurations = transform(logger.WithContext(ctx), transformedConfigurations)
					transformedConfigurations = transformedConfigurations.DeepCopy()
				}

				output = c.newConfigs
				pending = transformedConfigurations

			case output <- pending:
				output = nil
				pending = nil
			}
		}
	}
}

// applyConfigurations receives the full set of configurations from
// receiveConfigurations and applies them if they differ from the previous set.
// It waits for the required provider's configuration before applying any configs.
func (c *ConfigurationWatcher) applyConfigurations(ctx context.Context) {
	var lastConfigurations dynamic.Configurations
	for {
		select {
		case <-ctx.Done():
			return
		case newConfigs, ok := <-c.newConfigs:
			if !ok {
				return
			}

			// We wait for first configuration of the required provider before applying configurations.
			if _, ok := newConfigs[c.requiredProvider]; c.requiredProvider != "" && !ok {
				continue
			}

			if reflect.DeepEqual(newConfigs, lastConfigurations) {
				continue
			}

			conf := mergeConfiguration(newConfigs.DeepCopy(), c.defaultEntryPoints)
			conf = applyModel(conf)
			if conf.HTTP != nil {
				conf.HTTP.Routers = resolveHTTPTLSOptions(conf.HTTP.Routers, c.strictTLSOptions)
			}

			var applyErrors []error
			for _, listener := range c.configurationListeners {
				if err := listener(conf); err != nil {
					log.Error().Err(err).Msg("Configuration application failed")
					applyErrors = append(applyErrors, err)
				}
			}
			if missing := c.missingInitialProviders(newConfigs); len(missing) == 0 {
				err := errors.Join(applyErrors...)
				if err != nil {
					err = fmt.Errorf("applying initial configuration: %w", err)
				}
				c.finishInitialConfiguration(err)
			} else {
				log.Info().Strs("providers", missing).Msg("Initial configuration is waiting for provider snapshots")
			}
			lastConfigurations = newConfigs
		}
	}
}

func logConfiguration(logger zerolog.Logger, configMsg dynamic.Message) {
	if logger.GetLevel() > zerolog.DebugLevel {
		return
	}

	copyConf := configMsg.Configuration.DeepCopy()
	if copyConf.TLS != nil {
		copyConf.TLS.Certificates = nil

		if copyConf.TLS.Options != nil {
			cleanedOptions := make(map[string]tls.Options, len(copyConf.TLS.Options))
			for name, option := range copyConf.TLS.Options {
				option.ClientAuth.CAFiles = []types.FileOrContent{}
				cleanedOptions[name] = option
			}

			copyConf.TLS.Options = cleanedOptions
		}

		for k := range copyConf.TLS.Stores {
			st := copyConf.TLS.Stores[k]
			st.DefaultCertificate = nil
			copyConf.TLS.Stores[k] = st
		}
	}

	if copyConf.HTTP != nil {
		for _, transport := range copyConf.HTTP.ServersTransports {
			transport.Certificates = tls.Certificates{}
			transport.RootCAs = []types.FileOrContent{}
		}
	}

	if copyConf.TCP != nil {
		for _, transport := range copyConf.TCP.ServersTransports {
			if transport.TLS != nil {
				transport.TLS.Certificates = tls.Certificates{}
				transport.TLS.RootCAs = []types.FileOrContent{}
			}
		}
	}

	jsonConf, err := json.Marshal(copyConf)
	if err != nil {
		logger.Error().Err(err).Msg("Could not marshal dynamic configuration")
		logger.Debug().Msgf("Configuration received: [struct] %#v", copyConf)
	} else {
		logger.Debug().RawJSON("config", jsonConf).Msg("Configuration received")
	}
}

func isEmptyConfiguration(conf *dynamic.Configuration) bool {
	if conf.TCP == nil {
		conf.TCP = &dynamic.TCPConfiguration{}
	}
	if conf.HTTP == nil {
		conf.HTTP = &dynamic.HTTPConfiguration{}
	}
	if conf.UDP == nil {
		conf.UDP = &dynamic.UDPConfiguration{}
	}

	httpEmpty := conf.HTTP.Routers == nil && conf.HTTP.Services == nil && conf.HTTP.Middlewares == nil
	tlsEmpty := conf.TLS == nil || conf.TLS.Certificates == nil && conf.TLS.Stores == nil && conf.TLS.Options == nil
	tcpEmpty := conf.TCP.Routers == nil && conf.TCP.Services == nil && conf.TCP.Middlewares == nil
	udpEmpty := conf.UDP.Routers == nil && conf.UDP.Services == nil

	return httpEmpty && tlsEmpty && tcpEmpty && udpEmpty
}
