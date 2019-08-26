package dynamic

import (
	"reflect"

	"github.com/containous/traefik/v2/pkg/types"
)

// +k8s:deepcopy-gen=true

// TCPConfiguration contains all the TCP configuration parameters.
type TCPConfiguration struct {
	Routers  map[string]*TCPRouter  `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty"`
	Services map[string]*TCPService `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty"`
}

// +k8s:deepcopy-gen=true

// TCPService holds a tcp service configuration (can only be of one type at the same time).
type TCPService struct {
	LoadBalancer *TCPLoadBalancerService `json:"loadBalancer,omitempty" toml:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty"`
}

// +k8s:deepcopy-gen=true

// TCPRouter holds the router configuration.
type TCPRouter struct {
	EntryPoints []string            `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty"`
	Service     string              `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty"`
	Rule        string              `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
	TLS         *RouterTCPTLSConfig `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty"`
}

// +k8s:deepcopy-gen=true

// RouterTCPTLSConfig holds the TLS configuration for a router
type RouterTCPTLSConfig struct {
	Passthrough  bool           `json:"passthrough" toml:"passthrough" yaml:"passthrough"`
	Options      string         `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty"`
	CertResolver string         `json:"certResolver,omitempty" toml:"certResolver,omitempty" yaml:"certResolver,omitempty"`
	Domains      []types.Domain `json:"domains,omitempty" toml:"domains,omitempty" yaml:"domains,omitempty"`
}

// +k8s:deepcopy-gen=true

// TCPLoadBalancerService holds the LoadBalancerService configuration.
type TCPLoadBalancerService struct {
	Servers []TCPServer `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty" label-slice-as-struct:"server"`
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

// +k8s:deepcopy-gen=true

// TCPServer holds a TCP Server configuration
type TCPServer struct {
	Address string `json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty" label:"-"`
	Port    string `toml:"-" json:"-" yaml:"-"`
}
