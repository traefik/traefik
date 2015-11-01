package provider

import "github.com/emilevauge/traefik/types"

// Consul holds configurations of the Consul provider.
type Consul struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *Kv
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Consul) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewConsulProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
