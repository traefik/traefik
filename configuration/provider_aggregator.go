package configuration

import (
	"encoding/json"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// ProviderAggregator aggregate providers
type ProviderAggregator struct {
	providers   []provider.Provider
	constraints types.Constraints
}

// NewProviderAggregator return an aggregate of all the providers configured in GlobalConfiguration
func NewProviderAggregator(gc *GlobalConfiguration) ProviderAggregator {
	provider := ProviderAggregator{
		constraints: gc.Constraints,
	}
	if gc.Docker != nil {
		provider.quietAddProvider(gc.Docker)
	}
	if gc.Marathon != nil {
		provider.quietAddProvider(gc.Marathon)
	}
	if gc.File != nil {
		provider.quietAddProvider(gc.File)
	}
	if gc.Rest != nil {
		provider.quietAddProvider(gc.Rest)
	}
	if gc.Consul != nil {
		provider.quietAddProvider(gc.Consul)
	}
	if gc.ConsulCatalog != nil {
		provider.quietAddProvider(gc.ConsulCatalog)
	}
	if gc.Etcd != nil {
		provider.quietAddProvider(gc.Etcd)
	}
	if gc.Zookeeper != nil {
		provider.quietAddProvider(gc.Zookeeper)
	}
	if gc.Boltdb != nil {
		provider.quietAddProvider(gc.Boltdb)
	}
	if gc.Kubernetes != nil {
		provider.quietAddProvider(gc.Kubernetes)
	}
	if gc.Mesos != nil {
		provider.quietAddProvider(gc.Mesos)
	}
	if gc.Eureka != nil {
		provider.quietAddProvider(gc.Eureka)
	}
	if gc.ECS != nil {
		provider.quietAddProvider(gc.ECS)
	}
	if gc.Rancher != nil {
		provider.quietAddProvider(gc.Rancher)
	}
	if gc.DynamoDB != nil {
		provider.quietAddProvider(gc.DynamoDB)
	}
	if gc.ServiceFabric != nil {
		provider.quietAddProvider(gc.ServiceFabric)
	}
	return provider
}

func (p *ProviderAggregator) quietAddProvider(provider provider.Provider) {
	err := p.AddProvider(provider)
	if err != nil {
		log.Errorf("Error initializing provider %T: %v", provider, err)
	}
}

// AddProvider add a provider in the providers map
func (p *ProviderAggregator) AddProvider(provider provider.Provider) error {
	err := provider.Init(p.constraints)
	if err != nil {
		return err
	}
	p.providers = append(p.providers, provider)
	return nil
}

// Init the provider
func (p ProviderAggregator) Init(_ types.Constraints) error {
	return nil
}

// Provide call the provide method of every providers
func (p ProviderAggregator) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	for _, p := range p.providers {
		jsonConf, err := json.Marshal(p)
		if err != nil {
			log.Debugf("Unable to marshal provider conf %T with error: %v", p, err)
		}
		log.Infof("Starting provider %T %s", p, jsonConf)
		currentProvider := p
		safe.Go(func() {
			err := currentProvider.Provide(configurationChan, pool)
			if err != nil {
				log.Errorf("Error starting provider %T: %v", p, err)
			}
		})
	}
	return nil
}
