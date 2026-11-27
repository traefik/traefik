package dynamic

import (
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
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
	Options      map[string]tls.Options `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty" label:"-" export:"true"`
	Stores       map[string]tls.Store   `json:"stores,omitempty" toml:"stores,omitempty" yaml:"stores,omitempty" export:"true"`
}

func (c *Configuration) IsEmpty() bool {
	if c.TCP == nil {
		c.TCP = &TCPConfiguration{}
	}
	if c.HTTP == nil {
		c.HTTP = &HTTPConfiguration{}
	}
	if c.UDP == nil {
		c.UDP = &UDPConfiguration{}
	}

	httpEmpty := c.HTTP.Routers == nil && c.HTTP.Services == nil && c.HTTP.Middlewares == nil
	tlsEmpty := c.TLS == nil || c.TLS.Certificates == nil && c.TLS.Stores == nil && c.TLS.Options == nil
	tcpEmpty := c.TCP.Routers == nil && c.TCP.Services == nil && c.TCP.Middlewares == nil
	udpEmpty := c.UDP.Routers == nil && c.UDP.Services == nil

	return httpEmpty && tlsEmpty && tcpEmpty && udpEmpty
}

func (c *Configuration) GetSanitizedConfigurationTLS() *Configuration {
	copyConf := c.DeepCopy()

	if copyConf.TLS != nil {
		copyConf.TLS.Certificates = nil

		if copyConf.TLS.Options != nil {
			cleanedOptions := make(map[string]tls.Options, len(copyConf.TLS.Options))

			for name, option := range copyConf.TLS.Options {
				option.ClientAuth.CAFiles = []types.FileOrContent{}
				cleanedOptions[name] = option
			}

			copyConf.TLS.Options = cleanedOptions
		}

		for k := range copyConf.TLS.Stores {
			st := copyConf.TLS.Stores[k]
			st.DefaultCertificate = nil
			copyConf.TLS.Stores[k] = st
		}
	}

	if copyConf.HTTP != nil {
		for _, transport := range copyConf.HTTP.ServersTransports {
			transport.Certificates = tls.Certificates{}
			transport.RootCAs = []types.FileOrContent{}
		}
	}

	if copyConf.TCP != nil {
		for _, transport := range copyConf.TCP.ServersTransports {
			if transport.TLS != nil {
				transport.TLS.Certificates = tls.Certificates{}
				transport.TLS.RootCAs = []types.FileOrContent{}
			}
		}
	}
	return copyConf
}

func (c *Configuration) SummaryTLS() map[string]any {
	if c.TLS == nil {
		return nil
	}

	stores := make([]string, 0, len(c.TLS.Stores))
	defaultCerts := 0
	for name, st := range c.TLS.Stores {
		stores = append(stores, name)
		if st.DefaultCertificate != nil {
			defaultCerts++
		}
	}

	caFiles := 0
	if c.TLS.Options != nil {
		for _, opt := range c.TLS.Options {
			caFiles += len(opt.ClientAuth.CAFiles)
		}
	}

	return map[string]any{
		"certificates": map[string]any{
			"present": len(c.TLS.Certificates) > 0,
			"count":   len(c.TLS.Certificates),
			"value":   "<redacted>",
		},
		"stores": map[string]any{
			"names":              stores,
			"default_cert_count": defaultCerts,
			"default_cert":       "<redacted>",
		},
		"clientAuth": map[string]any{
			"ca_files_total": caFiles,
			"ca_files":       "<redacted>",
		},
	}
}
