package aggregator

import (
	"encoding/json"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/provider"
	"github.com/containous/traefik/pkg/provider/file"
	"github.com/containous/traefik/pkg/safe"
)

// ProviderAggregator aggregates providers.
type ProviderAggregator struct {
	fileProvider *file.Provider
	providers    []provider.Provider
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

	if conf.Kubernetes != nil {
		p.quietAddProvider(conf.Kubernetes)
	}

	if conf.KubernetesCRD != nil {
		p.quietAddProvider(conf.KubernetesCRD)
	}
	if conf.Rancher != nil {
		p.quietAddProvider(conf.Rancher)
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

	if fileProvider, ok := provider.(*file.Provider); ok {
		p.fileProvider = fileProvider
	} else {
		p.providers = append(p.providers, provider)
	}
	return nil
}

// Init the provider
func (p ProviderAggregator) Init() error {
	return nil
}

// Provide calls the provide method of every providers
func (p ProviderAggregator) Provide(configurationChan chan<- config.Message, pool *safe.Pool) error {
	if p.fileProvider != nil {
		launchProvider(configurationChan, pool, p.fileProvider)
	}

	for _, prd := range p.providers {
		prd := prd
		safe.Go(func() {
			launchProvider(configurationChan, pool, prd)
		})
	}
	return nil
}

func launchProvider(configurationChan chan<- config.Message, pool *safe.Pool, prd provider.Provider) {
	jsonConf, err := json.Marshal(prd)
	if err != nil {
		log.WithoutContext().Debugf("Cannot marshal the provider configuration %T: %v", prd, err)
	}

	log.WithoutContext().Infof("Starting provider %T %s", prd, jsonConf)

	currentProvider := prd
	err = currentProvider.Provide(configurationChan, pool)
	if err != nil {
		log.WithoutContext().Errorf("Cannot start the provider %T: %v", prd, err)
	}
}
