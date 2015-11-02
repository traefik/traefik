package provider

import "github.com/emilevauge/traefik/types"

// Etcd holds configurations of the Etcd provider.
type Etcd struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *Kv
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Etcd) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewEtcdProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
