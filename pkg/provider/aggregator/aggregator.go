package aggregator

import (
	"context"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/file"
	"github.com/traefik/traefik/v2/pkg/provider/traefik"
	"github.com/traefik/traefik/v2/pkg/redactor"
	"github.com/traefik/traefik/v2/pkg/safe"
)

// throttled defines what kind of config refresh throttling the aggregator should
// set up for a given provider.
// If a provider implements throttled, the configuration changes it sends will be
// taken into account no more often than the frequency inferred from ThrottleDuration().
// If ThrottleDuration returns zero, no throttling will take place.
// If throttled is not implemented, the throttling will be set up in accordance
// with the global providersThrottleDuration option.
type throttled interface {
	ThrottleDuration() time.Duration
}

// maybeThrottledProvide returns the Provide method of the given provider,
// potentially augmented with some throttling depending on whether and how the
// provider implements the throttled interface.
func maybeThrottledProvide(prd provider.Provider, defaultDuration time.Duration) func(chan<- dynamic.Message, *safe.Pool) error {
	providerThrottleDuration := defaultDuration
	if throttled, ok := prd.(throttled); ok {
		// per-provider throttling
		providerThrottleDuration = throttled.ThrottleDuration()
	}

	if providerThrottleDuration == 0 {
		// throttling disabled
		return prd.Provide
	}

	return func(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
		rc := newRingChannel()
		pool.GoCtx(func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-rc.out():
					configurationChan <- msg
					time.Sleep(providerThrottleDuration)
				}
			}
		})

		return prd.Provide(rc.in(), pool)
	}
}

// ProviderAggregator aggregates providers.
type ProviderAggregator struct {
	internalProvider          provider.Provider
	fileProvider              provider.Provider
	providers                 []provider.Provider
	providersThrottleDuration time.Duration
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
		for _, pvd := range conf.ConsulCatalog.BuildProviders() {
			p.quietAddProvider(pvd)
		}
	}

	if conf.Nomad != nil {
		for _, pvd := range conf.Nomad.BuildProviders() {
			p.quietAddProvider(pvd)
		}
	}

	if conf.Consul != nil {
		for _, pvd := range conf.Consul.BuildProviders() {
			p.quietAddProvider(pvd)
		}
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
	jsonConf, err := redactor.RemoveCredentials(prd)
	if err != nil {
		log.WithoutContext().Debugf("Cannot marshal the provider configuration %T: %v", prd, err)
	}

	log.WithoutContext().Infof("Starting provider %T", prd)
	log.WithoutContext().Debugf("%T provider configuration: %s", prd, jsonConf)

	if err := maybeThrottledProvide(prd, p.providersThrottleDuration)(configurationChan, pool); err != nil {
		log.WithoutContext().Errorf("Cannot start the provider %T: %v", prd, err)
		return
	}
}
