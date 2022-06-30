package consul

import (
	"errors"

	"github.com/kvtools/valkeyrie/store"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/kv"
)

// providerName is the Consul provider name.
const providerName = "consul"

var _ provider.Provider = (*Provider)(nil)

// ProviderBuilder is responsible for constructing namespaced instances of the Consul provider.
type ProviderBuilder struct {
	kv.Provider `export:"true"`

	// Deprecated: use Namespaces instead.
	Namespace  string   `description:"Sets the namespace used to discover the configuration (Consul Enterprise only)." json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	Namespaces []string `description:"Sets the namespaces used to discover the configuration (Consul Enterprise only)." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty"`
}

// SetDefaults sets the default values.
func (p *ProviderBuilder) SetDefaults() {
	p.Provider.SetDefaults()
	p.Endpoints = []string{"127.0.0.1:8500"}
}

// BuildProviders builds Consul provider instances for the given namespaces configuration.
func (p *ProviderBuilder) BuildProviders() []*Provider {
	// We can warn about that, because we've already made sure before that
	// Namespace and Namespaces are mutually exclusive.
	if p.Namespace != "" {
		log.WithoutContext().Warnf("Namespace option is deprecated, please use the Namespaces option instead.")
	}

	if len(p.Namespaces) == 0 {
		return []*Provider{{
			Provider: p.Provider,
			name:     providerName,
			// p.Namespace could very well be empty.
			namespace: p.Namespace,
		}}
	}

	var providers []*Provider
	for _, namespace := range p.Namespaces {
		providers = append(providers, &Provider{
			Provider:  p.Provider,
			name:      providerName + "-" + namespace,
			namespace: namespace,
		})
	}

	return providers
}

// Provider holds configurations of the provider.
type Provider struct {
	kv.Provider

	name      string
	namespace string
}

// Init the provider.
func (p *Provider) Init() error {
	// Wildcard namespace allows fetching KV values from any namespace for recursive requests (see https://www.consul.io/api/kv#ns).
	// As we are not supporting multiple namespaces at the same time, wildcard namespace is not allowed.
	if p.namespace == "*" {
		return errors.New("wildcard namespace is not supported")
	}

	// In case they didn't initialize with BuildProviders.
	if p.name == "" {
		p.name = providerName
	}

	return p.Provider.Init(store.CONSUL, p.name, p.namespace)
}
