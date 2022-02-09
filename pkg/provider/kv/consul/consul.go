package consul

import (
	"errors"

	"github.com/kvtools/valkeyrie/store"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/kv"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	kv.Provider `export:"true"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Provider.SetDefaults()
	p.Endpoints = []string{"127.0.0.1:8500"}
}

// Init the provider.
func (p *Provider) Init() error {
	// Wildcard namespace allows fetching KV values from any namespace for recursive requests (see https://www.consul.io/api/kv#ns).
	// As we are not supporting multiple namespaces at the same time, wildcard namespace is not allowed.
	if p.Namespace == "*" {
		return errors.New("wildcard namespace is not supported")
	}

	return p.Provider.Init(store.CONSUL, "consul")
}
