package cmd

import (
	"time"

	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/types"
)

// TraefikCmdConfiguration holds GlobalConfiguration and other stuff
type TraefikCmdConfiguration struct {
	static.Configuration `export:"true"`
	ConfigFile           string `description:"Configuration file to use." export:"true"`
}

// NewTraefikConfiguration creates a TraefikCmdConfiguration with default values
func NewTraefikConfiguration() *TraefikCmdConfiguration {
	return &TraefikCmdConfiguration{
		Configuration: static.Configuration{
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
		},
		ConfigFile: "",
	}
}
