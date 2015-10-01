package main

type BoltDbProvider struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *KvProvider
}

func (provider *BoltDbProvider) Provide(configurationChan chan<- configMessage) error {
	provider.KvProvider = NewBoltDbProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
