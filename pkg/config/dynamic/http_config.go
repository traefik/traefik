package dynamic

import "reflect"

// +k8s:deepcopy-gen=true

// HTTPConfiguration contains all the HTTP configuration parameters.
type HTTPConfiguration struct {
	Routers     map[string]*Router     `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty"`
	Middlewares map[string]*Middleware `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty"`
	Services    map[string]*Service    `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty"`
}

// +k8s:deepcopy-gen=true

// Service holds a service configuration (can only be of one type at the same time).
type Service struct {
	LoadBalancer *LoadBalancerService `json:"loadBalancer,omitempty" toml:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty"`
}

// +k8s:deepcopy-gen=true

// Router holds the router configuration.
type Router struct {
	EntryPoints []string         `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty"`
	Middlewares []string         `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty"`
	Service     string           `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty"`
	Rule        string           `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
	Priority    int              `json:"priority,omitempty" toml:"priority,omitempty,omitzero" yaml:"priority,omitempty"`
	TLS         *RouterTLSConfig `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty"`
}

// +k8s:deepcopy-gen=true

// RouterTLSConfig holds the TLS configuration for a router
type RouterTLSConfig struct {
	Options string `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty"`
}

// +k8s:deepcopy-gen=true

// LoadBalancerService holds the LoadBalancerService configuration.
type LoadBalancerService struct {
	Stickiness         *Stickiness         `json:"stickiness,omitempty" toml:"stickiness,omitempty" yaml:"stickiness,omitempty" label:"allowEmpty"`
	Servers            []Server            `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty" label-slice-as-struct:"server"`
	HealthCheck        *HealthCheck        `json:"healthCheck,omitempty" toml:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	PassHostHeader     bool                `json:"passHostHeader" toml:"passHostHeader" yaml:"passHostHeader"`
	ResponseForwarding *ResponseForwarding `json:"responseForwarding,omitempty" toml:"responseForwarding,omitempty" yaml:"responseForwarding,omitempty"`
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

// +k8s:deepcopy-gen=true

// ResponseForwarding holds configuration for the forward of the response.
type ResponseForwarding struct {
	FlushInterval string `json:"flushInterval,omitempty" toml:"flushInterval,omitempty" yaml:"flushInterval,omitempty"`
}

// +k8s:deepcopy-gen=true

// Stickiness holds the stickiness configuration.
type Stickiness struct {
	CookieName     string `json:"cookieName,omitempty" toml:"cookieName,omitempty" yaml:"cookieName,omitempty"`
	SecureCookie   bool   `json:"secureCookie,omitempty" toml:"secureCookie,omitempty" yaml:"secureCookie,omitempty"`
	HTTPOnlyCookie bool   `json:"httpOnlyCookie,omitempty" toml:"httpOnlyCookie,omitempty" yaml:"httpOnlyCookie,omitempty"`
}

// +k8s:deepcopy-gen=true

// Server holds the server configuration.
type Server struct {
	URL    string `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty" label:"-"`
	Scheme string `toml:"-" json:"-" yaml:"-"`
	Port   string `toml:"-" json:"-" yaml:"-"`
}

// SetDefaults Default values for a Server.
func (s *Server) SetDefaults() {
	s.Scheme = "http"
}

// +k8s:deepcopy-gen=true

// HealthCheck holds the HealthCheck configuration.
type HealthCheck struct {
	Scheme string `json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty"`
	Path   string `json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty"`
	Port   int    `json:"port,omitempty" toml:"port,omitempty,omitzero" yaml:"port,omitempty"`
	// FIXME change string to types.Duration
	Interval string `json:"interval,omitempty" toml:"interval,omitempty" yaml:"interval,omitempty"`
	// FIXME change string to types.Duration
	Timeout  string            `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty"`
	Hostname string            `json:"hostname,omitempty" toml:"hostname,omitempty" yaml:"hostname,omitempty"`
	Headers  map[string]string `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
}
