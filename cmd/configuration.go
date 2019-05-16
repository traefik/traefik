package cmd

import (
	"time"

	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/types"
)

// NewTraefikConfiguration creates a TraefikConfiguration with default values
func NewTraefikConfiguration() static.Configuration {
	return static.Configuration{
		Global: &static.Global{
			CheckNewVersion: true,
		},
		EntryPoints: make(static.EntryPoints),
		Providers: &static.Providers{
			ProvidersThrottleDuration: types.Duration(2 * time.Second),
		},
		ServersTransport: &static.ServersTransport{
			MaxIdleConnsPerHost: 200,
		},
	}
}
