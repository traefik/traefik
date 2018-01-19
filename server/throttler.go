package server

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/eapache/channels"
)

// providerUpdateThrottler is responsible for throttling configuration updates
// in order to prevent overloading the server.
type providerUpdateThrottler struct {
	server           *Server
	throttleDuration time.Duration
	// providerThrottledConfigUpdates manages per-provider channels to throttle
	// configuration updates.
	providerThrottledConfigUpdates map[string]chan types.ConfigMessage
	// currentDirtyConfigurations represents configurations which have been
	// accepted already but may still be dirty (i.e., not persisted into the
	// server struct yet).
	currentDirtyConfigurations types.Configurations
}

func newProviderUpdateThrottler(s *Server, throttleDuration time.Duration) *providerUpdateThrottler {
	return &providerUpdateThrottler{
		server:                         s,
		throttleDuration:               throttleDuration,
		providerThrottledConfigUpdates: make(map[string]chan types.ConfigMessage),
		currentDirtyConfigurations:     make(types.Configurations),
	}
}

func (t *providerUpdateThrottler) process(stop chan bool, configMsg types.ConfigMessage) {
	t.server.defaultConfigurationValues(configMsg.Configuration)
	provName := configMsg.ProviderName
	jsonConf, _ := json.Marshal(configMsg.Configuration)
	log.Debugf("Configuration received from provider %s: %s", provName, string(jsonConf))

	currentConfig, ok := t.currentDirtyConfigurations[provName]
	if !ok {
		currentConfigurations := t.server.currentConfigurations.Get().(types.Configurations)
		currentConfig = currentConfigurations[provName]
		t.currentDirtyConfigurations[provName] = currentConfig
	}

	if configMsg.Configuration == nil ||
		configMsg.Configuration.Backends == nil &&
			configMsg.Configuration.Frontends == nil &&
			configMsg.Configuration.TLSConfiguration == nil {
		log.Infof("Skipping empty configuration for provider %s", provName)
	} else if reflect.DeepEqual(currentConfig, configMsg.Configuration) {
		log.Infof("Skipping same configuration for provider %s", provName)
	} else {
		t.currentDirtyConfigurations[provName] = configMsg.Configuration
		providerConfigUpdateCh, ok := t.providerThrottledConfigUpdates[provName]
		if !ok {
			providerConfigUpdateCh = make(chan types.ConfigMessage)
			t.providerThrottledConfigUpdates[provName] = providerConfigUpdateCh
			safe.Go(func() {
				t.throttleProviderConfigReload(t.server.configurationValidatedChan, providerConfigUpdateCh, stop)
			})
		}
		providerConfigUpdateCh <- configMsg
	}
}

// throttleProviderConfigReload throttles the configuration reload speed for a single provider.
// It will immediately publish a new configuration and then only publish the next configuration after the throttle duration.
// Note that in the case it receives N new configs in the timeframe of the throttle duration after publishing,
// it will publish the last of the newly received configurations.
func (t *providerUpdateThrottler) throttleProviderConfigReload(publish chan<- types.ConfigMessage, in <-chan types.ConfigMessage, stop chan bool) {
	ring := channels.NewRingChannel(1)
	defer ring.Close()

	safe.Go(func() {
		for {
			select {
			case <-stop:
				return
			case nextConfig, more := <-ring.Out():
				if !more {
					return
				}
				publish <- nextConfig.(types.ConfigMessage)
				time.Sleep(t.throttleDuration)
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
