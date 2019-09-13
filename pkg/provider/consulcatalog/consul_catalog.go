package consulcatalog

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v3"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/job"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/types"
	"github.com/hashicorp/consul/api"
	"time"
)

type consulCatalog interface {
	Service(string, string, *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
	Services(*api.QueryOptions) (map[string][]string, *api.QueryMeta, error)
}

type prefixes struct {
	enabled           string
	protocol          string
	routerRule        string
	routerEntrypoints string
}

type EndpointHttpAuthConfig struct {
	Username string `description:"Basic Auth username" json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty" export:"true"`
	Password string `description:"Basic Auth password" json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" export:"true"`
}

type EndpointTLSConfig struct {
	Address            string `description:"Address of the Consul server" json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty" export:"true"`
	CAFile             string `description:"Path to the CA certificate used for Consul communication, defaults to the system bundle if not specified" json:"cafile,omitempty" toml:"cafile,omitempty" yaml:"cafile,omitempty" export:"true"`
	CAPath             string `description:"Path to a directory of CA certificates to use for Consul communication, defaults to the system bundle if not specified" json:"capath,omitempty" toml:"capath,omitempty" yaml:"capath,omitempty" export:"true"`
	CertFile           string `description:"Path to the certificate for Consul communication. If this is set then you need to also set KeyFile" json:"certfile,omitempty" toml:"certfile,omitempty" yaml:"certfile,omitempty" export:"true"`
	KeyFile            string `description:"Path to the private key for Consul communication. If this is set then you need to also set CertFile" json:"keyfile,omitempty" toml:"keyfile,omitempty" yaml:"keyfile,omitempty" export:"true"`
	InsecureSkipVerify bool   `description:"InsecureSkipVerify if set to true will disable TLS host verification" json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
}

type EndpointConfig struct {
	Address          string                 `description:"The address of the Consul server" json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty" export:"true"`
	Scheme           string                 `description:"The URI scheme for the Consul server" json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty" export:"true"`
	Datacenter       string                 `description:"Datacenter to use. If not provided, the default agent datacenter is used" json:"datacenter,omitempty" toml:"datacenter,omitempty" yaml:"datacenter,omitempty" export:"true"`
	Token            string                 `description:"Token is used to provide a per-request ACL token which overrides the agent's default token" json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" export:"true"`
	TLS              EndpointTLSConfig      `description:"TLSConfig is used to generate a TLSClientConfig that's useful for talking to Consul using TLS" json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	HttpAuth         EndpointHttpAuthConfig `description:"Auth info to use for http access" json:"httpAuth,omitempty" toml:"httpAuth,omitempty" yaml:"httpAuth,omitempty" export:"true"`
	EndpointWaitTime types.Duration         `description:"WaitTime limits how long a Watch will block. If not provided, the agent default values will be used" json:"endpointWaitTime,omitempty" toml:"endpointWaitTime,omitempty" yaml:"endpointWaitTime,omitempty" export:"true"`
}

type Provider struct {
	Endpoint        EndpointConfig `description:"Consul endpoint settings" json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
	Prefix          string         `description:"Prefix for consul service tags. Default 'traefik'" json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
	Entrypoints     []string       `description:"Default entrypoints" json:"entrypoints,omitempty" toml:"entrypoints,omitempty" yaml:"entrypoints,omitempty" export:"true"`
	RouterRule      string         `description:"Default router rule" json:"routerRule,omitempty" toml:"routerRule,omitempty" yaml:"routerRule,omitempty" export:"true"`
	RefreshInterval types.Duration `description:"Interval for check Consul API. Default 100ms" json:"refreshInterval,omitempty" toml:"refreshInterval,omitempty" yaml:"refreshInterval,omitempty" export:"true"`
	Protocol        string         `description:"Default protocol. 'http' or 'tcp' only'" json:"protocol,omitempty" toml:"protocol,omitempty" yaml:"protocol,omitempty" export:"true"`

	prefixes      prefixes
	clientCatalog consulCatalog
}

func (p *Provider) SetDefaults() {
	p.RefreshInterval = types.Duration(time.Millisecond * 100)
	p.Endpoint.Address = "http://127.0.0.1:8500"
	p.Prefix = "traefik"
	p.Entrypoints = []string{"web"}
	p.RouterRule = "Path(`/`)"
	p.Protocol = "http"
}

func (p *Provider) Init() error {
	if err := p.validateConfig(); err != nil {
		return err
	}

	p.prefixes.enabled = p.Prefix + ".enabled"
	p.prefixes.protocol = p.Prefix + ".protocol="
	p.prefixes.routerRule = p.Prefix + ".router.rule="
	p.prefixes.routerEntrypoints = p.Prefix + ".router.entrypoints="

	return nil
}

func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, "consulcatalog"))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			var err error

			client, err := p.createClient()
			if err != nil {
				return fmt.Errorf("error create consul client, %v", err)
			}

			p.clientCatalog = client.Catalog()

			t := time.NewTicker(time.Duration(p.RefreshInterval))

			for {
				select {
				case <-t.C:
					configuration, err := p.buildConfiguration(ctxLog)
					if err != nil {
						logger.Errorf("error building configuration, %v", err)
						return err
					}
					configurationChan <- dynamic.Message{
						ProviderName:  "consulcatalog",
						Configuration: configuration,
					}
				case <-routineCtx.Done():
					t.Stop()
					return nil
				}
			}
		}

		notify := func(err error, time time.Duration) {
			logger.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}

		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxLog), notify)

		if err != nil {
			logger.Errorf("Cannot connect to consulcatalog server %+v", err)
		}
	})

	return nil
}

func (p *Provider) createClient() (*api.Client, error) {

	cfg := api.Config{
		Address:    p.Endpoint.Address,
		Scheme:     p.Endpoint.Scheme,
		Datacenter: p.Endpoint.Datacenter,
		HttpAuth: &api.HttpBasicAuth{
			Username: p.Endpoint.HttpAuth.Username,
			Password: p.Endpoint.HttpAuth.Password,
		},
		WaitTime: time.Duration(p.Endpoint.EndpointWaitTime),
		Token:    p.Endpoint.Token,
		TLSConfig: api.TLSConfig{
			Address:            p.Endpoint.TLS.Address,
			CAFile:             p.Endpoint.TLS.CAFile,
			CAPath:             p.Endpoint.TLS.CAPath,
			CertFile:           p.Endpoint.TLS.CertFile,
			KeyFile:            p.Endpoint.TLS.KeyFile,
			InsecureSkipVerify: p.Endpoint.TLS.InsecureSkipVerify,
		},
	}

	return api.NewClient(&cfg)
}
