package cmd

import (
	"time"

	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/static"
)

// TraefikCmdConfiguration wraps the static configuration and extra parameters.
type TraefikCmdConfiguration struct {
	static.Configuration `export:"true"`
	// ConfigFile is the path to the configuration file.
	ConfigFile string `description:"Configuration file to use. If specified all other flags are ignored." export:"true"`
}

// NewTraefikConfiguration creates a TraefikCmdConfiguration with default values.
func NewTraefikConfiguration() *TraefikCmdConfiguration {
	return &TraefikCmdConfiguration{
		Configuration: static.Configuration{
			Global: &static.Global{
				CheckNewVersion: true,
			},
			EntryPoints: make(static.EntryPoints),
			Providers: &static.Providers{
				ProvidersThrottleDuration: ptypes.Duration(2 * time.Second),
			},
			ServersTransport: &static.ServersTransport{
				MaxIdleConnsPerHost: 200,
			},
		},
		ConfigFile: "",
	}
}
