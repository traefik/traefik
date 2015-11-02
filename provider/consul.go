package provider

import "github.com/emilevauge/traefik/types"

type Consul struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *Kv
}

func (provider *Consul) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewConsulProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
