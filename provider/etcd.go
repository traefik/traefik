package provider

import "github.com/emilevauge/traefik/types"

type Etcd struct {
	Watch      bool
	Endpoint   string
	Prefix     string
	Filename   string
	KvProvider *Kv
}

func (provider *Etcd) Provide(configurationChan chan<- types.ConfigMessage) error {
	provider.KvProvider = NewEtcdProvider(provider)
	return provider.KvProvider.provide(configurationChan)
}
