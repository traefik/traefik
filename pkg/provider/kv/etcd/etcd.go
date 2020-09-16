package etcd

import (
	"github.com/abronan/valkeyrie/store"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/kv"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	kv.Provider
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Provider.SetDefaults()
	p.Endpoints = []string{"127.0.0.1:2379"}
}

// Init the provider.
func (p *Provider) Init() error {
	return p.Provider.Init(store.ETCDV3, "etcd")
}
