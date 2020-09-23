package main

import (
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/plugins"
)

const outputDir = "./plugins-storage/"

func initPlugins(staticCfg *static.Configuration) (*plugins.Client, map[string]plugins.Descriptor, *plugins.DevPlugin, error) {
	if !isPilotEnabled(staticCfg) || !hasPlugins(staticCfg) {
		return nil, map[string]plugins.Descriptor{}, nil, nil
	}

	opts := plugins.ClientOptions{
		Output: outputDir,
		Token:  staticCfg.Pilot.Token,
	}

	client, err := plugins.NewClient(opts)
	if err != nil {
		return nil, nil, nil, err
	}

	err = plugins.Setup(client, staticCfg.Experimental.Plugins, staticCfg.Experimental.DevPlugin)
	if err != nil {
		return nil, nil, nil, err
	}

	return client, staticCfg.Experimental.Plugins, staticCfg.Experimental.DevPlugin, nil
}

func isPilotEnabled(staticCfg *static.Configuration) bool {
	return staticCfg.Pilot != nil && staticCfg.Pilot.Token != ""
}

func hasPlugins(staticCfg *static.Configuration) bool {
	return staticCfg.Experimental != nil &&
		(len(staticCfg.Experimental.Plugins) > 0 || staticCfg.Experimental.DevPlugin != nil)
}
