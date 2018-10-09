package aggregator

import (
	"encoding/json"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// ProviderAggregator aggregates providers.
type ProviderAggregator struct {
	providers   []provider.Provider
	constraints types.Constraints
}

// NewProviderAggregator returns an aggregate of all the providers configured in the static configuration.
func NewProviderAggregator(conf static.Configuration) ProviderAggregator {
	p := ProviderAggregator{
		constraints: conf.Constraints,
	}

	if conf.File != nil {
		p.quietAddProvider(conf.File)
	}

	return p
}

func (p *ProviderAggregator) quietAddProvider(provider provider.Provider) {
	err := p.AddProvider(provider)
	if err != nil {
		log.WithoutContext().Errorf("Error while initializing provider %T: %v", provider, err)
	}
}

// AddProvider adds a provider in the providers map.
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

// Provide calls the provide method of every providers
func (p ProviderAggregator) Provide(configurationChan chan<- config.Message, pool *safe.Pool) error {
	for _, prd := range p.providers {
		jsonConf, err := json.Marshal(prd)
		if err != nil {
			log.WithoutContext().Debugf("Cannot marshal the provider configuration %T: %v", prd, err)
		}
		log.WithoutContext().Infof("Starting provider %T %s", prd, jsonConf)
		currentProvider := prd
		safe.Go(func() {
			err := currentProvider.Provide(configurationChan, pool)
			if err != nil {
				log.WithoutContext().Errorf("Cannot start the provider %T: %v", prd, err)
			}
		})
	}
	return nil
}
