package provider

import "github.com/emilevauge/traefik/types"

// BoltDb holds configurations of the BoltDb provider.
type BoltDb struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *Kv
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *BoltDb) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewBoltDbProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
