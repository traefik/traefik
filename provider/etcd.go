package provider

import (
	"fmt"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
)

var _ Provider = (*Etcd)(nil)

// Etcd holds configurations of the Etcd provider.
type Etcd struct {
	Kv `mapstructure:",squash"`
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Etcd) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	store, err := provider.CreateStore()
	if err != nil {
		return fmt.Errorf("Failed to Connect to KV store: %v", err)
	}
	provider.kvclient = store
	return provider.provide(configurationChan, pool, constraints)
}

// CreateStore creates the KV store
func (provider *Etcd) CreateStore() (store.Store, error) {
	provider.storeType = store.ETCD
	etcd.Register()
	return provider.createStore()
}
