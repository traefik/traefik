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
		provider.initAndAddProvider(gc.Docker, gc.Constraints)
	}
	if gc.Marathon != nil {
		provider.initAndAddProvider(gc.Marathon, gc.Constraints)
	}
	if gc.File != nil {
		provider.initAndAddProvider(gc.File, gc.Constraints)
	}
	if gc.Rest != nil {
		provider.initAndAddProvider(gc.Rest, gc.Constraints)
	}
	if gc.Consul != nil {
		provider.initAndAddProvider(gc.Consul, gc.Constraints)
	}
	if gc.ConsulCatalog != nil {
		provider.initAndAddProvider(gc.ConsulCatalog, gc.Constraints)
	}
	if gc.Etcd != nil {
		provider.initAndAddProvider(gc.Etcd, gc.Constraints)
	}
	if gc.Zookeeper != nil {
		provider.initAndAddProvider(gc.Zookeeper, gc.Constraints)
	}
	if gc.Boltdb != nil {
		provider.initAndAddProvider(gc.Boltdb, gc.Constraints)
	}
	if gc.Kubernetes != nil {
		provider.initAndAddProvider(gc.Kubernetes, gc.Constraints)
	}
	if gc.Mesos != nil {
		provider.initAndAddProvider(gc.Mesos, gc.Constraints)
	}
	if gc.Eureka != nil {
		provider.initAndAddProvider(gc.Eureka, gc.Constraints)
	}
	if gc.ECS != nil {
		provider.initAndAddProvider(gc.ECS, gc.Constraints)
	}
	if gc.Rancher != nil {
		provider.initAndAddProvider(gc.Rancher, gc.Constraints)
	}
	if gc.DynamoDB != nil {
		provider.initAndAddProvider(gc.DynamoDB, gc.Constraints)
	}
	if gc.ServiceFabric != nil {
		provider.initAndAddProvider(gc.ServiceFabric, gc.Constraints)
	}
	return provider
}

func (p *ProviderAggregator) initAndAddProvider(provider provider.Provider, constraint types.Constraints) {
	err := provider.Init(constraint)
	if err != nil {
		providerType := reflect.TypeOf(provider)
		log.Errorf("Error initializing provider %v: %v", providerType, err)
	} else {
		p.AddProvider(provider)
	}
}

// AddProvider add a provider in the providers map
func (p *ProviderAggregator) AddProvider(provider provider.Provider) {
	p.providers = append(p.providers, provider)
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
				log.Errorf("Error starting provider %T: %s", p, err)
			}
		})
	}
	return nil
}
