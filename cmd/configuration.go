package cmd

import (
	"time"

	"github.com/baqupio/baqup/v3/pkg/config/static"
	ptypes "github.com/traefik/paerser/types"
)

// BaqupCmdConfiguration wraps the static configuration and extra parameters.
type BaqupCmdConfiguration struct {
	static.Configuration `export:"true"`
	// ConfigFile is the path to the configuration file.
	ConfigFile string `description:"Configuration file to use. If specified all other flags are ignored." export:"true"`
}

// NewBaqupConfiguration creates a BaqupCmdConfiguration with default values.
func NewBaqupConfiguration() *BaqupCmdConfiguration {
	return &BaqupCmdConfiguration{
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
			TCPServersTransport: &static.TCPServersTransport{
				DialTimeout:   ptypes.Duration(30 * time.Second),
				DialKeepAlive: ptypes.Duration(15 * time.Second),
			},
		},
		ConfigFile: "",
	}
}
