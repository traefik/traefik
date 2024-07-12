package traefik

import (
	"context"
	"fmt"
	"math"
	"net"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
)

const defaultInternalEntryPointName = "traefik"

var _ provider.Provider = (*Provider)(nil)

// Provider is a provider.Provider implementation that provides the internal routers.
type Provider struct {
	staticCfg static.Configuration
}

// New creates a new instance of the internal provider.
func New(staticCfg static.Configuration) *Provider {
	return &Provider{staticCfg: staticCfg}
}

// ThrottleDuration returns the throttle duration.
func (i Provider) ThrottleDuration() time.Duration {
	return 0
}

// Provide allows the provider to provide configurations to traefik using the given configuration channel.
func (i *Provider) Provide(configurationChan chan<- dynamic.Message, _ *safe.Pool) error {
	ctx := log.With().Str(logs.ProviderName, "internal").Logger().WithContext(context.Background())

	configurationChan <- dynamic.Message{
		ProviderName:  "internal",
		Configuration: i.createConfiguration(ctx),
	}

	return nil
}

// Init the provider.
func (i *Provider) Init() error {
	return nil
}

func (i *Provider) createConfiguration(ctx context.Context) *dynamic.Configuration {
	cfg := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			Models:            make(map[string]*dynamic.Model),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			Models:            make(map[string]*dynamic.TCPModel),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
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
	i.redirection(ctx, cfg)
	i.serverTransport(cfg)
	i.serverTransportTCP(cfg)

	i.acme(cfg)

	cfg.HTTP.Services["noop"] = &dynamic.Service{}

	return cfg
}

func (i *Provider) acme(cfg *dynamic.Configuration) {
	var eps []string

	uniq := map[string]struct{}{}
	for _, resolver := range i.staticCfg.CertificatesResolvers {
		if resolver.ACME != nil && resolver.ACME.HTTPChallenge != nil && resolver.ACME.HTTPChallenge.EntryPoint != "" {
			if _, ok := uniq[resolver.ACME.HTTPChallenge.EntryPoint]; !ok {
				eps = append(eps, resolver.ACME.HTTPChallenge.EntryPoint)
				uniq[resolver.ACME.HTTPChallenge.EntryPoint] = struct{}{}
			}
		}
	}

	if len(eps) > 0 {
		rt := &dynamic.Router{
			Rule:        "PathPrefix(`/.well-known/acme-challenge/`)",
			RuleSyntax:  "v3",
			EntryPoints: eps,
			Service:     "acme-http@internal",
			Priority:    math.MaxInt,
		}

		cfg.HTTP.Routers["acme-http"] = rt
		cfg.HTTP.Services["acme-http"] = &dynamic.Service{}
	}
}

func (i *Provider) redirection(ctx context.Context, cfg *dynamic.Configuration) {
	for name, ep := range i.staticCfg.EntryPoints {
		if ep.HTTP.Redirections == nil {
			continue
		}

		logger := log.Ctx(ctx).With().Str(logs.EntryPointName, name).Logger()

		def := ep.HTTP.Redirections
		if def.EntryPoint == nil || def.EntryPoint.To == "" {
			logger.Error().Msg("Unable to create redirection: the entry point or the port is missing")
			continue
		}

		port, err := i.getRedirectPort(name, def)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		rtName := provider.Normalize(name + "-to-" + def.EntryPoint.To)
		mdName := "redirect-" + rtName

		rt := &dynamic.Router{
			Rule:        "HostRegexp(`^.+$`)",
			RuleSyntax:  "v3",
			EntryPoints: []string{name},
			Middlewares: []string{mdName},
			Service:     "noop@internal",
			Priority:    def.EntryPoint.Priority,
		}

		cfg.HTTP.Routers[rtName] = rt

		rs := &dynamic.Middleware{
			RedirectScheme: &dynamic.RedirectScheme{
				Scheme:    def.EntryPoint.Scheme,
				Port:      port,
				Permanent: def.EntryPoint.Permanent,
			},
		}

		cfg.HTTP.Middlewares[mdName] = rs
	}
}

func (i *Provider) getRedirectPort(name string, def *static.Redirections) (string, error) {
	exp := regexp.MustCompile(`^:(\d+)$`)

	if exp.MatchString(def.EntryPoint.To) {
		_, port, err := net.SplitHostPort(def.EntryPoint.To)
		if err != nil {
			return "", fmt.Errorf("invalid port value: %w", err)
		}

		return port, nil
	}

	return i.getEntryPointPort(name, def)
}

func (i *Provider) getEntryPointPort(name string, def *static.Redirections) (string, error) {
	dst, ok := i.staticCfg.EntryPoints[def.EntryPoint.To]
	if !ok {
		return "", fmt.Errorf("'to' entry point field references a non-existing entry point: %s", def.EntryPoint.To)
	}

	_, port, err := net.SplitHostPort(dst.GetAddress())
	if err != nil {
		return "", fmt.Errorf("invalid entry point %q address %q: %w",
			name, i.staticCfg.EntryPoints[def.EntryPoint.To].Address, err)
	}

	return port, nil
}

