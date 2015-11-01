package provider

import "github.com/emilevauge/traefik/types"

type EtcdProvider struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *KvProvider
}

func (provider *EtcdProvider) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewEtcdProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
