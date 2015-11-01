package provider

import "github.com/emilevauge/traefik/types"

type ZookepperProvider struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *KvProvider
}

func (provider *ZookepperProvider) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewZkProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
