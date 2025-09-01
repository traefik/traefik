package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/plugins"
)

const outputDir = "./plugins-storage/"
const pluginsURL = "https://plugins.traefik.io/public/"

func createPluginBuilder(staticConfiguration *static.Configuration) (*plugins.Builder, error) {
	client, plgs, localPlgs, err := initPlugins(staticConfiguration)
	if err != nil {
		return nil, err
	}

	return plugins.NewBuilder(client, plgs, localPlgs)
}

func initPlugins(staticCfg *static.Configuration) (*plugins.Client, map[string]plugins.Descriptor, map[string]plugins.LocalDescriptor, error) {
	err := checkUniquePluginNames(staticCfg.Experimental)
	if err != nil {
		return nil, nil, nil, err
	}

	var client *plugins.Client
	plgs := map[string]plugins.Descriptor{}

	if hasPlugins(staticCfg) {
		httpClient := retryablehttp.NewClient()
		httpClient.Logger = logs.NewRetryableHTTPLogger(log.Logger)
		httpClient.HTTPClient = &http.Client{Timeout: 10 * time.Second}
		httpClient.RetryMax = 3

		opts := plugins.ClientOptions{
			HTTPClient: httpClient.StandardClient(),
			BaseURL:    pluginsURL,
			Output:     outputDir,
		}

		var err error
		client, err = plugins.NewClient(opts)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to create plugins client: %w", err)
		}

		err = plugins.SetupRemotePlugins(client, staticCfg.Experimental.Plugins)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to set up plugins environment: %w", err)
		}

		plgs = staticCfg.Experimental.Plugins
	}

	localPlgs := map[string]plugins.LocalDescriptor{}

	if hasLocalPlugins(staticCfg) {
		err := plugins.SetupLocalPlugins(staticCfg.Experimental.LocalPlugins)
		if err != nil {
			return nil, nil, nil, err
		}

		localPlgs = staticCfg.Experimental.LocalPlugins
	}

	return client, plgs, localPlgs, nil
}

func checkUniquePluginNames(e *static.Experimental) error {
	if e == nil {
		return nil
	}

	for s := range e.LocalPlugins {
		if _, ok := e.Plugins[s]; ok {
			return fmt.Errorf("the plugin's name %q must be unique", s)
		}
	}

	return nil
}

func hasPlugins(staticCfg *static.Configuration) bool {
	return staticCfg.Experimental != nil && len(staticCfg.Experimental.Plugins) > 0
}

func hasLocalPlugins(staticCfg *static.Configuration) bool {
	return staticCfg.Experimental != nil && len(staticCfg.Experimental.LocalPlugins) > 0
}
