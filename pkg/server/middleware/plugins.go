package middleware

import (
	"errors"
	"fmt"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
)

func findPluginConfig(rawConfig map[string]dynamic.PluginConf) (string, map[string]interface{}, error) {
	if len(rawConfig) != 1 {
		return "", nil, errors.New("plugin: invalid configuration: no configuration or too many plugin definition")
	}

	var pluginType string
	var rawPluginConfig map[string]interface{}

	for pType, pConfig := range rawConfig {
		pluginType = pType
		rawPluginConfig = pConfig
	}

	if pluginType == "" {
		return "", nil, errors.New("plugin: missing plugin type")
	}

	if len(rawPluginConfig) == 0 {
		return "", nil, fmt.Errorf("plugin: missing plugin configuration: %s", pluginType)
	}

	return pluginType, rawPluginConfig, nil
}
