package provider

import "github.com/emilevauge/traefik/types"

type ConsulProvider struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *KvProvider
}

func (provider *ConsulProvider) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewConsulProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
