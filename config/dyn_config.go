package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	traefiktls "github.com/containous/traefik/tls"
)

// Router holds the router configuration.
type Router struct {
	EntryPoints []string `json:"entryPoints"`
	Middlewares []string `json:"middlewares,omitempty" toml:",omitempty"`
	Service     string   `json:"service,omitempty" toml:",omitempty"`
	Rule        string   `json:"rule,omitempty" toml:",omitempty"`
	Priority    int      `json:"priority,omitempty" toml:"priority,omitzero"`
}

// LoadBalancerService holds the LoadBalancerService configuration.
type LoadBalancerService struct {
	Stickiness         *Stickiness         `json:"stickiness,omitempty" toml:",omitempty" label:"allowEmpty"`
	Servers            []Server            `json:"servers,omitempty" toml:",omitempty" label-slice-as-struct:"server"`
	Method             string              `json:"method,omitempty" toml:",omitempty"`
	HealthCheck        *HealthCheck        `json:"healthCheck,omitempty" toml:",omitempty"`
	PassHostHeader     bool                `json:"passHostHeader" toml:",omitempty"`
	ResponseForwarding *ResponseForwarding `json:"forwardingResponse,omitempty" toml:",omitempty"`
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
	l.Method = "wrr"
}

// ResponseForwarding holds configuration for the forward of the response.
type ResponseForwarding struct {
	FlushInterval string `json:"flushInterval,omitempty" toml:",omitempty"`
}

// Stickiness holds the stickiness configuration.
type Stickiness struct {
	CookieName string `json:"cookieName,omitempty" toml:",omitempty"`
}

// Server holds the server configuration.
type Server struct {
	URL    string `json:"url" label:"-"`
	Scheme string `toml:"-" json:"-"`
	Port   string `toml:"-" json:"-"`
	Weight int    `json:"weight"`
}

// SetDefaults Default values for a Server.
func (s *Server) SetDefaults() {
	s.Weight = 1
	s.Scheme = "http"
}

// HealthCheck holds the HealthCheck configuration.
type HealthCheck struct {
	Scheme string `json:"scheme,omitempty" toml:",omitempty"`
	Path   string `json:"path,omitempty" toml:",omitempty"`
	Port   int    `json:"port,omitempty" toml:",omitempty,omitzero"`
	// FIXME change string to parse.Duration
	Interval string `json:"interval,omitempty" toml:",omitempty"`
	// FIXME change string to parse.Duration
	Timeout  string            `json:"timeout,omitempty" toml:",omitempty"`
	Hostname string            `json:"hostname,omitempty" toml:",omitempty"`
	Headers  map[string]string `json:"headers,omitempty" toml:",omitempty"`
}

// ClientTLS holds the TLS specific configurations as client
// CA, Cert and Key can be either path or file contents.
type ClientTLS struct {
	CA                 string `description:"TLS CA" json:"ca,omitempty"`
	CAOptional         bool   `description:"TLS CA.Optional" json:"caOptional,omitempty"`
	Cert               string `description:"TLS cert" json:"cert,omitempty"`
	Key                string `description:"TLS key" json:"key,omitempty"`
	InsecureSkipVerify bool   `description:"TLS insecure skip verify" json:"insecureSkipVerify,omitempty"`
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

// Configurations is for currentConfigurations Map.
type Configurations map[string]*Configuration

// Configuration FIXME better name?
type Configuration struct {
	Routers     map[string]*Router          `json:"routers,omitempty" toml:",omitempty"`
	Middlewares map[string]*Middleware      `json:"middlewares,omitempty" toml:",omitempty"`
	Services    map[string]*Service         `json:"services,omitempty" toml:",omitempty"`
	TLS         []*traefiktls.Configuration `json:"-" label:"-"`
}

// Service holds a service configuration (can only be of one type at the same time).
type Service struct {
	LoadBalancer *LoadBalancerService `json:"loadbalancer,omitempty" toml:",omitempty,omitzero"`
}
