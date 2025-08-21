package dynamic

import (
	"reflect"
	"time"

	ptypes "github.com/traefik/paerser/types"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	"google.golang.org/grpc/codes"
)

const (
	// DefaultHealthCheckInterval is the default value for the ServerHealthCheck interval.
	DefaultHealthCheckInterval = ptypes.Duration(30 * time.Second)
	// DefaultHealthCheckTimeout is the default value for the ServerHealthCheck timeout.
	DefaultHealthCheckTimeout = ptypes.Duration(5 * time.Second)

	// DefaultPassHostHeader is the default value for the ServersLoadBalancer passHostHeader.
	DefaultPassHostHeader = true

	// DefaultFlushInterval is the default value for the ResponseForwarding flush interval.
	DefaultFlushInterval = ptypes.Duration(100 * time.Millisecond)

	// MirroringDefaultMirrorBody is the Mirroring.MirrorBody option default value.
	MirroringDefaultMirrorBody = true
	// MirroringDefaultMaxBodySize is the Mirroring.MaxBodySize option default value.
	MirroringDefaultMaxBodySize int64 = -1
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

// Model holds model configuration.
type Model struct {
	Middlewares       []string                  `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	TLS               *RouterTLSConfig          `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	Observability     RouterObservabilityConfig `json:"observability,omitempty" toml:"observability,omitempty" yaml:"observability,omitempty" export:"true"`
	DefaultRuleSyntax string                    `json:"-" toml:"-" yaml:"-" label:"-" file:"-" kv:"-" export:"true"`
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
	EntryPoints []string `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	Middlewares []string `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	Service     string   `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	Rule        string   `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
	// Deprecated: Please do not use this field and rewrite the router rules to use the v3 syntax.
	RuleSyntax    string                     `json:"ruleSyntax,omitempty" toml:"ruleSyntax,omitempty" yaml:"ruleSyntax,omitempty" export:"true"`
	Priority      int                        `json:"priority,omitempty" toml:"priority,omitempty,omitzero" yaml:"priority,omitempty" export:"true"`
	TLS           *RouterTLSConfig           `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	Observability *RouterObservabilityConfig `json:"observability,omitempty" toml:"observability,omitempty" yaml:"observability,omitempty" export:"true"`
	DefaultRule   bool                       `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`
}

// +k8s:deepcopy-gen=true

// RouterTLSConfig holds the TLS configuration for a router.
type RouterTLSConfig struct {
	Options      string         `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty" export:"true"`
	CertResolver string         `json:"certResolver,omitempty" toml:"certResolver,omitempty" yaml:"certResolver,omitempty" export:"true"`
	Domains      []types.Domain `json:"domains,omitempty" toml:"domains,omitempty" yaml:"domains,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// RouterObservabilityConfig holds the observability configuration for a router.
type RouterObservabilityConfig struct {
	// AccessLogs enables access logs for this router.
	AccessLogs *bool `json:"accessLogs,omitempty" toml:"accessLogs,omitempty" yaml:"accessLogs,omitempty" export:"true"`
	// Metrics enables metrics for this router.
	Metrics *bool `json:"metrics,omitempty" toml:"metrics,omitempty" yaml:"metrics,omitempty" export:"true"`
	// Tracing enables tracing for this router.
	Tracing *bool `json:"tracing,omitempty" toml:"tracing,omitempty" yaml:"tracing,omitempty" export:"true"`
	// TraceVerbosity defines the verbosity level of the tracing for this router.
	// +kubebuilder:validation:Enum=minimal;detailed
	// +kubebuilder:default=minimal
	TraceVerbosity types.TracingVerbosity `json:"traceVerbosity,omitempty" toml:"traceVerbosity,omitempty" yaml:"traceVerbosity,omitempty" export:"true"`
}

// SetDefaults Default values for a RouterObservabilityConfig.
func (r *RouterObservabilityConfig) SetDefaults() {
	r.TraceVerbosity = types.MinimalVerbosity
}

// +k8s:deepcopy-gen=true

// Mirroring holds the Mirroring configuration.
type Mirroring struct {
	Service     string          `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	MirrorBody  *bool           `json:"mirrorBody,omitempty" toml:"mirrorBody,omitempty" yaml:"mirrorBody,omitempty" export:"true"`
	MaxBodySize *int64          `json:"maxBodySize,omitempty" toml:"maxBodySize,omitempty" yaml:"maxBodySize,omitempty" export:"true"`
	Mirrors     []MirrorService `json:"mirrors,omitempty" toml:"mirrors,omitempty" yaml:"mirrors,omitempty" export:"true"`
	HealthCheck *HealthCheck    `json:"healthCheck,omitempty" toml:"healthCheck,omitempty" yaml:"healthCheck,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// SetDefaults Default values for a WRRService.
func (m *Mirroring) SetDefaults() {
	defaultMirrorBody := MirroringDefaultMirrorBody
	m.MirrorBody = &defaultMirrorBody
	defaultMaxBodySize := MirroringDefaultMaxBodySize
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

	// Status defines an HTTP status code that should be returned when calling the service.
	// This is required by the Gateway API implementation which expects specific HTTP status to be returned.
	Status *int `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`
	// GRPCStatus defines a GRPC status code that should be returned when calling the service.
	// This is required by the Gateway API implementation which expects specific GRPC status to be returned.
	GRPCStatus *GRPCStatus `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`
}

// SetDefaults Default values for a WRRService.
func (w *WRRService) SetDefaults() {
	defaultWeight := 1
	w.Weight = &defaultWeight
}

// +k8s:deepcopy-gen=true

type GRPCStatus struct {
	Code codes.Code `json:"code,omitempty" toml:"code,omitempty" yaml:"code,omitempty" export:"true"`
	Msg  string     `json:"msg,omitempty" toml:"msg,omitempty" yaml:"msg,omitempty" export:"true"`
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
	// +kubebuilder:validation:Enum=none;lax;strict
	SameSite string `json:"sameSite,omitempty" toml:"sameSite,omitempty" yaml:"sameSite,omitempty" export:"true"`
	// MaxAge defines the number of seconds until the cookie expires.
	// When set to a negative number, the cookie expires immediately.
	// When set to zero, the cookie never expires.
	MaxAge int `json:"maxAge,omitempty" toml:"maxAge,omitempty" yaml:"maxAge,omitempty" export:"true"`
	// Path defines the path that must exist in the requested URL for the browser to send the Cookie header.
	// When not provided the cookie will be sent on every request to the domain.
	// More info: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie#pathpath-value
	Path *string `json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty" export:"true"`
	// Domain defines the host to which the cookie will be sent.
	// More info: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie#domaindomain-value
	Domain string `json:"domain,omitempty" toml:"domain,omitempty" yaml:"domain,omitempty"`
}

// SetDefaults set the default values for a Cookie.
func (c *Cookie) SetDefaults() {
	defaultPath := "/"
	c.Path = &defaultPath
}

type BalancerStrategy string

const (
	// BalancerStrategyWRR is the weighted round-robin strategy.
	BalancerStrategyWRR BalancerStrategy = "wrr"
	// BalancerStrategyP2C is the power of two choices strategy.
	BalancerStrategyP2C BalancerStrategy = "p2c"
)

// +k8s:deepcopy-gen=true

// ServersLoadBalancer holds the ServersLoadBalancer configuration.
type ServersLoadBalancer struct {
	Sticky   *Sticky          `json:"sticky,omitempty" toml:"sticky,omitempty" yaml:"sticky,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	Servers  []Server         `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty" label-slice-as-struct:"server" export:"true"`
	Strategy BalancerStrategy `json:"strategy,omitempty" toml:"strategy,omitempty" yaml:"strategy,omitempty" export:"true"`
	// HealthCheck enables regular active checks of the responsiveness of the
	// children servers of this load-balancer. To propagate status changes (e.g. all
	// servers of this service are down) upwards, HealthCheck must also be enabled on
	// the parent(s) of this service.
	HealthCheck *ServerHealthCheck `json:"healthCheck,omitempty" toml:"healthCheck,omitempty" yaml:"healthCheck,omitempty" export:"true"`
	// PassiveHealthCheck enables passive health checks for children servers of this load-balancer.
	PassiveHealthCheck *PassiveServerHealthCheck `json:"passiveHealthCheck,omitempty" toml:"passiveHealthCheck,omitempty" yaml:"passiveHealthCheck,omitempty" export:"true"`
	PassHostHeader     *bool                     `json:"passHostHeader" toml:"passHostHeader" yaml:"passHostHeader" export:"true"`
	ResponseForwarding *ResponseForwarding       `json:"responseForwarding,omitempty" toml:"responseForwarding,omitempty" yaml:"responseForwarding,omitempty" export:"true"`
	ServersTransport   string                    `json:"serversTransport,omitempty" toml:"serversTransport,omitempty" yaml:"serversTransport,omitempty" export:"true"`
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

// SetDefaults Default values for a ServersLoadBalancer.
func (l *ServersLoadBalancer) SetDefaults() {
	defaultPassHostHeader := DefaultPassHostHeader
	l.PassHostHeader = &defaultPassHostHeader

	l.Strategy = BalancerStrategyWRR
	l.ResponseForwarding = &ResponseForwarding{}
	l.ResponseForwarding.SetDefaults()
}

// +k8s:deepcopy-gen=true

// ResponseForwarding holds the response forwarding configuration.
type ResponseForwarding struct {
	// FlushInterval defines the interval, in milliseconds, in between flushes to the client while copying the response body.
	// A negative value means to flush immediately after each write to the client.
	// This configuration is ignored when ReverseProxy recognizes a response as a streaming response;
	// for such responses, writes are flushed to the client immediately.
	// Default: 100ms
	FlushInterval ptypes.Duration `json:"flushInterval,omitempty" toml:"flushInterval,omitempty" yaml:"flushInterval,omitempty" export:"true"`
}

// SetDefaults Default values for a ResponseForwarding.
func (r *ResponseForwarding) SetDefaults() {
	r.FlushInterval = DefaultFlushInterval
}

// +k8s:deepcopy-gen=true

// Server holds the server configuration.
type Server struct {
	URL          string `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty"`
	Weight       *int   `json:"weight,omitempty" toml:"weight,omitempty" yaml:"weight,omitempty" export:"true"`
	PreservePath bool   `json:"preservePath,omitempty" toml:"preservePath,omitempty" yaml:"preservePath,omitempty" export:"true"`
	Fenced       bool   `json:"fenced,omitempty" toml:"-" yaml:"-" label:"-" file:"-" kv:"-"`
	// Scheme can only be defined with label Providers.
	Scheme string `json:"-" toml:"-" yaml:"-" file:"-" kv:"-"`
	Port   string `json:"-" toml:"-" yaml:"-" file:"-" kv:"-"`
}

// +k8s:deepcopy-gen=true

// ServerHealthCheck holds the HealthCheck configuration.
type ServerHealthCheck struct {
	Scheme            string            `json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty" export:"true"`
	Mode              string            `json:"mode,omitempty" toml:"mode,omitempty" yaml:"mode,omitempty" export:"true"`
	Path              string            `json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty" export:"true"`
	Method            string            `json:"method,omitempty" toml:"method,omitempty" yaml:"method,omitempty" export:"true"`
	Status            int               `json:"status,omitempty" toml:"status,omitempty" yaml:"status,omitempty" export:"true"`
	Port              int               `json:"port,omitempty" toml:"port,omitempty,omitzero" yaml:"port,omitempty" export:"true"`
	Interval          ptypes.Duration   `json:"interval,omitempty" toml:"interval,omitempty" yaml:"interval,omitempty" export:"true"`
	UnhealthyInterval *ptypes.Duration  `json:"unhealthyInterval,omitempty" toml:"unhealthyInterval,omitempty" yaml:"unhealthyInterval,omitempty" export:"true"`
	Timeout           ptypes.Duration   `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	Hostname          string            `json:"hostname,omitempty" toml:"hostname,omitempty" yaml:"hostname,omitempty"`
	FollowRedirects   *bool             `json:"followRedirects,omitempty" toml:"followRedirects,omitempty" yaml:"followRedirects,omitempty" export:"true"`
	Headers           map[string]string `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
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

type PassiveServerHealthCheck struct {
	// FailureWindow defines the time window during which the failed attempts must occur for the server to be marked as unhealthy. It also defines for how long the server will be considered unhealthy.
	FailureWindow ptypes.Duration `json:"failureWindow,omitempty" toml:"failureWindow,omitempty" yaml:"failureWindow,omitempty" export:"true"`
	// MaxFailedAttempts is the number of consecutive failed attempts allowed within the failure window before marking the server as unhealthy.
	MaxFailedAttempts int `json:"maxFailedAttempts,omitempty" toml:"maxFailedAttempts,omitempty" yaml:"maxFailedAttempts,omitempty" export:"true"`
}

func (p *PassiveServerHealthCheck) SetDefaults() {
	p.FailureWindow = ptypes.Duration(10 * time.Second)
	p.MaxFailedAttempts = 1
}

// +k8s:deepcopy-gen=true

// HealthCheck controls healthcheck awareness and propagation at the services level.
type HealthCheck struct{}

// +k8s:deepcopy-gen=true

// ServersTransport options to configure communication between Traefik and the servers.
type ServersTransport struct {
	ServerName          string                  `description:"Defines the serverName used to contact the server." json:"serverName,omitempty" toml:"serverName,omitempty" yaml:"serverName,omitempty"`
	InsecureSkipVerify  bool                    `description:"Disables SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
	RootCAs             []types.FileOrContent   `description:"Defines a list of CA certificates used to validate server certificates." json:"rootCAs,omitempty" toml:"rootCAs,omitempty" yaml:"rootCAs,omitempty"`
	Certificates        traefiktls.Certificates `description:"Defines a list of client certificates for mTLS." json:"certificates,omitempty" toml:"certificates,omitempty" yaml:"certificates,omitempty" export:"true"`
	MaxIdleConnsPerHost int                     `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used" json:"maxIdleConnsPerHost,omitempty" toml:"maxIdleConnsPerHost,omitempty" yaml:"maxIdleConnsPerHost,omitempty" export:"true"`
	ForwardingTimeouts  *ForwardingTimeouts     `description:"Defines the timeouts for requests forwarded to the backend servers." json:"forwardingTimeouts,omitempty" toml:"forwardingTimeouts,omitempty" yaml:"forwardingTimeouts,omitempty" export:"true"`
	DisableHTTP2        bool                    `description:"Disables HTTP/2 for connections with backend servers." json:"disableHTTP2,omitempty" toml:"disableHTTP2,omitempty" yaml:"disableHTTP2,omitempty" export:"true"`
	PeerCertURI         string                  `description:"Defines the URI used to match against SAN URI during the peer certificate verification." json:"peerCertURI,omitempty" toml:"peerCertURI,omitempty" yaml:"peerCertURI,omitempty" export:"true"`
	Spiffe              *Spiffe                 `description:"Defines the SPIFFE configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Spiffe holds the SPIFFE configuration.
type Spiffe struct {
	// IDs defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain).
	IDs []string `description:"Defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain)." json:"ids,omitempty" toml:"ids,omitempty" yaml:"ids,omitempty"`
	// TrustDomain defines the allowed SPIFFE trust domain.
	TrustDomain string `description:"Defines the allowed SPIFFE trust domain." json:"trustDomain,omitempty" toml:"trustDomain,omitempty" yaml:"trustDomain,omitempty"`
}

// +k8s:deepcopy-gen=true

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           ptypes.Duration `description:"The amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists." json:"dialTimeout,omitempty" toml:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty" export:"true"`
	ResponseHeaderTimeout ptypes.Duration `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists." json:"responseHeaderTimeout,omitempty" toml:"responseHeaderTimeout,omitempty" yaml:"responseHeaderTimeout,omitempty" export:"true"`
	IdleConnTimeout       ptypes.Duration `description:"The maximum period for which an idle HTTP keep-alive connection will remain open before closing itself." json:"idleConnTimeout,omitempty" toml:"idleConnTimeout,omitempty" yaml:"idleConnTimeout,omitempty" export:"true"`
	ReadIdleTimeout       ptypes.Duration `description:"The timeout after which a health check using ping frame will be carried out if no frame is received on the HTTP/2 connection. If zero, no health check is performed." json:"readIdleTimeout,omitempty" toml:"readIdleTimeout,omitempty" yaml:"readIdleTimeout,omitempty" export:"true"`
	PingTimeout           ptypes.Duration `description:"The timeout after which the HTTP/2 connection will be closed if a response to ping is not received." json:"pingTimeout,omitempty" toml:"pingTimeout,omitempty" yaml:"pingTimeout,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (f *ForwardingTimeouts) SetDefaults() {
	f.DialTimeout = ptypes.Duration(30 * time.Second)
	f.IdleConnTimeout = ptypes.Duration(90 * time.Second)
	f.PingTimeout = ptypes.Duration(15 * time.Second)
}
