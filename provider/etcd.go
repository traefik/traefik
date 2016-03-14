package provider

import (
	"github.com/containous/traefik/types"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
)

// Etcd holds configurations of the Etcd provider.
type Etcd struct {
	Kv `mapstructure:",squash"`
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Etcd) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.storeType = store.ETCD
	etcd.Register()
	return provider.provide(configurationChan)
}
