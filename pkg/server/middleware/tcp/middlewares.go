package tcpmiddleware

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/middlewares/tcp/inflightconn"
	"github.com/traefik/traefik/v3/pkg/middlewares/tcp/ipallowlist"
	"github.com/traefik/traefik/v3/pkg/middlewares/tcp/ipwhitelist"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

type middlewareStackType int

const (
	middlewareStackKey middlewareStackType = iota
)

// PluginsBuilder the plugin's builder interface.
type PluginsBuilder interface {
	BuildTCP(pName string, config map[string]interface{}, middlewareName string) (tcp.Constructor, error)
}

// Builder the middleware builder.
type Builder struct {
	configs       map[string]*runtime.TCPMiddlewareInfo
	pluginBuilder PluginsBuilder
}

// NewBuilder creates a new Builder.
func NewBuilder(configs map[string]*runtime.TCPMiddlewareInfo, pluginBuilder PluginsBuilder) *Builder {
	return &Builder{
		configs:       configs,
		pluginBuilder: pluginBuilder,
	}
}

// BuildChain creates a middleware chain.
func (b *Builder) BuildChain(ctx context.Context, middlewares []string) *tcp.Chain {
	chain := tcp.NewChain()

	for _, name := range middlewares {
		middlewareName := provider.GetQualifiedName(ctx, name)

		chain = chain.Append(func(next tcp.Handler) (tcp.Handler, error) {
			constructorContext := provider.AddInContext(ctx, middlewareName)
			if midInf, ok := b.configs[middlewareName]; !ok || midInf.TCPMiddleware == nil {
				return nil, fmt.Errorf("middleware %q does not exist", middlewareName)
			}

			var err error
			if constructorContext, err = checkRecursion(constructorContext, middlewareName); err != nil {
				b.configs[middlewareName].AddError(err, true)
				return nil, err
			}

			constructor, err := b.buildConstructor(constructorContext, middlewareName)
			if err != nil {
				b.configs[middlewareName].AddError(err, true)
				return nil, err
			}

			handler, err := constructor(next)
			if err != nil {
				b.configs[middlewareName].AddError(err, true)
				return nil, err
			}

			return handler, nil
		})
	}

	return &chain
}

func checkRecursion(ctx context.Context, middlewareName string) (context.Context, error) {
	currentStack, ok := ctx.Value(middlewareStackKey).([]string)
	if !ok {
		currentStack = []string{}
	}

	if slices.Contains(currentStack, middlewareName) {
		return ctx, fmt.Errorf("could not instantiate middleware %s: recursion detected in %s", middlewareName, strings.Join(append(currentStack, middlewareName), "->"))
	}

	return context.WithValue(ctx, middlewareStackKey, append(currentStack, middlewareName)), nil
}

func (b *Builder) buildConstructor(ctx context.Context, middlewareName string) (tcp.Constructor, error) {
	config := b.configs[middlewareName]
	if config == nil || config.TCPMiddleware == nil {
		return nil, fmt.Errorf("invalid middleware %q configuration", middlewareName)
	}

	var middleware tcp.Constructor

	// InFlightConn
	if config.InFlightConn != nil {
		middleware = func(next tcp.Handler) (tcp.Handler, error) {
			return inflightconn.New(ctx, next, *config.InFlightConn, middlewareName)
		}
	}

	// IPWhiteList
	if config.IPWhiteList != nil {
		log.Warn().Msg("IPWhiteList is deprecated, please use IPAllowList instead.")

		middleware = func(next tcp.Handler) (tcp.Handler, error) {
			return ipwhitelist.New(ctx, next, *config.IPWhiteList, middlewareName)
		}
	}

	// IPAllowList
	if config.IPAllowList != nil {
		middleware = func(next tcp.Handler) (tcp.Handler, error) {
			return ipallowlist.New(ctx, next, *config.IPAllowList, middlewareName)
		}
	}

	// Plugin
	if config.Plugin != nil {
		if b.pluginBuilder == nil {
			return nil, fmt.Errorf("plugin middleware %q is not available: no plugin builder", middlewareName)
		}

		pluginType, pluginConfig, err := findPluginConfig(config.Plugin)
		if err != nil {
			return nil, fmt.Errorf("plugin config error: %w", err)
		}

		middleware = func(next tcp.Handler) (tcp.Handler, error) {
			constructor, err := b.pluginBuilder.BuildTCP(pluginType, pluginConfig, middlewareName)
			if err != nil {
				return nil, err
			}
			return constructor(next)
		}
	}

	if middleware == nil {
		return nil, fmt.Errorf("invalid middleware %q configuration: invalid middleware type or middleware does not exist", middlewareName)
	}

	return middleware, nil
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
