package provider

import "github.com/emilevauge/traefik/types"

type Zookepper struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *Kv
}

func (provider *Zookepper) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewZkProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
