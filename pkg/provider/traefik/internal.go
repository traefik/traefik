package traefik

import (
	"context"
	"fmt"
	"math"
	"net"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/log"
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
			Models:      make(map[string]*dynamic.Model),
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
	i.entryPointModels(cfg)
	i.redirection(cfg)

	cfg.HTTP.Services["noop"] = &dynamic.Service{}

	return cfg
}

func (i *Provider) redirection(cfg *dynamic.Configuration) {
	for name, ep := range i.staticCfg.EntryPoints {
		if ep.HTTP.Redirections == nil || ep.HTTP.Redirections.EntryPoint == nil {
			continue
		}

		def := ep.HTTP.Redirections
		rtName := provider.Normalize(name + "-to-" + def.EntryPoint.To)
		mdName := "redirect-" + rtName

		rt := &dynamic.Router{
			Rule:        "HostRegexp(`{host:.+}`)",
			EntryPoints: []string{name},
			Middlewares: []string{mdName},
			Service:     "noop@internal",
		}

		port, err := i.getEntryPointPort(name, def)
		if err != nil {
			log.FromContext(context.Background()).WithField(log.EntryPointName, name).Error(err)
			continue
		}

		cfg.HTTP.Routers[rtName] = rt

		rs := &dynamic.Middleware{
			RedirectScheme: &dynamic.RedirectScheme{
				Scheme:    def.EntryPoint.Scheme,
				Port:      port,
				Permanent: true,
			},
		}

		cfg.HTTP.Middlewares[mdName] = rs
	}
}

func (i *Provider) entryPointModels(cfg *dynamic.Configuration) {
	for name, ep := range i.staticCfg.EntryPoints {
		if len(ep.HTTP.Middlewares) == 0 && ep.HTTP.TLS == nil {
			continue
		}

		m := &dynamic.Model{
			Middlewares: ep.HTTP.Middlewares,
		}

		if ep.HTTP.TLS != nil {
			m.TLS = &dynamic.RouterTLSConfig{
				Options:      ep.HTTP.TLS.Options,
				CertResolver: ep.HTTP.TLS.CertResolver,
				Domains:      ep.HTTP.TLS.Domains,
			}
		}

		cfg.HTTP.Models[name] = m
	}
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

func (i *Provider) getEntryPointPort(name string, def *static.Redirections) (string, error) {
	dst, ok := i.staticCfg.EntryPoints[def.EntryPoint.To]
	if !ok {
		return "", fmt.Errorf("'to' entry point field references a non-existing entry point: %s", name)
	}

	_, port, err := net.SplitHostPort(dst.Address)
	if err != nil {
		return "", fmt.Errorf("invalid entry point %q address %q: %v",
			name, i.staticCfg.EntryPoints[def.EntryPoint.To].Address, err)
	}

	return port, nil
}
