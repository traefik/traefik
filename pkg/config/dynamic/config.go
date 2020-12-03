package dynamic

import (
	"github.com/traefik/traefik/v2/pkg/tls"
)

// +k8s:deepcopy-gen=true

// Message holds configuration information exchanged between parts of traefik.
type Message struct {
	ProviderName  string
	Configuration *Configuration
}

// +k8s:deepcopy-gen=true

// Configurations is for currentConfigurations Map.
type Configurations map[string]*Configuration

// +k8s:deepcopy-gen=true

// Configuration is the root of the dynamic configuration.
type Configuration struct {
	HTTP *HTTPConfiguration `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" export:"true"`
	TCP  *TCPConfiguration  `json:"tcp,omitempty" toml:"tcp,omitempty" yaml:"tcp,omitempty" export:"true"`
	UDP  *UDPConfiguration  `json:"udp,omitempty" toml:"udp,omitempty" yaml:"udp,omitempty" export:"true"`
	TLS  *TLSConfiguration  `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TLSConfiguration contains all the configuration parameters of a TLS connection.
type TLSConfiguration struct {
	Certificates []*tls.CertAndStores   `json:"certificates,omitempty"  toml:"certificates,omitempty" yaml:"certificates,omitempty" label:"-" export:"true"`
	Options      map[string]tls.Options `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty" export:"true"`
	Stores       map[string]tls.Store   `json:"stores,omitempty" toml:"stores,omitempty" yaml:"stores,omitempty" export:"true"`
}
