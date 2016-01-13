package provider

import (
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/emilevauge/traefik/types"
)

// BoltDb holds configurations of the BoltDb provider.
type BoltDb struct {
	Kv `mapstructure:",squash"`
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *BoltDb) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.storeType = store.BOLTDB
	boltdb.Register()
	return provider.provide(configurationChan)
}
