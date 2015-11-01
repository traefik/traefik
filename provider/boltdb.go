package provider

import "github.com/emilevauge/traefik/types"

type BoltDbProvider struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *KvProvider
}

func (provider *BoltDbProvider) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewBoltDbProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
