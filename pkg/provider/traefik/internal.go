package traefik

import (
	"math"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/tls"
)

var _ provider.Provider = (*Provider)(nil)

// Provider is a provider.Provider implementation that provides the internal routers.
type Provider struct {
	staticCfg static.Configuration
}

// New creates a new instance of the internal provider.
func New(staticCfg static.Configuration) *Provider {
	return &Provider{staticCfg: staticCfg}
}

// Provide allows the provider to provide configurations to traefik using the given configuration channel.
func (i *Provider) Provide(configurationChan chan<- dynamic.Message, _ *safe.Pool) error {
	configurationChan <- dynamic.Message{
		ProviderName:  "internal",
		Configuration: i.createConfiguration(),
	}

	return nil
}

// Init the provider.
func (i *Provider) Init() error {
	return nil
}

func (i *Provider) createConfiguration() *dynamic.Configuration {
	cfg := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     make(map[string]*dynamic.Router),
			Middlewares: make(map[string]*dynamic.Middleware),
			Services:    make(map[string]*dynamic.Service),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  make(map[string]*dynamic.TCPRouter),
			Services: make(map[string]*dynamic.TCPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
	}

	i.apiConfiguration(cfg)
	i.pingConfiguration(cfg)
	i.restConfiguration(cfg)
	i.prometheusConfiguration(cfg)

	return cfg
}

func (i *Provider) apiConfiguration(cfg *dynamic.Configuration) {
	if i.staticCfg.API == nil {
		return
	}

	if i.staticCfg.API.Insecure {
		cfg.HTTP.Routers["api"] = &dynamic.Router{
			EntryPoints: []string{"traefik"},
			Service:     "api@internal",
			Priority:    math.MaxInt32 - 1,
			Rule:        "PathPrefix(`/api`)",
		}

		if i.staticCfg.API.Dashboard {
			cfg.HTTP.Routers["dashboard"] = &dynamic.Router{
				EntryPoints: []string{"traefik"},
				Service:     "dashboard@internal",
				Priority:    math.MaxInt32 - 2,
				Rule:        "PathPrefix(`/`)",
				Middlewares: []string{"dashboard_redirect@internal", "dashboard_stripprefix@internal"},
			}

			cfg.HTTP.Middlewares["dashboard_redirect"] = &dynamic.Middleware{
				RedirectRegex: &dynamic.RedirectRegex{
					Regex:       `^(http:\/\/[^:\/]+(:\d+)?)\/$`,
					Replacement: "${1}/dashboard/",
					Permanent:   true,
				},
			}
			cfg.HTTP.Middlewares["dashboard_stripprefix"] = &dynamic.Middleware{
				StripPrefix: &dynamic.StripPrefix{Prefixes: []string{"/dashboard/", "/dashboard"}},
			}
		}

		if i.staticCfg.API.Debug {
			cfg.HTTP.Routers["debug"] = &dynamic.Router{
				EntryPoints: []string{"traefik"},
				Service:     "api@internal",
				Priority:    math.MaxInt32 - 1,
				Rule:        "PathPrefix(`/debug`)",
			}
		}
	}

	cfg.HTTP.Services["api"] = &dynamic.Service{}

	if i.staticCfg.API.Dashboard {
		cfg.HTTP.Services["dashboard"] = &dynamic.Service{}
	}
}

func (i *Provider) pingConfiguration(cfg *dynamic.Configuration) {
	if i.staticCfg.Ping == nil {
		return
	}

	if !i.staticCfg.Ping.ManualRouting {
		cfg.HTTP.Routers["ping"] = &dynamic.Router{
			EntryPoints: []string{i.staticCfg.Ping.EntryPoint},
			Service:     "ping@internal",
			Priority:    math.MaxInt32,
			Rule:        "PathPrefix(`/ping`)",
		}
	}

	cfg.HTTP.Services["ping"] = &dynamic.Service{}
}

func (i *Provider) restConfiguration(cfg *dynamic.Configuration) {
	if i.staticCfg.Providers == nil || i.staticCfg.Providers.Rest == nil {
		return
	}

	if i.staticCfg.Providers.Rest.Insecure {
		cfg.HTTP.Routers["rest"] = &dynamic.Router{
			EntryPoints: []string{"traefik"},
			Service:     "rest@internal",
			Priority:    math.MaxInt32,
			Rule:        "PathPrefix(`/api/providers`)",
		}
	}

	cfg.HTTP.Services["rest"] = &dynamic.Service{}
}

func (i *Provider) prometheusConfiguration(cfg *dynamic.Configuration) {
	if i.staticCfg.Metrics == nil || i.staticCfg.Metrics.Prometheus == nil {
		return
	}

	if !i.staticCfg.Metrics.Prometheus.ManualRouting {
		cfg.HTTP.Routers["prometheus"] = &dynamic.Router{
			EntryPoints: []string{i.staticCfg.Metrics.Prometheus.EntryPoint},
			Service:     "prometheus@internal",
			Priority:    math.MaxInt32,
			Rule:        "PathPrefix(`/metrics`)",
		}
	}

	cfg.HTTP.Services["prometheus"] = &dynamic.Service{}
}
