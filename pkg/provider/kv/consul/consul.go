package consul

import (
	"errors"

	"github.com/kvtools/valkeyrie/store"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/kv"
)

// providerName is the Consul provider name.
const providerName = "consul"

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	kv.Provider `export:"true"`

	Namespaces []string `description:"Sets the namespaces used to discover the configuration (Consul Enterprise only)." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`

	name string
}

// BuildNamespacedProviders builds Consul provider instances for the given namespace configuration.
func BuildNamespacedProviders(conf *Provider) []*Provider {
	if len(conf.Namespaces) == 0 {
		confCopy := *conf
		return []*Provider{&confCopy}
	}

	var providers []*Provider
	for _, namespace := range conf.Namespaces {
		confCopy := *conf
		confCopy.name = providerName + "-" + namespace
		providers = append(providers, &confCopy)
	}

	return providers
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Provider.SetDefaults()
	p.Endpoints = []string{"127.0.0.1:8500"}
	p.name = providerName
}

// Init the provider.
func (p *Provider) Init() error {
	// Wildcard namespace allows fetching KV values from any namespace for recursive requests (see https://www.consul.io/api/kv#ns).
	// As we are not supporting multiple namespaces at the same time, wildcard namespace is not allowed.
	if p.Namespace == "*" {
		return errors.New("wildcard namespace is not supported")
	}

	return p.Provider.Init(store.CONSUL, p.name)
}
