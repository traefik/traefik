package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/plugins"
)

const typeName = "Plugin"

// PluginsBuilder the plugin's builder interface.
type PluginsBuilder interface {
	Build(pName string, config map[string]interface{}, middlewareName string) (plugins.Constructor, error)
}

func findPluginConfig(rawConfig map[string]dynamic.PluginConf) (string, map[string]interface{}, error) {
	if len(rawConfig) != 1 {
		return "", nil, errors.New("invalid configuration: no configuration or too many plugin definition")
	}

	var pluginType string
	var rawPluginConfig map[string]interface{}

	for pType, pConfig := range rawConfig {
		pluginType = pType
		rawPluginConfig = pConfig
	}

	if pluginType == "" {
		return "", nil, errors.New("missing plugin type")
	}

	return pluginType, rawPluginConfig, nil
}

type traceablePlugin struct {
	name string
	h    http.Handler
}

func newTraceablePlugin(ctx context.Context, name string, plug plugins.Constructor, next http.Handler) (*traceablePlugin, error) {
	h, err := plug(ctx, next)
	if err != nil {
		return nil, err
	}

	return &traceablePlugin{name: name, h: h}, nil
}

func (s *traceablePlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.h.ServeHTTP(rw, req)
}

func (s *traceablePlugin) GetTracingInformation() (string, string) {
	return s.name, typeName
}
