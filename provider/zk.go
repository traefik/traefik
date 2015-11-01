package provider

import "github.com/emilevauge/traefik/types"

// Zookepper holds configurations of the Zookepper provider.
type Zookepper struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *Kv
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Zookepper) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewZkProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
