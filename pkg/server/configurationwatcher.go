package server

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/eapache/channels"
	"github.com/sirupsen/logrus"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/tls"
)

// ConfigurationWatcher watches configuration changes.
type ConfigurationWatcher struct {
	provider provider.Provider

	defaultEntryPoints []string

	providersThrottleDuration time.Duration

	allProvidersConfigs chan dynamic.Message

	newConfigs *channels.RingChannel

	requiredProvider       string
	configurationListeners []func(dynamic.Configuration)

	routinesPool *safe.Pool
}

// NewConfigurationWatcher creates a new ConfigurationWatcher.
func NewConfigurationWatcher(
	routinesPool *safe.Pool,
	pvd provider.Provider,
	providersThrottleDuration time.Duration,
	defaultEntryPoints []string,
	requiredProvider string,
) *ConfigurationWatcher {
	watcher := &ConfigurationWatcher{
		provider:                  pvd,
		allProvidersConfigs:       make(chan dynamic.Message, 100),
		newConfigs:                channels.NewRingChannel(1),
		providersThrottleDuration: providersThrottleDuration,
		routinesPool:              routinesPool,
		defaultEntryPoints:        defaultEntryPoints,
		requiredProvider:          requiredProvider,
	}

	return watcher
}

// Start the configuration watcher.
func (c *ConfigurationWatcher) Start() {
	c.routinesPool.GoCtx(c.receiveConfigurations)
	c.routinesPool.GoCtx(c.throttleAndApplyConfigurations)
	c.startProvider()
}

// Stop the configuration watcher.
func (c *ConfigurationWatcher) Stop() {
	close(c.allProvidersConfigs)
	c.newConfigs.Close()
}

// AddListener adds a new listener function used when new configuration is provided.
func (c *ConfigurationWatcher) AddListener(listener func(dynamic.Configuration)) {
	if c.configurationListeners == nil {
		c.configurationListeners = make([]func(dynamic.Configuration), 0)
	}
	c.configurationListeners = append(c.configurationListeners, listener)
}

func (c *ConfigurationWatcher) startProvider() {
	logger := log.WithoutContext()

	jsonConf, err := json.Marshal(c.provider)
	if err != nil {
		logger.Debugf("Unable to marshal provider configuration %T: %v", c.provider, err)
	}

	logger.Infof("Starting provider %T %s", c.provider, jsonConf)
	currentProvider := c.provider

	safe.Go(func() {
		err := currentProvider.Provide(c.allProvidersConfigs, c.routinesPool)
		if err != nil {
			logger.Errorf("Error starting provider %T: %s", currentProvider, err)
		}
	})
}

// receiveConfigurations receives configuration changes from the providers.
// The configuration message then gets passed along a series of check, notably
// to verify that, for a given provider, the configuration that was just received
// is at least different from the previously received one.
// The full set of configurations is then sent to the throttling goroutine,
// (throttleAndApplyConfigurations) via a RingChannel, which ensures that we can
// constantly send in a non-blocking way to the throttling goroutine the last
// global state we are aware of.
func (c *ConfigurationWatcher) receiveConfigurations(ctx context.Context) {
	lastConfigurations := make(dynamic.Configurations)
	for {
		select {
		case <-ctx.Done():
			return
		case configMsg, ok := <-c.allProvidersConfigs:
			if !ok {
				return
			}

			logger := log.WithoutContext().WithField(log.ProviderName, configMsg.ProviderName)

			if configMsg.Configuration == nil {
				logger.Debug("Received nil configuration from provider, skipping.")
				continue
			}

			preLoadConfiguration(logger, configMsg)

			if isEmptyConfiguration(configMsg.Configuration) {
				logger.Info("Skipping empty Configuration")
				continue
			}

			if reflect.DeepEqual(lastConfigurations[configMsg.ProviderName], configMsg.Configuration) {
				// no change, do nothing
				logger.Info("Skipping unchanged configuration")
				continue
			}

			newConfigurations := lastConfigurations.DeepCopy()
			newConfigurations[configMsg.ProviderName] = configMsg.Configuration.DeepCopy()

			c.newConfigs.In() <- newConfigurations
			lastConfigurations = newConfigurations
		}
	}
}

func preLoadConfiguration(logger log.Logger, configMsg dynamic.Message) {
	if log.GetLevel() != logrus.DebugLevel {
		return
	}

	copyConf := configMsg.Configuration.DeepCopy()
	if copyConf.TLS != nil {
		copyConf.TLS.Certificates = nil

		if copyConf.TLS.Options != nil {
			cleanedOptions := make(map[string]tls.Options, len(copyConf.TLS.Options))
			for name, option := range copyConf.TLS.Options {
				option.ClientAuth.CAFiles = []tls.FileOrContent{}
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
			transport.RootCAs = []tls.FileOrContent{}
		}
	}

	jsonConf, err := json.Marshal(copyConf)
	if err != nil {
		// TODO verify that we have the provider name in the log
		logger.Errorf("Could not marshal dynamic configuration: %v", err)
		logger.Debugf("Configuration received: [struct] %#v", copyConf)
	} else {
		logger.Debugf("Configuration received: %s", string(jsonConf))
	}
}

// throttleAndApplyConfigurations blocks on a RingChannel that receives the new
// set of configurations that is compiled and sent by receiveConfigurations as soon
// as a provider change occurs. If the new set is different from the previous set
// that had been applied, the new set is applied, and we sleep for a a while before
// listening on the channel again.
func (c *ConfigurationWatcher) throttleAndApplyConfigurations(ctx context.Context) {
	var lastConfigurations dynamic.Configurations
	for {
		select {
		case <-ctx.Done():
			return
		case newConfigs := <-c.newConfigs.Out():
			currentConfigurations := newConfigs.(dynamic.Configurations)

			if !reflect.DeepEqual(currentConfigurations, lastConfigurations) {
				c.applyConfigurations(currentConfigurations)
				lastConfigurations = currentConfigurations
			}

			time.Sleep(c.providersThrottleDuration)
		}
	}
}

func (c *ConfigurationWatcher) applyConfigurations(currentConfigurations dynamic.Configurations) {
	// We wait for first configuration of the required provider before applying configurations.
	if _, ok := currentConfigurations[c.requiredProvider]; c.requiredProvider != "" && !ok {
		return
	}

	conf := mergeConfiguration(currentConfigurations.DeepCopy(), c.defaultEntryPoints)
	conf = applyModel(conf)

	for _, listener := range c.configurationListeners {
		listener(conf)
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
