package main

type ZookepperProvider struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *KvProvider
}

func (provider *ZookepperProvider) Provide(configurationChan chan<- configMessage) error {
	provider.KvProvider = NewZkProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
