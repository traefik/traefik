package provider

import (
	"github.com/containous/traefik/types"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
)

// Consul holds configurations of the Consul provider.
type Consul struct {
	Kv `mapstructure:",squash"`
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Consul) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.storeType = store.CONSUL
	consul.Register()
	return provider.provide(configurationChan)
}