func (i *Provider) entryPointModels(cfg *dynamic.Configuration) {
	defaultRuleSyntax := ""
	if i.staticCfg.Core != nil && i.staticCfg.Core.DefaultRuleSyntax != "" {
		defaultRuleSyntax = i.staticCfg.Core.DefaultRuleSyntax
	}

	for name, ep := range i.staticCfg.EntryPoints {
		if len(ep.HTTP.Middlewares) == 0 && ep.HTTP.TLS == nil && defaultRuleSyntax == "" {
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

		m.DefaultRuleSyntax = defaultRuleSyntax

		cfg.HTTP.Models[name] = m

		if cfg.TCP == nil {
			continue
		}

		mTCP := &dynamic.TCPModel{
			DefaultRuleSyntax: defaultRuleSyntax,
		}

		cfg.TCP.Models[name] = mTCP
	}
}

func (i *Provider) apiConfiguration(cfg *dynamic.Configuration) {
	if i.staticCfg.API == nil {
		return
	}

	if i.staticCfg.API.Insecure {
		cfg.HTTP.Routers["api"] = &dynamic.Router{
			EntryPoints: []string{defaultInternalEntryPointName},
			Service:     "api@internal",
			Priority:    math.MaxInt - 1,
			Rule:        "PathPrefix(`/api`)",
			RuleSyntax:  "v3",
		}

		if i.staticCfg.API.Dashboard {
			cfg.HTTP.Routers["dashboard"] = &dynamic.Router{
				EntryPoints: []string{defaultInternalEntryPointName},
				Service:     "dashboard@internal",
				Priority:    math.MaxInt - 2,
				Rule:        "PathPrefix(`/`)",
				RuleSyntax:  "v3",
				Middlewares: []string{"dashboard_redirect@internal", "dashboard_stripprefix@internal"},
			}

			cfg.HTTP.Middlewares["dashboard_redirect"] = &dynamic.Middleware{
				RedirectRegex: &dynamic.RedirectRegex{
					Regex:       `^(http:\/\/(\[[\w:.]+\]|[\w\._-]+)(:\d+)?)\/$`,
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
				EntryPoints: []string{defaultInternalEntryPointName},
				Service:     "api@internal",
				Priority:    math.MaxInt - 1,
				Rule:        "PathPrefix(`/debug`)",
				RuleSyntax:  "v3",
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
			Priority:    math.MaxInt,
			Rule:        "PathPrefix(`/ping`)",
			RuleSyntax:  "v3",
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
			EntryPoints: []string{defaultInternalEntryPointName},
			Service:     "rest@internal",
			Priority:    math.MaxInt,
			Rule:        "PathPrefix(`/api/providers`)",
			RuleSyntax:  "v3",
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
			Priority:    math.MaxInt,
			Rule:        "PathPrefix(`/metrics`)",
			RuleSyntax:  "v3",
		}
	}

	cfg.HTTP.Services["prometheus"] = &dynamic.Service{}
}

func (i *Provider) serverTransport(cfg *dynamic.Configuration) {
	if i.staticCfg.ServersTransport == nil {
		return
	}

	st := &dynamic.ServersTransport{
		InsecureSkipVerify:  i.staticCfg.ServersTransport.InsecureSkipVerify,
		RootCAs:             i.staticCfg.ServersTransport.RootCAs,
		MaxIdleConnsPerHost: i.staticCfg.ServersTransport.MaxIdleConnsPerHost,
	}

	if i.staticCfg.ServersTransport.Spiffe != nil {
		st.Spiffe = &dynamic.Spiffe{
			IDs:         i.staticCfg.ServersTransport.Spiffe.IDs,
			TrustDomain: i.staticCfg.ServersTransport.Spiffe.TrustDomain,
		}
	}

	if i.staticCfg.ServersTransport.ForwardingTimeouts != nil {
		st.ForwardingTimeouts = &dynamic.ForwardingTimeouts{
			DialTimeout:           i.staticCfg.ServersTransport.ForwardingTimeouts.DialTimeout,
			ResponseHeaderTimeout: i.staticCfg.ServersTransport.ForwardingTimeouts.ResponseHeaderTimeout,
			IdleConnTimeout:       i.staticCfg.ServersTransport.ForwardingTimeouts.IdleConnTimeout,
		}
	}

	cfg.HTTP.ServersTransports["default"] = st
}

func (i *Provider) serverTransportTCP(cfg *dynamic.Configuration) {
	if i.staticCfg.TCPServersTransport == nil {
		return
	}

	st := &dynamic.TCPServersTransport{
		DialTimeout:   i.staticCfg.TCPServersTransport.DialTimeout,
		DialKeepAlive: i.staticCfg.TCPServersTransport.DialKeepAlive,
	}

	if i.staticCfg.TCPServersTransport.TLS != nil {
		st.TLS = &dynamic.TLSClientConfig{
			InsecureSkipVerify: i.staticCfg.TCPServersTransport.TLS.InsecureSkipVerify,
			RootCAs:            i.staticCfg.TCPServersTransport.TLS.RootCAs,
		}

		if i.staticCfg.TCPServersTransport.TLS.Spiffe != nil {
			st.TLS.Spiffe = &dynamic.Spiffe{
				IDs:         i.staticCfg.ServersTransport.Spiffe.IDs,
				TrustDomain: i.staticCfg.ServersTransport.Spiffe.TrustDomain,
			}
		}
	}

	cfg.TCP.ServersTransports["default"] = st
}
