package dynamic

import (
	"reflect"
	"time"

	ptypes "github.com/traefik/paerser/types"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// +k8s:deepcopy-gen=true

// TCPConfiguration contains all the TCP configuration parameters.
type TCPConfiguration struct {
	Routers           map[string]*TCPRouter           `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty" export:"true"`
	Services          map[string]*TCPService          `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty" export:"true"`
	Middlewares       map[string]*TCPMiddleware       `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	Models            map[string]*TCPModel            `json:"-" toml:"-" yaml:"-" label:"-" file:"-" kv:"-" export:"true"`
	ServersTransports map[string]*TCPServersTransport `json:"serversTransports,omitempty" toml:"serversTransports,omitempty" yaml:"serversTransports,omitempty" label:"-" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPModel is a set of default router's values.
type TCPModel struct {
	DefaultRuleSyntax string `json:"-" toml:"-" yaml:"-" label:"-" file:"-" kv:"-" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPService holds a tcp service configuration (can only be of one type at the same time).
type TCPService struct {
	LoadBalancer *TCPServersLoadBalancer `json:"loadBalancer,omitempty" toml:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty" export:"true"`
	Weighted     *TCPWeightedRoundRobin  `json:"weighted,omitempty" toml:"weighted,omitempty" yaml:"weighted,omitempty" label:"-" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPWeightedRoundRobin is a weighted round robin tcp load-balancer of services.
type TCPWeightedRoundRobin struct {
	Services []TCPWRRService `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPWRRService is a reference to a tcp service load-balanced with weighted round robin.
type TCPWRRService struct {
	Name   string `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty" export:"true"`
	Weight *int   `json:"weight,omitempty" toml:"weight,omitempty" yaml:"weight,omitempty" export:"true"`
}

// SetDefaults Default values for a TCPWRRService.
func (w *TCPWRRService) SetDefaults() {
	defaultWeight := 1
	w.Weight = &defaultWeight
}

// +k8s:deepcopy-gen=true

// TCPRouter holds the router configuration.
type TCPRouter struct {
	EntryPoints []string `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	Middlewares []string `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	Service     string   `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	Rule        string   `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
	// Deprecated: Please do not use this field and rewrite the router rules to use the v3 syntax.
	RuleSyntax string              `json:"ruleSyntax,omitempty" toml:"ruleSyntax,omitempty" yaml:"ruleSyntax,omitempty" export:"true"`
	Priority   int                 `json:"priority,omitempty" toml:"priority,omitempty,omitzero" yaml:"priority,omitempty" export:"true"`
	TLS        *RouterTCPTLSConfig `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// RouterTCPTLSConfig holds the TLS configuration for a router.
type RouterTCPTLSConfig struct {
	Passthrough  bool           `json:"passthrough" toml:"passthrough" yaml:"passthrough" export:"true"`
	Options      string         `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty" export:"true"`
	CertResolver string         `json:"certResolver,omitempty" toml:"certResolver,omitempty" yaml:"certResolver,omitempty" export:"true"`
	Domains      []types.Domain `json:"domains,omitempty" toml:"domains,omitempty" yaml:"domains,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TCPServersLoadBalancer holds the LoadBalancerService configuration.
type TCPServersLoadBalancer struct {
	ProxyProtocol    *ProxyProtocol `json:"proxyProtocol,omitempty" toml:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	Servers          []TCPServer    `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty" label-slice-as-struct:"server" export:"true"`
	ServersTransport string         `json:"serversTransport,omitempty" toml:"serversTransport,omitempty" yaml:"serversTransport,omitempty" export:"true"`

	// TerminationDelay, corresponds to the deadline that the proxy sets, after one
	// of its connected peers indicates it has closed the writing capability of its
	// connection, to close the reading capability as well, hence fully terminating the
	// connection. It is a duration in milliseconds, defaulting to 100. A negative value
	// means an infinite deadline (i.e. the reading capability is never closed).
	// Deprecated: use ServersTransport to configure the TerminationDelay instead.
	TerminationDelay *int `json:"terminationDelay,omitempty" toml:"terminationDelay,omitempty" yaml:"terminationDelay,omitempty" export:"true"`
}

// Mergeable tells if the given service is mergeable.
func (l *TCPServersLoadBalancer) Mergeable(loadBalancer *TCPServersLoadBalancer) bool {
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

// +k8s:deepcopy-gen=true

// TCPServer holds a TCP Server configuration.
type TCPServer struct {
	Address string `json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty" label:"-"`
	Port    string `json:"-" toml:"-" yaml:"-"`
	TLS     bool   `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

// +k8s:deepcopy-gen=true

// ProxyProtocol holds the PROXY Protocol configuration.
// More info: https://doc.traefik.io/traefik/v3.5/routing/services/#proxy-protocol
type ProxyProtocol struct {
	// Version defines the PROXY Protocol version to use.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=2
	Version int `json:"version,omitempty" toml:"version,omitempty" yaml:"version,omitempty" export:"true"`
}

// SetDefaults Default values for a ProxyProtocol.
func (p *ProxyProtocol) SetDefaults() {
	p.Version = 2
}

// +k8s:deepcopy-gen=true

// TCPServersTransport options to configure communication between Traefik and the servers.
type TCPServersTransport struct {
	DialKeepAlive ptypes.Duration `description:"Defines the interval between keep-alive probes for an active network connection. If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alive probes are disabled" json:"dialKeepAlive,omitempty" toml:"dialKeepAlive,omitempty" yaml:"dialKeepAlive,omitempty" export:"true"`
	DialTimeout   ptypes.Duration `description:"Defines the amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists." json:"dialTimeout,omitempty" toml:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty" export:"true"`
	// TerminationDelay, corresponds to the deadline that the proxy sets, after one
	// of its connected peers indicates it has closed the writing capability of its
	// connection, to close the reading capability as well, hence fully terminating the
	// connection. It is a duration in milliseconds, defaulting to 100. A negative value
	// means an infinite deadline (i.e. the reading capability is never closed).
	TerminationDelay ptypes.Duration  `description:"Defines the delay to wait before fully terminating the connection, after one connected peer has closed its writing capability." json:"terminationDelay,omitempty" toml:"terminationDelay,omitempty" yaml:"terminationDelay,omitempty" export:"true"`
	TLS              *TLSClientConfig `description:"Defines the TLS configuration." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TLSClientConfig options to configure TLS communication between Traefik and the servers.
type TLSClientConfig struct {
	ServerName         string                  `description:"Defines the serverName used to contact the server." json:"serverName,omitempty" toml:"serverName,omitempty" yaml:"serverName,omitempty"`
	InsecureSkipVerify bool                    `description:"Disables SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
	RootCAs            []types.FileOrContent   `description:"Defines a list of CA certificates used to validate server certificates." json:"rootCAs,omitempty" toml:"rootCAs,omitempty" yaml:"rootCAs,omitempty"`
	Certificates       traefiktls.Certificates `description:"Defines a list of client certificates for mTLS." json:"certificates,omitempty" toml:"certificates,omitempty" yaml:"certificates,omitempty" export:"true"`
	PeerCertURI        string                  `description:"Defines the URI used to match against SAN URI during the peer certificate verification." json:"peerCertURI,omitempty" toml:"peerCertURI,omitempty" yaml:"peerCertURI,omitempty" export:"true"`
	Spiffe             *Spiffe                 `description:"Defines the SPIFFE TLS configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// SetDefaults sets the default values for a TCPServersTransport.
func (t *TCPServersTransport) SetDefaults() {
	t.DialTimeout = ptypes.Duration(30 * time.Second)
	t.DialKeepAlive = ptypes.Duration(15 * time.Second)
	t.TerminationDelay = ptypes.Duration(100 * time.Millisecond)
}
