package hub

import (
	"fmt"
	"net"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/tls"
)

var _ provider.Provider = (*Provider)(nil)

// DefaultEntryPointName the name of the default internal entry point.
const DefaultEntryPointName = "traefik-hub"

// Provider holds configurations of the provider.
type Provider struct {
	EntryPoint string `description:"Entrypoint that exposes data for Traefik Hub." json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty"`

	server *http.Server
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.EntryPoint = DefaultEntryPointName
}

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Provide allows the hub provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, _ *safe.Pool) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	p.server = &http.Server{Handler: newHandler(p.EntryPoint, port, configurationChan)}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.WithoutContext().Errorf("Panic recovered: %v", err)
			}
		}()

		if err = p.server.Serve(listener); err != nil {
			log.WithoutContext().Errorf("Unexpected error while running server: %v", err)
			return
		}
	}()

	exposeAPIAndMetrics(configurationChan, p.EntryPoint, port)

	return nil
}

func exposeAPIAndMetrics(cfgChan chan<- dynamic.Message, ep string, port int) {
	cfg := emptyDynamicConfiguration()

	patchDynamicConfiguration(cfg, ep, port)

	cfgChan <- dynamic.Message{ProviderName: "hub", Configuration: cfg}
}

func patchDynamicConfiguration(cfg *dynamic.Configuration, ep string, port int) {
	cfg.HTTP.Routers["traefik-hub-agent-api"] = &dynamic.Router{
		EntryPoints: []string{ep},
		Service:     "api@internal",
		Rule:        "PathPrefix(`/api`)",
	}
	cfg.HTTP.Routers["traefik-hub-agent-metrics"] = &dynamic.Router{
		EntryPoints: []string{ep},
		Service:     "prometheus@internal",
		Rule:        "PathPrefix(`/metrics`)",
	}

	cfg.HTTP.Routers["traefik-hub-agent-service"] = &dynamic.Router{
		EntryPoints: []string{ep},
		Service:     "traefik-hub-agent-service",
		Rule:        "PathPrefix(`/config`) || PathPrefix(`/discover-ip`) || PathPrefix(`/state`)",
	}

	cfg.HTTP.Services["traefik-hub-agent-service"] = &dynamic.Service{
		LoadBalancer: &dynamic.ServersLoadBalancer{
			Servers: []dynamic.Server{
				{
					URL: fmt.Sprintf("http://127.0.0.1:%d", port),
				},
			},
		},
	}
}

func emptyDynamicConfiguration() *dynamic.Configuration {
	return &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  make(map[string]*dynamic.TCPRouter),
			Services: make(map[string]*dynamic.TCPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
	}
}
