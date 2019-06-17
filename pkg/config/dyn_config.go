package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	traefiktls "github.com/containous/traefik/pkg/tls"
)

// Router holds the router configuration.
type Router struct {
	EntryPoints []string         `json:"entryPoints"`
	Middlewares []string         `json:"middlewares,omitempty" toml:",omitempty"`
	Service     string           `json:"service,omitempty" toml:",omitempty"`
	Rule        string           `json:"rule,omitempty" toml:",omitempty"`
	Priority    int              `json:"priority,omitempty" toml:"priority,omitzero"`
	TLS         *RouterTLSConfig `json:"tls,omitempty" toml:"tls,omitzero" label:"allowEmpty"`
}

// RouterTLSConfig holds the TLS configuration for a router
type RouterTLSConfig struct{}

// TCPRouter holds the router configuration.
type TCPRouter struct {
	EntryPoints []string            `json:"entryPoints"`
	Service     string              `json:"service,omitempty" toml:",omitempty"`
	Rule        string              `json:"rule,omitempty" toml:",omitempty"`
	TLS         *RouterTCPTLSConfig `json:"tls,omitempty" toml:"tls,omitzero" label:"allowEmpty"`
}

// RouterTCPTLSConfig holds the TLS configuration for a router
type RouterTCPTLSConfig struct {
	Passthrough bool `json:"passthrough" toml:"passthrough,omitzero"`
}

// LoadBalancerService holds the LoadBalancerService configuration.
type LoadBalancerService struct {
	Stickiness         *Stickiness         `json:"stickiness,omitempty" toml:",omitempty" label:"allowEmpty"`
	Servers            []Server            `json:"servers,omitempty" toml:",omitempty" label-slice-as-struct:"server"`
	HealthCheck        *HealthCheck        `json:"healthCheck,omitempty" toml:",omitempty"`
	PassHostHeader     bool                `json:"passHostHeader" toml:",omitempty"`
	ResponseForwarding *ResponseForwarding `json:"forwardingResponse,omitempty" toml:",omitempty"`
}

// TCPLoadBalancerService holds the LoadBalancerService configuration.
type TCPLoadBalancerService struct {
	Servers []TCPServer `json:"servers,omitempty" toml:",omitempty" label-slice-as-struct:"server"`
}

// Mergeable tells if the given service is mergeable.
func (l *TCPLoadBalancerService) Mergeable(loadBalancer *TCPLoadBalancerService) bool {
	savedServers := l.Servers
	defer func() {
		l.Servers = savedServers
	}()
	l.Servers = nil

	savedServersLB := loadBalancer.Servers
	defer func() {
		loadBalancer.Servers = savedServersLB
	}()
	loadBalancer.Servers = nil

	return reflect.DeepEqual(l, loadBalancer)
}

// Mergeable tells if the given service is mergeable.
func (l *LoadBalancerService) Mergeable(loadBalancer *LoadBalancerService) bool {
	savedServers := l.Servers
	defer func() {
		l.Servers = savedServers
	}()
	l.Servers = nil

	savedServersLB := loadBalancer.Servers
	defer func() {
		loadBalancer.Servers = savedServersLB
	}()
	loadBalancer.Servers = nil

	return reflect.DeepEqual(l, loadBalancer)
}

// SetDefaults Default values for a LoadBalancerService.
func (l *LoadBalancerService) SetDefaults() {
	l.PassHostHeader = true
}

// ResponseForwarding holds configuration for the forward of the response.
type ResponseForwarding struct {
	FlushInterval string `json:"flushInterval,omitempty" toml:",omitempty"`
}

// Stickiness holds the stickiness configuration.
type Stickiness struct {
	CookieName     string `json:"cookieName,omitempty" toml:",omitempty"`
	SecureCookie   bool   `json:"secureCookie,omitempty" toml:",omitempty"`
	HTTPOnlyCookie bool   `json:"httpOnlyCookie,omitempty" toml:",omitempty"`
}

// Server holds the server configuration.
type Server struct {
	URL    string `json:"url" label:"-"`
	Scheme string `toml:"-" json:"-"`
	Port   string `toml:"-" json:"-"`
}

// TCPServer holds a TCP Server configuration
type TCPServer struct {
	Address string `json:"address" label:"-"`
	Port    string `toml:"-" json:"-"`
}

// SetDefaults Default values for a Server.
func (s *Server) SetDefaults() {
	s.Scheme = "http"
}

