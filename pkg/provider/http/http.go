package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/job"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/safe"
)

var _ provider.Provider = (*Provider)(nil)

// Provider is a provider.Provider implementation that queries an endpoint for a configuration.
type Provider struct {
	endpoint         string
	endpointInsecure bool
	pollInterval     time.Duration
	pollTimeout      time.Duration
}

// Init the provider.
func (p *Provider) Init() error {
	if len(p.endpoint) > 0 {
		return nil
	}

	return fmt.Errorf("a non-empty endpoint is required")
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, "http"))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			ctx, cancel := context.WithCancel(ctxLog)
			defer cancel()

			ctx = log.With(ctx, log.Str(log.ProviderName, "http"))
			errChan := make(chan error)
			ticker := time.NewTicker(time.Duration(p.pollInterval))

			pool.GoCtx(func(ctx context.Context) {
				ctx = log.With(ctx, log.Str(log.ProviderName, "http"))
				logger := log.FromContext(ctx)

				defer close(errChan)
				for {
					select {
					case <-ticker.C:
						data, err := p.getDataFromEndpoint(ctxLog)
						if err != nil {
							logger.Errorf("Failed to get config from endpoint: %w", err)
							errChan <- err
							return
						}

						configuration := p.buildConfiguration(ctx, data)
						if configuration != nil {
							configurationChan <- dynamic.Message{
								ProviderName:  "http",
								Configuration: configuration,
							}
						}

					case <-ctx.Done():
						ticker.Stop()
						return
					}
				}
			})
			if err, ok := <-errChan; ok {
				return err
			}
			// channel closed
			return nil
		}

		notify := func(err error, time time.Duration) {
			logger.Errorf("Provider connection error %w, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxLog), notify)
		if err != nil {
			logger.Errorf("Cannot connect to http server %w", err)
		}
	})

	return nil
}

// getDataFromEndpoint gets data from the configured provider endpoint, and returns the data.
func (p *Provider) getDataFromEndpoint(ctx context.Context) ([]byte, error) {
	b := []byte{}
	req, err := http.NewRequest(http.MethodGet, p.endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	client := &http.Client{Timeout: p.pollTimeout}
	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()

		var bodyData []byte
		var bodyErr error
		if bodyData, bodyErr = ioutil.ReadAll(resp.Body); bodyErr != nil {
			return nil, fmt.Errorf("unable to read response body: %w", bodyErr)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-ok response code: %d", resp.StatusCode)
		}

		log.FromContext(ctx).Debugf("Successfully received data from endpoint: %q", p.endpoint)
		return bodyData, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to get data from endpoint: %w", err)
	}

	return nil, fmt.Errorf("recieved no data from endpoint")
}

// buildConfiguration builds a configuration from the provided data.
func (p *Provider) buildConfiguration(ctx context.Context, data []byte) *dynamic.Configuration {
	configuration := &dynamic.Configuration{}

	if err := json.Unmarshal(data, configuration); err != nil {
		log.FromContext(ctx).Errorf("Error parsing configuration %w", err)
		return nil
	}

	return configuration
}
