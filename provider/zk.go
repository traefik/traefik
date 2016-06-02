package provider

import (
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/zookeeper"
)

// Zookepper holds configurations of the Zookepper provider.
type Zookepper struct {
	Kv
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Zookepper) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints []types.Constraint) error {
	provider.storeType = store.ZK
	zookeeper.Register()
	return provider.provide(configurationChan, pool, constraints)
}
