package main

type EtcdProvider struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *KvProvider
}

func (provider *EtcdProvider) Provide(configurationChan chan<- configMessage) error {
	provider.KvProvider = NewEtcdProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
