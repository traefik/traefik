package aggregator

import (
	"encoding/json"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
)

// ProviderAggregator aggregates providers.
type ProviderAggregator struct {
	providers []provider.Provider
}

// NewProviderAggregator returns an aggregate of all the providers configured in the static configuration.
func NewProviderAggregator(conf static.Providers) ProviderAggregator {
	p := ProviderAggregator{}

	if conf.File != nil {
		p.quietAddProvider(conf.File)
	}

	if conf.Docker != nil {
		p.quietAddProvider(conf.Docker)
	}

	if conf.Marathon != nil {
		p.quietAddProvider(conf.Marathon)
	}

	if conf.Rest != nil {
		p.quietAddProvider(conf.Rest)
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
	err := provider.Init()
	if err != nil {
		return err
	}
	p.providers = append(p.providers, provider)
	return nil
}

// Init the provider
func (p ProviderAggregator) Init() error {
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
