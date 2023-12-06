package dynamic

import (
	"reflect"
	"time"

	ptypes "github.com/traefik/paerser/types"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

const (
	// DefaultHealthCheckInterval is the default value for the ServerHealthCheck interval.
	DefaultHealthCheckInterval = ptypes.Duration(30 * time.Second)
	// DefaultHealthCheckTimeout is the default value for the ServerHealthCheck timeout.
	DefaultHealthCheckTimeout = ptypes.Duration(5 * time.Second)

	// DefaultPassHostHeader is the default value for the HTTPClientConfig passHostHeader.
	DefaultPassHostHeader = true

	// DefaultIdleConnTimeout is the default value for the ForwardingTimeouts idleConnTimeout.
	DefaultIdleConnTimeout = ptypes.Duration(90 * time.Second)
)

// +k8s:deepcopy-gen=true

// HTTPConfiguration contains all the HTTP configuration parameters.
type HTTPConfiguration struct {
	Routers           map[string]*Router           `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty" export:"true"`
	Services          map[string]*Service          `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty" export:"true"`
	Middlewares       map[string]*Middleware       `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	Models            map[string]*Model            `json:"models,omitempty" toml:"models,omitempty" yaml:"models,omitempty" export:"true"`
	ServersTransports map[string]*ServersTransport `json:"serversTransports,omitempty" toml:"serversTransports,omitempty" yaml:"serversTransports,omitempty" label:"-" export:"true"`
}

// +k8s:deepcopy-gen=true

// Model is a set of default router's values.
type Model struct {
	Middlewares []string         `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	TLS         *RouterTLSConfig `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Service holds a service configuration (can only be of one type at the same time).
type Service struct {
	LoadBalancer *ServersLoadBalancer `json:"loadBalancer,omitempty" toml:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty" export:"true"`
	Weighted     *WeightedRoundRobin  `json:"weighted,omitempty" toml:"weighted,omitempty" yaml:"weighted,omitempty" label:"-" export:"true"`
	Mirroring    *Mirroring           `json:"mirroring,omitempty" toml:"mirroring,omitempty" yaml:"mirroring,omitempty" label:"-" export:"true"`
	Failover     *Failover            `json:"failover,omitempty" toml:"failover,omitempty" yaml:"failover,omitempty" label:"-" export:"true"`
}

// +k8s:deepcopy-gen=true

// Router holds the router configuration.
type Router struct {
	EntryPoints []string         `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	Middlewares []string         `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	Service     string           `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	Rule        string           `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
	Priority    int              `json:"priority,omitempty" toml:"priority,omitempty,omitzero" yaml:"priority,omitempty" export:"true"`
	TLS         *RouterTLSConfig `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	DefaultRule bool             `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`
}

// +k8s:deepcopy-gen=true

// RouterTLSConfig holds the TLS configuration for a router.
type RouterTLSConfig struct {
	Options      string         `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty" export:"true"`
	CertResolver string         `json:"certResolver,omitempty" toml:"certResolver,omitempty" yaml:"certResolver,omitempty" export:"true"`
	Domains      []types.Domain `json:"domains,omitempty" toml:"domains,omitempty" yaml:"domains,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Mirroring holds the Mirroring configuration.
type Mirroring struct {
	Service     string          `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	MaxBodySize *int64          `json:"maxBodySize,omitempty" toml:"maxBodySize,omitempty" yaml:"maxBodySize,omitempty" export:"true"`
	Mirrors     []MirrorService `json:"mirrors,omitempty" toml:"mirrors,omitempty" yaml:"mirrors,omitempty" export:"true"`
	HealthCheck *HealthCheck    `json:"healthCheck,omitempty" toml:"healthCheck,omitempty" yaml:"healthCheck,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// SetDefaults Default values for a WRRService.
func (m *Mirroring) SetDefaults() {
	var defaultMaxBodySize int64 = -1
	m.MaxBodySize = &defaultMaxBodySize
}

// +k8s:deepcopy-gen=true

// Failover holds the Failover configuration.
type Failover struct {
	Service     string       `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	Fallback    string       `json:"fallback,omitempty" toml:"fallback,omitempty" yaml:"fallback,omitempty" export:"true"`
	HealthCheck *HealthCheck `json:"healthCheck,omitempty" toml:"healthCheck,omitempty" yaml:"healthCheck,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// MirrorService holds the MirrorService configuration.
type MirrorService struct {
	Name    string `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty" export:"true"`
	Percent int    `json:"percent,omitempty" toml:"percent,omitempty" yaml:"percent,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// WeightedRoundRobin is a weighted round robin load-balancer of services.
type WeightedRoundRobin struct {
	Services []WRRService `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty" export:"true"`
	Sticky   *Sticky      `json:"sticky,omitempty" toml:"sticky,omitempty" yaml:"sticky,omitempty" export:"true"`
	// HealthCheck enables automatic self-healthcheck for this service, i.e.
	// whenever one of its children is reported as down, this service becomes aware of it,
	// and takes it into account (i.e. it ignores the down child) when running the
	// load-balancing algorithm. In addition, if the parent of this service also has
	// HealthCheck enabled, this service reports to its parent any status change.
	HealthCheck *HealthCheck `json:"healthCheck,omitempty" toml:"healthCheck,omitempty" yaml:"healthCheck,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// WRRService is a reference to a service load-balanced with weighted round-robin.
type WRRService struct {
	Name   string `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty" export:"true"`
	Weight *int   `json:"weight,omitempty" toml:"weight,omitempty" yaml:"weight,omitempty" export:"true"`
}

// SetDefaults Default values for a WRRService.
func (w *WRRService) SetDefaults() {
	defaultWeight := 1
	w.Weight = &defaultWeight
}

// +k8s:deepcopy-gen=true

// Sticky holds the sticky configuration.
type Sticky struct {
	// Cookie defines the sticky cookie configuration.
	Cookie *Cookie `json:"cookie,omitempty" toml:"cookie,omitempty" yaml:"cookie,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Cookie holds the sticky configuration based on cookie.
type Cookie struct {
	// Name defines the Cookie name.
	Name string `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty" export:"true"`
	// Secure defines whether the cookie can only be transmitted over an encrypted connection (i.e. HTTPS).
	Secure bool `json:"secure,omitempty" toml:"secure,omitempty" yaml:"secure,omitempty" export:"true"`
	// HTTPOnly defines whether the cookie can be accessed by client-side APIs, such as JavaScript.
	HTTPOnly bool `json:"httpOnly,omitempty" toml:"httpOnly,omitempty" yaml:"httpOnly,omitempty" export:"true"`
	// SameSite defines the same site policy.
	// More info: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite
	SameSite string `json:"sameSite,omitempty" toml:"sameSite,omitempty" yaml:"sameSite,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ServersLoadBalancer holds the ServersLoadBalancer configuration.
type ServersLoadBalancer struct {
	Sticky  *Sticky  `json:"sticky,omitempty" toml:"sticky,omitempty" yaml:"sticky,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	Servers []Server `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty" label-slice-as-struct:"server" export:"true"`
	// HealthCheck enables regular active checks of the responsiveness of the
	// children servers of this load-balancer. To propagate status changes (e.g. all
	// servers of this service are down) upwards, HealthCheck must also be enabled on
	// the parent(s) of this service.
	HealthCheck      *ServerHealthCheck `json:"healthCheck,omitempty" toml:"healthCheck,omitempty" yaml:"healthCheck,omitempty" export:"true"`
	ServersTransport string             `json:"serversTransport,omitempty" toml:"serversTransport,omitempty" yaml:"serversTransport,omitempty" export:"true"`
}

// Mergeable tells if the given service is mergeable.
func (l *ServersLoadBalancer) Mergeable(loadBalancer *ServersLoadBalancer) bool {
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

// Server holds the server configuration.
type Server struct {
	URL    string `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty" label:"-"`
	Scheme string `json:"-" toml:"-" yaml:"-" file:"-"`
	Port   string `json:"-" toml:"-" yaml:"-" file:"-"`
}

// SetDefaults Default values for a Server.
func (s *Server) SetDefaults() {
	s.Scheme = "http"
}

// +k8s:deepcopy-gen=true

// ServerHealthCheck holds the HealthCheck configuration.
type ServerHealthCheck struct {
	Scheme          string            `json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty" export:"true"`
	Mode            string            `json:"mode,omitempty" toml:"mode,omitempty" yaml:"mode,omitempty" export:"true"`
	Path            string            `json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty" export:"true"`
	Method          string            `json:"method,omitempty" toml:"method,omitempty" yaml:"method,omitempty" export:"true"`
	Status          int               `json:"status,omitempty" toml:"status,omitempty" yaml:"status,omitempty" export:"true"`
	Port            int               `json:"port,omitempty" toml:"port,omitempty,omitzero" yaml:"port,omitempty" export:"true"`
	Interval        ptypes.Duration   `json:"interval,omitempty" toml:"interval,omitempty" yaml:"interval,omitempty" export:"true"`
	Timeout         ptypes.Duration   `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	Hostname        string            `json:"hostname,omitempty" toml:"hostname,omitempty" yaml:"hostname,omitempty"`
	FollowRedirects *bool             `json:"followRedirects" toml:"followRedirects" yaml:"followRedirects" export:"true"`
	Headers         map[string]string `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
}

// SetDefaults Default values for a HealthCheck.
func (h *ServerHealthCheck) SetDefaults() {
	fr := true
	h.FollowRedirects = &fr
	h.Mode = "http"
	h.Interval = DefaultHealthCheckInterval
	h.Timeout = DefaultHealthCheckTimeout
}

// +k8s:deepcopy-gen=true

// HealthCheck controls healthcheck awareness and propagation at the services level.
type HealthCheck struct{}

// +k8s:deepcopy-gen=true

// ServersTransport holds the configuration of the communication between Traefik and the servers.
type ServersTransport struct {
	// PassHostHeader defines whether to forward the client Host header to the server.
	PassHostHeader bool `json:"passHostHeader,omitempty" toml:"passHostHeader,omitempty" yaml:"passHostHeader,omitempty" export:"true"`
	// MaxIdleConnsPerHost controls the maximum number of idle (keep-alive) connections to keep per-host.
	// If zero, DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int `json:"maxIdleConnsPerHost,omitempty" toml:"maxIdleConnsPerHost,omitempty" yaml:"maxIdleConnsPerHost,omitempty" export:"true"`
	// ForwardingTimeouts defines the timeouts for the requests forwarded to the backend servers.
	ForwardingTimeouts *ForwardingTimeouts `json:"forwardingTimeouts,omitempty" toml:"forwardingTimeouts,omitempty" yaml:"forwardingTimeouts,omitempty" export:"true"`
	// EnableHTTP2 enables HTTP/2 between Traefik and the backends.
	EnableHTTP2 bool             `json:"enableHTTP2,omitempty" toml:"enableHTTP2,omitempty" yaml:"enableHTTP2,omitempty" export:"true"`
	TLS         *TLSClientConfig `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
}

func (s *ServersTransport) SetDefaults() {
	s.PassHostHeader = DefaultPassHostHeader
	s.MaxIdleConnsPerHost = 200
	s.ForwardingTimeouts = &ForwardingTimeouts{}
	s.ForwardingTimeouts.SetDefaults()
}

// +k8s:deepcopy-gen=true

// TLSClientConfig holds the TLS configuration to be used between Traefik and the servers.
type TLSClientConfig struct {
	ServerName         string                     `description:"Defines the serverName used to contact the server." json:"serverName,omitempty" toml:"serverName,omitempty" yaml:"serverName,omitempty"`
	InsecureSkipVerify bool                       `description:"Disables SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
	RootCAs            []traefiktls.FileOrContent `description:"Defines a list of CA secret used to validate self-signed certificate" json:"rootCAs,omitempty" toml:"rootCAs,omitempty" yaml:"rootCAs,omitempty"`
	Certificates       traefiktls.Certificates    `description:"Defines a list of secret storing client certificates for mTLS." json:"certificates,omitempty" toml:"certificates,omitempty" yaml:"certificates,omitempty" export:"true"`
	DisableHTTP2       bool                       `description:"Disables HTTP/2 for connections with backend servers." json:"disableHTTP2,omitempty" toml:"disableHTTP2,omitempty" yaml:"disableHTTP2,omitempty" export:"true"`
	PeerCertURI        string                     `description:"Defines the URI used to match against SAN URI during the peer certificate verification." json:"peerCertURI,omitempty" toml:"peerCertURI,omitempty" yaml:"peerCertURI,omitempty" export:"true"`
	Spiffe             *Spiffe                    `description:"Defines the SPIFFE configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Spiffe holds the SPIFFE configuration.
type Spiffe struct {
	// IDs defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain).
	IDs []string `json:"ids,omitempty" toml:"ids,omitempty" yaml:"ids,omitempty"`
	// TrustDomain defines the allowed SPIFFE trust domain.
	TrustDomain string `json:"trustDomain,omitempty" toml:"trustDomain,omitempty" yaml:"trustDomain,omitempty"`
}

// +k8s:deepcopy-gen=true

// ForwardingTimeouts contains the timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	// DialTimeout defines the amount of time to wait until a connection to a backend server can be established.
	// If zero, no timeout exists.
	DialTimeout ptypes.Duration `json:"dialTimeout,omitempty" toml:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty" export:"true"`
	// ResponseHeaderTimeout defines the amount of time to wait for a server's response headers after fully writing the request (including its body, if any).
	// If zero, no timeout exists.
	ResponseHeaderTimeout ptypes.Duration `json:"responseHeaderTimeout,omitempty" toml:"responseHeaderTimeout,omitempty" yaml:"responseHeaderTimeout,omitempty" export:"true"`
	// IdleConnTimeout defines the maximum period for which an idle HTTP keep-alive connection will remain open before closing itself.
	IdleConnTimeout ptypes.Duration `json:"idleConnTimeout,omitempty" toml:"idleConnTimeout,omitempty" yaml:"idleConnTimeout,omitempty" export:"true"`
	// ReadIdleTimeout defines the timeout after which a health check using ping frame will be carried out if no frame is received on the HTTP/2 connection.
	// If zero, no health check is performed.
	ReadIdleTimeout ptypes.Duration `json:"readIdleTimeout,omitempty" toml:"readIdleTimeout,omitempty" yaml:"readIdleTimeout,omitempty" export:"true"`
	// PingTimeout defines the timeout after which the HTTP/2 connection will be closed if a response to ping is not received.
	PingTimeout ptypes.Duration `json:"pingTimeout,omitempty" toml:"pingTimeout,omitempty" yaml:"pingTimeout,omitempty" export:"true"`
}

// SetDefaults sets the default ForwardingTimeouts values.
func (f *ForwardingTimeouts) SetDefaults() {
	f.DialTimeout = ptypes.Duration(30 * time.Second)
	f.IdleConnTimeout = DefaultIdleConnTimeout
	f.PingTimeout = ptypes.Duration(15 * time.Second)
}
