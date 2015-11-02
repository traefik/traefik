package provider

import "github.com/emilevauge/traefik/types"

type BoltDb struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *Kv
}

func (provider *BoltDb) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewBoltDbProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
