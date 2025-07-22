package kv

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/kv"
	"github.com/traefik/traefik/v3/pkg/job"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/safe"
)

// Provider holds configurations of the provider.
type Provider struct {
	RootKey string `description:"Root key used for KV store." json:"rootKey,omitempty" toml:"rootKey,omitempty" yaml:"rootKey,omitempty"`

	Endpoints []string `description:"KV store endpoints." json:"endpoints,omitempty" toml:"endpoints,omitempty" yaml:"endpoints,omitempty"`

	name     string
	kvClient store.Store
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.RootKey = "traefik"
}

// Init the provider.
func (p *Provider) Init(storeType, name string, config valkeyrie.Config) error {
	ctx := log.With().Str(logs.ProviderName, name).Logger().WithContext(context.Background())

	p.name = name

	kvClient, err := p.createKVClient(ctx, storeType, config)
	if err != nil {
		return fmt.Errorf("failed to Connect to KV store: %w", err)
	}

	p.kvClient = kvClient

	return nil
}

// Provide allows the docker provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, p.name).Logger()
	ctx := logger.WithContext(context.Background())

	operation := func() error {
		if _, err := p.kvClient.Exists(ctx, path.Join(p.RootKey, "qmslkjdfmqlskdjfmqlksjazçueznbvbwzlkajzebvkwjdcqmlsfj"), nil); err != nil {
			return fmt.Errorf("KV store connection error: %w", err)
		}
		return nil
	}

	notify := func(err error, time time.Duration) {
		logger.Error().Err(err).Msgf("KV connection error, retrying in %s", time)
	}

	err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctx), notify)
	if err != nil {
		return fmt.Errorf("cannot connect to KV server: %w", err)
	}

	configuration, err := p.buildConfiguration(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Cannot build the configuration")
	} else {
		configurationChan <- dynamic.Message{
			ProviderName:  p.name,
			Configuration: configuration,
		}
	}

	pool.GoCtx(func(ctxPool context.Context) {
		ctxLog := logger.With().Str(logs.ProviderName, p.name).Logger().WithContext(ctxPool)

		err := p.watchKv(ctxLog, configurationChan)
		if err != nil {
			logger.Error().Err(err).Msg("Cannot retrieve data")
		}
	})

	return nil
}

func (p *Provider) watchKv(ctx context.Context, configurationChan chan<- dynamic.Message) error {
	operation := func() error {
		events, err := p.kvClient.WatchTree(ctx, p.RootKey, nil)
		if err != nil {
			return fmt.Errorf("failed to watch KV: %w", err)
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case _, ok := <-events:
				if !ok {
					return errors.New("the WatchTree channel is closed")
				}

				configuration, errC := p.buildConfiguration(ctx)
				if errC != nil {
					return errC
				}

				if configuration != nil {
					configurationChan <- dynamic.Message{
						ProviderName:  p.name,
						Configuration: configuration,
					}
				}
			}
		}
	}

	notify := func(err error, time time.Duration) {
		log.Ctx(ctx).Error().Err(err).Msgf("Provider error, retrying in %s", time)
	}

	return backoff.RetryNotify(safe.OperationWithRecover(operation),
		backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctx), notify)
}

func (p *Provider) buildConfiguration(ctx context.Context) (*dynamic.Configuration, error) {
	pairs, err := p.kvClient.List(ctx, p.RootKey, nil)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			// This empty configuration satisfies the pkg/server/configurationwatcher.go isEmptyConfiguration func constraints,
			// and will not be discarded by the configuration watcher.
			return &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: make(map[string]*dynamic.Router),
				},
			}, nil
		}

		return nil, err
	}

	cfg := &dynamic.Configuration{}
	err = kv.Decode(pairs, cfg, p.RootKey)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (p *Provider) createKVClient(ctx context.Context, storeType string, config valkeyrie.Config) (store.Store, error) {
	kvStore, err := valkeyrie.NewStore(ctx, storeType, p.Endpoints, config)
	if err != nil {
		return nil, err
	}

	return &storeWrapper{Store: kvStore}, nil
}