// HealthCheck holds the HealthCheck configuration.
type HealthCheck struct {
	Scheme string `json:"scheme,omitempty" toml:",omitempty"`
	Path   string `json:"path,omitempty" toml:",omitempty"`
	Port   int    `json:"port,omitempty" toml:",omitempty,omitzero"`
	// FIXME change string to types.Duration
	Interval string `json:"interval,omitempty" toml:",omitempty"`
	// FIXME change string to types.Duration
	Timeout  string            `json:"timeout,omitempty" toml:",omitempty"`
	Hostname string            `json:"hostname,omitempty" toml:",omitempty"`
	Headers  map[string]string `json:"headers,omitempty" toml:",omitempty"`
}

// CreateTLSConfig creates a TLS config from ClientTLS structures.
func (clientTLS *ClientTLS) CreateTLSConfig() (*tls.Config, error) {
	if clientTLS == nil {
		return nil, nil
	}

	var err error
	caPool := x509.NewCertPool()
	clientAuth := tls.NoClientCert
	if clientTLS.CA != "" {
		var ca []byte
		if _, errCA := os.Stat(clientTLS.CA); errCA == nil {
			ca, err = ioutil.ReadFile(clientTLS.CA)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA. %s", err)
			}
		} else {
			ca = []byte(clientTLS.CA)
		}

		if !caPool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("failed to parse CA")
		}

		if clientTLS.CAOptional {
			clientAuth = tls.VerifyClientCertIfGiven
		} else {
			clientAuth = tls.RequireAndVerifyClientCert
		}
	}

	cert := tls.Certificate{}
	_, errKeyIsFile := os.Stat(clientTLS.Key)

	if !clientTLS.InsecureSkipVerify && (len(clientTLS.Cert) == 0 || len(clientTLS.Key) == 0) {
		return nil, fmt.Errorf("TLS Certificate or Key file must be set when TLS configuration is created")
	}

	if len(clientTLS.Cert) > 0 && len(clientTLS.Key) > 0 {
		if _, errCertIsFile := os.Stat(clientTLS.Cert); errCertIsFile == nil {
			if errKeyIsFile == nil {
				cert, err = tls.LoadX509KeyPair(clientTLS.Cert, clientTLS.Key)
				if err != nil {
					return nil, fmt.Errorf("failed to load TLS keypair: %v", err)
				}
			} else {
				return nil, fmt.Errorf("tls cert is a file, but tls key is not")
			}
		} else {
			if errKeyIsFile != nil {
				cert, err = tls.X509KeyPair([]byte(clientTLS.Cert), []byte(clientTLS.Key))
				if err != nil {
					return nil, fmt.Errorf("failed to load TLS keypair: %v", err)

				}
			} else {
				return nil, fmt.Errorf("TLS key is a file, but tls cert is not")
			}
		}
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caPool,
		InsecureSkipVerify: clientTLS.InsecureSkipVerify,
		ClientAuth:         clientAuth,
	}, nil
}

// Message holds configuration information exchanged between parts of traefik.
type Message struct {
	ProviderName  string
	Configuration *Configuration
}

// Configuration is the root of the dynamic configuration
type Configuration struct {
	HTTP       *HTTPConfiguration
	TCP        *TCPConfiguration
	TLS        []*traefiktls.Configuration `json:"-" label:"-"`
	TLSOptions map[string]traefiktls.TLS
	TLSStores  map[string]traefiktls.Store
}

// Configurations is for currentConfigurations Map.
type Configurations map[string]*Configuration

// HTTPConfiguration FIXME better name?
type HTTPConfiguration struct {
	Routers     map[string]*Router     `json:"routers,omitempty" toml:",omitempty"`
	Middlewares map[string]*Middleware `json:"middlewares,omitempty" toml:",omitempty"`
	Services    map[string]*Service    `json:"services,omitempty" toml:",omitempty"`
}

// TCPConfiguration FIXME better name?
type TCPConfiguration struct {
	Routers  map[string]*TCPRouter  `json:"routers,omitempty" toml:",omitempty"`
	Services map[string]*TCPService `json:"services,omitempty" toml:",omitempty"`
}

// Service holds a service configuration (can only be of one type at the same time).
type Service struct {
	LoadBalancer *LoadBalancerService `json:"loadbalancer,omitempty" toml:",omitempty,omitzero"`
}

// TCPService holds a tcp service configuration (can only be of one type at the same time).
type TCPService struct {
	LoadBalancer *TCPLoadBalancerService `json:"loadbalancer,omitempty" toml:",omitempty,omitzero"`
}
