package consulcatalog

import (
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/job"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/types"
	"github.com/hashicorp/consul/api"
)

// DefaultTemplateRule The default template for the default rule.
const DefaultTemplateRule = "Host(`{{ normalize .Name }}`)"

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	Constraints      string          `description:"Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container." json:"constraints,omitempty" toml:"constraints,omitempty" yaml:"constraints,omitempty" export:"true"`
	Endpoint         *EndpointConfig `description:"Consul endpoint settings" json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
	Prefix           string          `description:"Prefix for consul service tags. Default 'traefik'" json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
	RefreshInterval  types.Duration  `description:"Interval for check Consul API. Default 100ms" json:"refreshInterval,omitempty" toml:"refreshInterval,omitempty" yaml:"refreshInterval,omitempty" export:"true"`
	ExposedByDefault bool            `description:"Expose containers by default." json:"exposedByDefault,omitempty" toml:"exposedByDefault,omitempty" yaml:"exposedByDefault,omitempty" export:"true"`
	DefaultRule      string          `description:"Default rule." json:"defaultRule,omitempty" toml:"defaultRule,omitempty" yaml:"defaultRule,omitempty"`

	client         *api.Client
	defaultRuleTpl *template.Template
}

// EndpointConfig holds configurations of the endpoint.
type EndpointConfig struct {
	Address          string                 `description:"The address of the Consul server" json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty" export:"true"`
	Scheme           string                 `description:"The URI scheme for the Consul server" json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty" export:"true"`
	Datacenter       string                 `description:"Datacenter to use. If not provided, the default agent datacenter is used" json:"datacenter,omitempty" toml:"datacenter,omitempty" yaml:"datacenter,omitempty" export:"true"`
	Token            string                 `description:"Token is used to provide a per-request ACL token which overrides the agent's default token" json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" export:"true"`
	TLS              *types.ClientTLS       `description:"Enable TLS support." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	HttpAuth         EndpointHttpAuthConfig `description:"Auth info to use for http access" json:"httpAuth,omitempty" toml:"httpAuth,omitempty" yaml:"httpAuth,omitempty" export:"true"`
	EndpointWaitTime types.Duration         `description:"WaitTime limits how long a Watch will block. If not provided, the agent default values will be used" json:"endpointWaitTime,omitempty" toml:"endpointWaitTime,omitempty" yaml:"endpointWaitTime,omitempty" export:"true"`
}

// EndpointHttpAuthConfig holds configurations of the authentication.
type EndpointHttpAuthConfig struct {
	Username string `description:"Basic Auth username" json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty" export:"true"`
	Password string `description:"Basic Auth password" json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" export:"true"`
}

// SetDefaults sets the default values.

func (p *Provider) SetDefaults() {
	p.Endpoint = &EndpointConfig{
		Address: "http://127.0.0.1:8500",
	}
	p.RefreshInterval = types.Duration(15 * time.Second)
	p.Prefix = "traefik"
	p.ExposedByDefault = true
	p.DefaultRule = DefaultTemplateRule
}

// Init the provider.
func (p *Provider) Init() error {
	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %v", err)
	}

	p.defaultRuleTpl = defaultRuleTpl
	return nil
}

type itemData struct {
	ID      string
	Name    string
	Address string
	Port    string
	Enable  bool
	Status  string
	Labels  map[string]string
}

func createClient(cfg *EndpointConfig) (*api.Client, error) {
	config := api.Config{
		Address:    cfg.Address,
		Scheme:     cfg.Scheme,
		Datacenter: cfg.Datacenter,
		HttpAuth: &api.HttpBasicAuth{
			Username: cfg.HttpAuth.Username,
			Password: cfg.HttpAuth.Password,
		},
		WaitTime: time.Duration(cfg.EndpointWaitTime),
		Token:    cfg.Token,
		TLSConfig: api.TLSConfig{
			Address:            cfg.Address,
			CAFile:             cfg.TLS.CA,
			CertFile:           cfg.TLS.Cert,
			KeyFile:            cfg.TLS.Key,
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		},
	}

	return api.NewClient(&config)
}

func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, "consulcatalog"))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			var err error

			p.client, err = createClient(p.Endpoint)
			if err != nil {
				return fmt.Errorf("error create consul client, %v", err)
			}

			t := time.NewTicker(time.Duration(p.RefreshInterval))

			for {
				select {
				case <-t.C:
					data, err := p.getConsulServicesData(routineCtx)
					if err != nil {
						logger.Errorf("error get consulCatalog data, %v", err)
						return err
					}

					configuration := p.buildConfiguration(routineCtx, data)
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

func (p *Provider) getConsulServicesData(ctx context.Context) ([]itemData, error) {
	var data []itemData

	consulServiceNames, err := p.fetchServices(ctx)
	if err != nil {
		return nil, err
	}

	for name := range consulServiceNames {
		consulServices, err := p.fetchService(ctx, name)
		if err != nil {
			return nil, err
		}

		for _, consulService := range consulServices {
			labels := tagsToNeutralLabels(consulService.ServiceTags, p.Prefix)
			item := itemData{
				ID:      consulService.ServiceID,
				Name:    consulService.ServiceName,
				Address: consulService.ServiceAddress,
				Port:    string(consulService.ServicePort),
				Labels:  labels,
			}

			data = append(data, item)
		}
	}
	return data, nil
}

func (p *Provider) fetchService(ctx context.Context, name string) ([]*api.CatalogService, error) {
	var tagFilter string
	if !p.ExposedByDefault {
		tagFilter = p.Prefix + ".enable=true"
	}

	consulServices, _, err := p.client.Catalog().Service(name, tagFilter, &api.QueryOptions{})
	return consulServices, err
}

func (p *Provider) fetchServices(ctx context.Context) (map[string][]string, error) {
	serviceNames, _, err := p.client.Catalog().Services(&api.QueryOptions{})
	return serviceNames, err
}
