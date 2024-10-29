package http

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/file"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/job"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

var _ provider.Provider = (*Provider)(nil)

// Provider is a provider.Provider implementation that queries an HTTP(s) endpoint for a configuration.
type Provider struct {
	Endpoint     string            `description:"Load configuration from this endpoint." json:"endpoint" toml:"endpoint" yaml:"endpoint"`
	PollInterval ptypes.Duration   `description:"Polling interval for endpoint." json:"pollInterval,omitempty" toml:"pollInterval,omitempty" yaml:"pollInterval,omitempty" export:"true"`
	PollTimeout  ptypes.Duration   `description:"Polling timeout for endpoint." json:"pollTimeout,omitempty" toml:"pollTimeout,omitempty" yaml:"pollTimeout,omitempty" export:"true"`
	Headers      map[string]string `description:"Define custom headers to be sent to the endpoint." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
	TLS          *types.ClientTLS  `description:"Enable TLS support." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`

	httpClient            *http.Client
	lastConfigurationHash uint64
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.PollInterval = ptypes.Duration(5 * time.Second)
	p.PollTimeout = ptypes.Duration(5 * time.Second)
}

// Init the provider.
func (p *Provider) Init() error {
	if p.Endpoint == "" {
		return errors.New("non-empty endpoint is required")
	}

	if p.PollInterval <= 0 {
		return errors.New("poll interval must be greater than 0")
	}

	p.httpClient = &http.Client{
		Timeout: time.Duration(p.PollTimeout),
	}

	if p.TLS != nil {
		tlsConfig, err := p.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return fmt.Errorf("unable to create client TLS configuration: %w", err)
		}

		p.httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	return nil
}

// Provide allows the provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		logger := log.Ctx(routineCtx).With().Str(logs.ProviderName, "http").Logger()
		ctxLog := logger.WithContext(routineCtx)

		operation := func() error {
			if err := p.updateConfiguration(configurationChan); err != nil {
				return err
			}

			ticker := time.NewTicker(time.Duration(p.PollInterval))
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := p.updateConfiguration(configurationChan); err != nil {
						return err
					}

				case <-routineCtx.Done():
					return nil
				}
			}
		}

		notify := func(err error, time time.Duration) {
			logger.Error().Err(err).Msgf("Provider error, retrying in %s", time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxLog), notify)
		if err != nil {
			logger.Error().Err(err).Msg("Cannot retrieve data")
		}
	})

	return nil
}

func (p *Provider) updateConfiguration(configurationChan chan<- dynamic.Message) error {
	configData, err := p.fetchConfigurationData()
	if err != nil {
		return fmt.Errorf("cannot fetch configuration data: %w", err)
	}

	fnvHasher := fnv.New64()

	if _, err = fnvHasher.Write(configData); err != nil {
		return fmt.Errorf("cannot hash configuration data: %w", err)
	}

	hash := fnvHasher.Sum64()
	if hash == p.lastConfigurationHash {
		return nil
	}

	p.lastConfigurationHash = hash

	configuration, err := decodeConfiguration(configData)
	if err != nil {
		return fmt.Errorf("cannot decode configuration data: %w", err)
	}

	configurationChan <- dynamic.Message{
		ProviderName:  "http",
		Configuration: configuration,
	}

	return nil
}

// fetchConfigurationData fetches the configuration data from the configured endpoint.
func (p *Provider) fetchConfigurationData() ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, p.Endpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create fetch request: %w", err)
	}

	for k, v := range p.Headers {
		if strings.EqualFold(k, "Host") {
			req.Host = v
		} else {
			req.Header.Set(k, v)
		}
	}

	res, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do fetch request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-ok response code: %d", res.StatusCode)
	}

	return io.ReadAll(res.Body)
}

// decodeConfiguration decodes and returns the dynamic configuration from the given data.
func decodeConfiguration(data []byte) (*dynamic.Configuration, error) {
	configuration := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
	}

	err := file.DecodeContent(string(data), ".yaml", configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}
