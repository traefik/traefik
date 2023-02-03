package hub

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
	ttls "github.com/traefik/traefik/v3/pkg/tls"
)

var _ provider.Provider = (*Provider)(nil)

// Entrypoints created for Hub.
const (
	APIEntrypoint    = "traefikhub-api"
	TunnelEntrypoint = "traefikhub-tunl"
)

// Provider holds configurations of the provider.
type Provider struct {
	TLS *TLS `description:"TLS configuration for mTLS communication between Traefik and Hub Agent." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`

	server *http.Server
}

// TLS configures the mTLS connection between Traefik Proxy and the Traefik Hub Agent.
type TLS struct {
	Insecure bool               `description:"Enables an insecure TLS connection that uses default credentials, and which has no peer authentication between Traefik Proxy and the Traefik Hub Agent." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	CA       ttls.FileOrContent `description:"The certificate authority authenticates the Traefik Hub Agent certificate." json:"ca,omitempty" toml:"ca,omitempty" yaml:"ca,omitempty" loggable:"false"`
	Cert     ttls.FileOrContent `description:"The TLS certificate for Traefik Proxy as a TLS client." json:"cert,omitempty" toml:"cert,omitempty" yaml:"cert,omitempty" loggable:"false"`
	Key      ttls.FileOrContent `description:"The TLS key for Traefik Proxy as a TLS client." json:"key,omitempty" toml:"key,omitempty" yaml:"key,omitempty" loggable:"false"`
}

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Provide allows the hub provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, _ *safe.Pool) error {
	if p.TLS == nil {
		return nil
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listener: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	client, err := createAgentClient(p.TLS)
	if err != nil {
		return fmt.Errorf("creating Hub Agent HTTP client: %w", err)
	}

	p.server = &http.Server{Handler: newHandler(APIEntrypoint, port, configurationChan, p.TLS, client)}

	// TODO: this is going to be leaky (because no context to make it terminate)
	// if/when Provide lifecycle differs with Traefik lifecycle.
	go func() {
		if err = p.server.Serve(listener); err != nil {
			log.Error().Str(logs.ProviderName, "hub").Err(err).Msg("Unexpected error while running server")
			return
		}
	}()

	exposeAPIAndMetrics(configurationChan, APIEntrypoint, port, p.TLS)

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

	if tlsCfg == nil {
		return
	}

	if tlsCfg.Insecure {
		cfg.TLS.Options["traefik-hub"] = ttls.Options{
			MinVersion: "VersionTLS13",
		}

		return
	}

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

func emptyDynamicConfiguration() *dynamic.Configuration {
	return &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
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
	if tlsCfg.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS13,
			},
		}

		return client, nil
	}

	caContent, err := tlsCfg.CA.Read()
	if err != nil {
		return client, fmt.Errorf("reading CA: %w", err)
	}

	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(caContent); !ok {
		return client, errors.New("appending CA error")
	}

	certContent, err := tlsCfg.Cert.Read()
	if err != nil {
		return client, fmt.Errorf("reading Cert: %w", err)
	}
	keyContent, err := tlsCfg.Key.Read()
	if err != nil {
		return client, fmt.Errorf("reading Key: %w", err)
	}

	certificate, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return client, fmt.Errorf("creating key pair: %w", err)
	}

	// mTLS
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      roots,
			Certificates: []tls.Certificate{certificate},
			ServerName:   "agent.traefik",
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS13,
		},
	}

	return client, nil
}
