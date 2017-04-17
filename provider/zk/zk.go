package zk

import (
	"fmt"

	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/kv"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/zookeeper"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	kv.Provider `mapstructure:",squash"`
}

// Provide allows the zk provider to Provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	store, err := p.CreateStore()
	if err != nil {
		return fmt.Errorf("Failed to Connect to KV store: %v", err)
	}
	p.Kvclient = store
	return p.Provider.Provide(configurationChan, pool, constraints)
}

// CreateStore creates the KV store
func (p *Provider) CreateStore() (store.Store, error) {
	p.StoreType = store.ZK
	zookeeper.Register()
	return p.Provider.CreateStore()
}
