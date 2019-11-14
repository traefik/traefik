package server

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/eapache/channels"
	"github.com/sirupsen/logrus"
)

// ConfigurationWatcher watches configuration changes.
type ConfigurationWatcher struct {
	provider provider.Provider

	providersThrottleDuration time.Duration

	currentConfigurations safe.Safe

	configurationChan          chan dynamic.Message
	configurationValidatedChan chan dynamic.Message
	providerConfigUpdateMap    map[string]chan dynamic.Message

	configurationListeners []func(dynamic.Configuration)

	routinesPool *safe.Pool
}

// NewConfigurationWatcher creates a new ConfigurationWatcher.
func NewConfigurationWatcher(routinesPool *safe.Pool, pvd provider.Provider, providersThrottleDuration time.Duration) *ConfigurationWatcher {
	watcher := &ConfigurationWatcher{
		provider:                   pvd,
		configurationChan:          make(chan dynamic.Message, 100),
		configurationValidatedChan: make(chan dynamic.Message, 100),
		providerConfigUpdateMap:    make(map[string]chan dynamic.Message),
		providersThrottleDuration:  providersThrottleDuration,
		routinesPool:               routinesPool,
	}

	currentConfigurations := make(dynamic.Configurations)
	watcher.currentConfigurations.Set(currentConfigurations)

	return watcher
}

// Start the configuration watcher.
func (c *ConfigurationWatcher) Start() {
	c.routinesPool.Go(func(stop chan bool) { c.listenProviders(stop) })
	c.routinesPool.Go(func(stop chan bool) { c.listenConfigurations(stop) })
	c.startProvider()
}

// Stop the configuration watcher.
func (c *ConfigurationWatcher) Stop() {
	close(c.configurationChan)
	close(c.configurationValidatedChan)
}

// AddListener adds a new listener function used when new configuration is provided
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
		err := currentProvider.Provide(c.configurationChan, c.routinesPool)
		if err != nil {
			logger.Errorf("Error starting provider %T: %s", c.provider, err)
		}
	})
}

// listenProviders receives configuration changes from the providers.
// The configuration message then gets passed along a series of check
// to finally end up in a throttler that sends it to listenConfigurations (through c. configurationValidatedChan).
func (c *ConfigurationWatcher) listenProviders(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-c.configurationChan:
			if !ok {
				return
			}

			if configMsg.Configuration == nil {
				log.WithoutContext().WithField(log.ProviderName, configMsg.ProviderName).
					Debug("Received nil configuration from provider, skipping.")
				return
			}

			c.preLoadConfiguration(configMsg)
		}
	}
}

func (c *ConfigurationWatcher) listenConfigurations(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-c.configurationValidatedChan:
			if !ok || configMsg.Configuration == nil {
				return
			}
			c.loadMessage(configMsg)
		}
	}
}

func (c *ConfigurationWatcher) loadMessage(configMsg dynamic.Message) {
	currentConfigurations := c.currentConfigurations.Get().(dynamic.Configurations)

	// Copy configurations to new map so we don't change current if LoadConfig fails
	newConfigurations := currentConfigurations.DeepCopy()
	newConfigurations[configMsg.ProviderName] = configMsg.Configuration

	c.currentConfigurations.Set(newConfigurations)

	conf := mergeConfiguration(newConfigurations)

	for _, listener := range c.configurationListeners {
		listener(conf)
	}
}

func (c *ConfigurationWatcher) preLoadConfiguration(configMsg dynamic.Message) {
	currentConfigurations := c.currentConfigurations.Get().(dynamic.Configurations)

	logger := log.WithoutContext().WithField(log.ProviderName, configMsg.ProviderName)
	if log.GetLevel() == logrus.DebugLevel {
		copyConf := configMsg.Configuration.DeepCopy()
		if copyConf.TLS != nil {
			copyConf.TLS.Certificates = nil

			for _, v := range copyConf.TLS.Stores {
				v.DefaultCertificate = nil
			}
		}

		jsonConf, err := json.Marshal(copyConf)
		if err != nil {
			logger.Errorf("Could not marshal dynamic configuration: %v", err)
			logger.Debugf("Configuration received from provider %s: [struct] %#v", configMsg.ProviderName, copyConf)
		} else {
			logger.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))
		}
	}

	if isEmptyConfiguration(configMsg.Configuration) {
		logger.Infof("Skipping empty Configuration for provider %s", configMsg.ProviderName)
		return
	}

	if reflect.DeepEqual(currentConfigurations[configMsg.ProviderName], configMsg.Configuration) {
		logger.Infof("Skipping same configuration for provider %s", configMsg.ProviderName)
		return
	}

	providerConfigUpdateCh, ok := c.providerConfigUpdateMap[configMsg.ProviderName]
	if !ok {
		providerConfigUpdateCh = make(chan dynamic.Message)
		c.providerConfigUpdateMap[configMsg.ProviderName] = providerConfigUpdateCh
		c.routinesPool.Go(func(stop chan bool) {
			c.throttleProviderConfigReload(c.providersThrottleDuration, c.configurationValidatedChan, providerConfigUpdateCh, stop)
		})
	}

	providerConfigUpdateCh <- configMsg
}

// throttleProviderConfigReload throttles the configuration reload speed for a single provider.
// It will immediately publish a new configuration and then only publish the next configuration after the throttle duration.
// Note that in the case it receives N new configs in the timeframe of the throttle duration after publishing,
// it will publish the last of the newly received configurations.
func (c *ConfigurationWatcher) throttleProviderConfigReload(throttle time.Duration, publish chan<- dynamic.Message, in <-chan dynamic.Message, stop chan bool) {
	ring := channels.NewRingChannel(1)
	defer ring.Close()

	c.routinesPool.Go(func(stop chan bool) {
		for {
			select {
			case <-stop:
				return
			case nextConfig := <-ring.Out():
				if config, ok := nextConfig.(dynamic.Message); ok {
					publish <- config
					time.Sleep(throttle)
				}
			}
		}
	})

	for {
		select {
		case <-stop:
			return
		case nextConfig := <-in:
			ring.In() <- nextConfig
		}
	}
}

func isEmptyConfiguration(conf *dynamic.Configuration) bool {
	if conf == nil {
		return true
	}

	if conf.TCP == nil {
		conf.TCP = &dynamic.TCPConfiguration{}
	}
	if conf.HTTP == nil {
		conf.HTTP = &dynamic.HTTPConfiguration{}
	}

	httpEmpty := conf.HTTP.Routers == nil && conf.HTTP.Services == nil && conf.HTTP.Middlewares == nil
	tlsEmpty := conf.TLS == nil || conf.TLS.Certificates == nil && conf.TLS.Stores == nil && conf.TLS.Options == nil
	tcpEmpty := conf.TCP.Routers == nil && conf.TCP.Services == nil

	return httpEmpty && tlsEmpty && tcpEmpty
}
