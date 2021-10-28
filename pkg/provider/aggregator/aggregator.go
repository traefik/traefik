package aggregator

import (
	"context"
	"encoding/json"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/file"
	"github.com/traefik/traefik/v2/pkg/provider/traefik"
	"github.com/traefik/traefik/v2/pkg/safe"
)

type throttledProvide func(configurationChan chan<- dynamic.Message, pool *safe.Pool) error

func throttleProvide(provider provider.Provider, providerThrottleDuration time.Duration) throttledProvide {
	return func(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
		rc := NewRingChannel()
		pool.GoCtx(func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-rc.Out():
					configurationChan <- msg
					time.Sleep(providerThrottleDuration)
				}
			}
		})

		return provider.Provide(rc.In(), pool)
	}
}

// ProviderAggregator aggregates providers.
type ProviderAggregator struct {
	providersThrottleDuration time.Duration
	internalProvider          provider.Provider
	fileProvider              provider.Provider
	providers                 []provider.Provider
}

// NewProviderAggregator returns an aggregate of all the providers configured in the static configuration.
func NewProviderAggregator(conf static.Providers) ProviderAggregator {
	p := ProviderAggregator{
		providersThrottleDuration: time.Duration(conf.ProvidersThrottleDuration),
	}

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

	if conf.KubernetesIngress != nil {
		p.quietAddProvider(conf.KubernetesIngress)
	}

	if conf.KubernetesCRD != nil {
		p.quietAddProvider(conf.KubernetesCRD)
	}

	if conf.KubernetesGateway != nil {
		p.quietAddProvider(conf.KubernetesGateway)
	}

	if conf.Rancher != nil {
		p.quietAddProvider(conf.Rancher)
	}

	if conf.Ecs != nil {
		p.quietAddProvider(conf.Ecs)
	}

	if conf.ConsulCatalog != nil {
		p.quietAddProvider(conf.ConsulCatalog)
	}

	if conf.Consul != nil {
		p.quietAddProvider(conf.Consul)
	}

	if conf.Etcd != nil {
		p.quietAddProvider(conf.Etcd)
	}

	if conf.ZooKeeper != nil {
		p.quietAddProvider(conf.ZooKeeper)
	}

	if conf.Redis != nil {
		p.quietAddProvider(conf.Redis)
	}

	if conf.HTTP != nil {
		p.quietAddProvider(conf.HTTP)
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

	switch provider.(type) {
	case *file.Provider:
		p.fileProvider = provider
	case *traefik.Provider:
		p.internalProvider = provider
	default:
		p.providers = append(p.providers, provider)
	}

	return nil
}

// Init the provider.
func (p ProviderAggregator) Init() error {
	return nil
}

// ThrottleDuration returns the throttle duration.
func (p ProviderAggregator) ThrottleDuration() *time.Duration {
	return nil
}

// Provide calls the provide method of every providers.
func (p ProviderAggregator) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	if p.fileProvider != nil {
		p.launchProvider(configurationChan, pool, p.fileProvider)
	}

	for _, prd := range p.providers {
		prd := prd
		safe.Go(func() {
			p.launchProvider(configurationChan, pool, prd)
		})
	}

	// internal provider must be the last because we use it to know if all the providers are loaded.
	// ConfigurationWatcher will wait for this requiredProvider before applying configurations.
	if p.internalProvider != nil {
		p.launchProvider(configurationChan, pool, p.internalProvider)
	}

	return nil
}

func (p ProviderAggregator) launchProvider(configurationChan chan<- dynamic.Message, pool *safe.Pool, prd provider.Provider) {
	jsonConf, err := json.Marshal(prd)
	if err != nil {
		log.WithoutContext().Debugf("Cannot marshal the provider configuration %T: %v", prd, err)
	}

	log.WithoutContext().Infof("Starting provider %T %s", prd, jsonConf)

	// Default throttling
	if prd.ThrottleDuration() == nil {
		err = throttleProvide(prd, p.providersThrottleDuration)(configurationChan, pool)
		if err != nil {
			log.WithoutContext().Errorf("Cannot start the provider %T: %v", prd, err)
		}

		return
	}

	// Custom throttling
	if *prd.ThrottleDuration() != 0 {
		err = throttleProvide(prd, *prd.ThrottleDuration())(configurationChan, pool)
		if err != nil {
			log.WithoutContext().Errorf("Cannot start the provider %T: %v", prd, err)
		}

		return
	}

	// Throttling deactivation
	err = prd.Provide(configurationChan, pool)
	if err != nil {
		log.WithoutContext().Errorf("Cannot start the provider %T: %v", prd, err)
	}
}
