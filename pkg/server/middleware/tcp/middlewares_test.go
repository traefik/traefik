package tcpmiddleware

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/plugins"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

type fakePluginsBuilder struct {
	pluginType string
	config     map[string]any
	buildErr   error
}

func (b *fakePluginsBuilder) BuildTCP(pName string, config map[string]any, _ string) (plugins.TCPConstructor, error) {
	b.pluginType = pName
	b.config = config
	if b.buildErr != nil {
		return nil, b.buildErr
	}

	return func(_ context.Context, next tcp.Handler) (tcp.Handler, error) {
		return next, nil
	}, nil
}

func TestBuilder_BuildChain_plugin(t *testing.T) {
	testCases := []struct {
		desc           string
		plugin         map[string]dynamic.PluginConf
		pluginsBuilder *fakePluginsBuilder
		expectedErr    string
	}{
		{
			desc:           "plugin",
			plugin:         map[string]dynamic.PluginConf{"my-plugin": {"banner": "hello"}},
			pluginsBuilder: &fakePluginsBuilder{},
		},
		{
			desc:        "no plugin builder",
			plugin:      map[string]dynamic.PluginConf{"my-plugin": {"banner": "hello"}},
			expectedErr: `invalid middleware "plugin@file" configuration: invalid middleware type or middleware does not exist`,
		},
		{
			desc:           "several plugins",
			plugin:         map[string]dynamic.PluginConf{"my-plugin": {}, "other-plugin": {}},
			pluginsBuilder: &fakePluginsBuilder{},
			expectedErr:    "plugin: invalid configuration: no configuration or too many plugin definition",
		},
		{
			desc:           "build error",
			plugin:         map[string]dynamic.PluginConf{"my-plugin": {"banner": "hello"}},
			pluginsBuilder: &fakePluginsBuilder{buildErr: errors.New("plugin my-plugin does not support TCP middlewares")},
			expectedErr:    "plugin: plugin my-plugin does not support TCP middlewares",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			rtConf := runtime.NewConfig(dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"plugin@file": {Plugin: test.plugin},
					},
				},
			})

			var pluginsBuilder PluginsBuilder
			if test.pluginsBuilder != nil {
				pluginsBuilder = test.pluginsBuilder
			}

			builder := NewBuilder(rtConf.TCPMiddlewares, pluginsBuilder)

			handler, err := builder.BuildChain(t.Context(), []string{"plugin@file"}).Then(tcp.HandlerFunc(func(tcp.WriteCloser) {}))
			if test.expectedErr != "" {
				require.EqualError(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, handler)
			assert.Equal(t, "my-plugin", test.pluginsBuilder.pluginType)
			assert.Equal(t, map[string]any{"banner": "hello"}, test.pluginsBuilder.config)
		})
	}
}
