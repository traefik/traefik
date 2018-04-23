package configuration

import (
	"encoding/json"
	"reflect"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// ProviderAggregator aggregate providers
type ProviderAggregator struct {
	providers []provider.Provider
}

// NewProviderAggregator return an aggregate of all the providers configured in GlobalConfiguration
func NewProviderAggregator(gc *GlobalConfiguration) ProviderAggregator {
	provider := ProviderAggregator{}
	if gc.Docker != nil {
		provider.providers = append(provider.providers, gc.Docker)
	}
	if gc.Marathon != nil {
		provider.providers = append(provider.providers, gc.Marathon)
	}
	if gc.File != nil {
		provider.providers = append(provider.providers, gc.File)
	}
	if gc.Rest != nil {
		provider.providers = append(provider.providers, gc.Rest)
	}
	if gc.Consul != nil {
		provider.providers = append(provider.providers, gc.Consul)
	}
	if gc.ConsulCatalog != nil {
		provider.providers = append(provider.providers, gc.ConsulCatalog)
	}
	if gc.Etcd != nil {
		provider.providers = append(provider.providers, gc.Etcd)
	}
	if gc.Zookeeper != nil {
		provider.providers = append(provider.providers, gc.Zookeeper)
	}
	if gc.Boltdb != nil {
		provider.providers = append(provider.providers, gc.Boltdb)
	}
	if gc.Kubernetes != nil {
		provider.providers = append(provider.providers, gc.Kubernetes)
	}
	if gc.Mesos != nil {
		provider.providers = append(provider.providers, gc.Mesos)
	}
	if gc.Eureka != nil {
		provider.providers = append(provider.providers, gc.Eureka)
	}
	if gc.ECS != nil {
		provider.providers = append(provider.providers, gc.ECS)
	}
	if gc.Rancher != nil {
		provider.providers = append(provider.providers, gc.Rancher)
	}
	if gc.DynamoDB != nil {
		provider.providers = append(provider.providers, gc.DynamoDB)
	}
	if gc.ServiceFabric != nil {
		provider.providers = append(provider.providers, gc.ServiceFabric)
	}
	return provider
}

// AddProvider add a provider in the providers map
func (p *ProviderAggregator) AddProvider(provider provider.Provider) {
	p.providers = append(p.providers, provider)
}

// Provide call the provide method of every providers
func (p ProviderAggregator) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	for _, p := range p.providers {
		providerType := reflect.TypeOf(p)
		jsonConf, err := json.Marshal(p)
		if err != nil {
			log.Debugf("Unable to marshal provider conf %v with error: %v", providerType, err)
		}
		log.Infof("Starting provider %v %s", providerType, jsonConf)
		currentProvider := p
		safe.Go(func() {
			err := currentProvider.Provide(configurationChan, pool, constraints)
			if err != nil {
				log.Errorf("Error starting provider %v: %s", providerType, err)
			}
		})
	}
	return nil
}
