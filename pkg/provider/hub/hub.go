package hub

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/safe"
	ttls "github.com/traefik/traefik/v2/pkg/tls"
)

var _ provider.Provider = (*Provider)(nil)

// DefaultEntryPointName the name of the default internal entry point.
const DefaultEntryPointName = "traefik-hub"

// Provider holds configurations of the provider.
type Provider struct {
	EntryPoint string `description:"Entrypoint that exposes data for Traefik Hub." json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty"`
	Insecure   bool   `description:"Allows the Hub provider to run over an insecure connection for testing purposes." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty"`
	TLS        *TLS   `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`

	server *http.Server
}

// TLS holds TLS configuration to use mTLS communication between Traefik and Hub Agent.
type TLS struct {
	CA   ttls.FileOrContent `description:"Certificate authority to use for securing communication with the Agent." json:"ca,omitempty" toml:"ca,omitempty" yaml:"ca,omitempty"`
	Cert ttls.FileOrContent `description:"Certificate to use for securing communication with the Agent." json:"cert,omitempty" toml:"cert,omitempty" yaml:"cert,omitempty"`
	Key  ttls.FileOrContent `description:"Key to use for securing communication with the Agent." json:"key,omitempty" toml:"key,omitempty" yaml:"key,omitempty"`
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

	client, err := createAgentClient(p.TLS)
	if err != nil {
		return fmt.Errorf("create Hub Agent HTTP client: %w", err)
	}

	p.server = &http.Server{Handler: newHandler(p.EntryPoint, port, configurationChan, p.TLS, client)}

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

	exposeAPIAndMetrics(configurationChan, p.EntryPoint, port, p.TLS)

	return nil
}

func exposeAPIAndMetrics(cfgChan chan<- dynamic.Message, ep string, port int, tlsCfg *TLS) {
	cfg := emptyDynamicConfiguration()

	patchDynamicConfiguration(cfg, ep, port, tlsCfg)

	cfgChan <- dynamic.Message{ProviderName: "hub", Configuration: cfg}
}

func patchDynamicConfiguration(cfg *dynamic.Configuration, ep string, port int, tlsCfg *TLS) {
	cfg.HTTP.Routers["traefik-hub-agent-api"] = &dynamic.Router{
		EntryPoints: []string{ep},
		Service:     "api@internal",
		Rule:        "Host(`proxy.traefik`) && PathPrefix(`/api`)",
	}
	cfg.HTTP.Routers["traefik-hub-agent-metrics"] = &dynamic.Router{
		EntryPoints: []string{ep},
		Service:     "prometheus@internal",
		Rule:        "Host(`proxy.traefik`) && PathPrefix(`/metrics`)",
	}

	cfg.HTTP.Routers["traefik-hub-agent-service"] = &dynamic.Router{
		EntryPoints: []string{ep},
		Service:     "traefik-hub-agent-service",
		Rule:        "Host(`proxy.traefik`) && (PathPrefix(`/config`) || PathPrefix(`/discover-ip`) || PathPrefix(`/state`))",
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

	if tlsCfg != nil {
		cfg.TLS.Options["traefik-hub"] = ttls.Options{
			ClientAuth: ttls.ClientAuth{
				CAFiles:        []ttls.FileOrContent{tlsCfg.CA},
				ClientAuthType: "RequireAndVerifyClientCert",
			},
			SniStrict:  true,
			MinVersion: "VersionTLS13",
		}

		cfg.TLS.Certificates = append(cfg.TLS.Certificates, &ttls.CertAndStores{
			Certificate: ttls.Certificate{
				CertFile: tlsCfg.Cert,
				KeyFile:  tlsCfg.Key,
			},
		})
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
			Stores:  make(map[string]ttls.Store),
			Options: make(map[string]ttls.Options),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
	}
}

func createAgentClient(tlsCfg *TLS) (http.Client, error) {
	var client http.Client
	if tlsCfg != nil {
		roots := x509.NewCertPool()
		caContent, err := tlsCfg.CA.Read()
		if err != nil {
			return client, fmt.Errorf("read CA: %w", err)
		}

		roots.AppendCertsFromPEM(caContent)
		certContent, err := tlsCfg.Cert.Read()
		if err != nil {
			return client, fmt.Errorf("read Cert: %w", err)
		}
		keyContent, err := tlsCfg.Key.Read()
		if err != nil {
			return client, fmt.Errorf("read Key: %w", err)
		}

		certificate, err := tls.X509KeyPair(certContent, keyContent)
		if err != nil {
			return client, fmt.Errorf("create key pair: %w", err)
		}

		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      roots,
				Certificates: []tls.Certificate{certificate},
				ServerName:   "agent.traefik",
				ClientAuth:   tls.RequireAndVerifyClientCert,
				MinVersion:   tls.VersionTLS13,
			},
		}
	}

	return client, nil
}
