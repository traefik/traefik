package consulcatalog

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"strconv"
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

type itemData struct {
	ID        string
	Node      string
	Name      string
	Address   string
	Port      string
	Status    string
	Labels    map[string]string
	Tags      []string
	ExtraConf configuration
}

// Provider holds configurations of the provider.
type Provider struct {
	Constraints       string          `description:"Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container." json:"constraints,omitempty" toml:"constraints,omitempty" yaml:"constraints,omitempty" export:"true"`
	Endpoint          *EndpointConfig `description:"Consul endpoint settings" json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
	Prefix            string          `description:"Prefix for consul service tags. Default 'traefik'" json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
	RefreshInterval   types.Duration  `description:"Interval for check Consul API. Default 100ms" json:"refreshInterval,omitempty" toml:"refreshInterval,omitempty" yaml:"refreshInterval,omitempty" export:"true"`
	RequireConsistent bool            `description:"Forces the read to be fully consistent." json:"requireConsistent,omitempty" toml:"requireConsistent,omitempty" yaml:"requireConsistent,omitempty" export:"true"`
	Stale             bool            `description:"Use stale consistency for catalog reads." json:"stale,omitempty" toml:"stale,omitempty" yaml:"stale,omitempty" export:"true"`
	Cache             bool            `description:"Use local agent caching for catalog reads." json:"cache,omitempty" toml:"cache,omitempty" yaml:"cache,omitempty" export:"true"`
	ExposedByDefault  bool            `description:"Expose containers by default." json:"exposedByDefault,omitempty" toml:"exposedByDefault,omitempty" yaml:"exposedByDefault,omitempty" export:"true"`
	DefaultRule       string          `description:"Default rule." json:"defaultRule,omitempty" toml:"defaultRule,omitempty" yaml:"defaultRule,omitempty"`
	ConnectAware      bool            `description:"Enable Consul Connect support." json:"connectAware,omitEmpty" toml:"connectAware,omitempty" yaml:"connectAware,omitEmpty"`
	ConnectNative     bool            `description:"Register and manage traefik in Consul Catalog as a Connect Native service." json:"connectNative,omitempty" toml:"connectNative,omitempty" yaml:"connectNative,omitempty"`
	ServiceName       string          `description:"Name of the traefik service in Consul Catalog." json:"serviceName,omitempty" toml:"serviceName,omitempty" yaml:"serviceName,omitempty"`
	ServicePort       int             `description:"Port of the traefik service to register in Consul Catalog" json:"servicePort,omitempty" toml:"servicePort,omitempty" yaml:"servicePort,omitempty"`

	client         *api.Client
	defaultRuleTpl *template.Template
	tlsChan        chan *api.LeafCert
}

// EndpointConfig holds configurations of the endpoint.
type EndpointConfig struct {
	Address          string                  `description:"The address of the Consul server" json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty" export:"true"`
	Scheme           string                  `description:"The URI scheme for the Consul server" json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty" export:"true"`
	DataCenter       string                  `description:"Data center to use. If not provided, the default agent data center is used" json:"datacenter,omitempty" toml:"datacenter,omitempty" yaml:"datacenter,omitempty" export:"true"`
	Token            string                  `description:"Token is used to provide a per-request ACL token which overrides the agent's default token" json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" export:"true"`
	TLS              *types.ClientTLS        `description:"Enable TLS support." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	HTTPAuth         *EndpointHTTPAuthConfig `description:"Auth info to use for http access" json:"httpAuth,omitempty" toml:"httpAuth,omitempty" yaml:"httpAuth,omitempty" export:"true"`
	EndpointWaitTime types.Duration          `description:"WaitTime limits how long a Watch will block. If not provided, the agent default values will be used" json:"endpointWaitTime,omitempty" toml:"endpointWaitTime,omitempty" yaml:"endpointWaitTime,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *EndpointConfig) SetDefaults() {
	c.Address = "http://127.0.0.1:8500"
}

// EndpointHTTPAuthConfig holds configurations of the authentication.
type EndpointHTTPAuthConfig struct {
	Username string `description:"Basic Auth username" json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty" export:"true"`
	Password string `description:"Basic Auth password" json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	endpoint := &EndpointConfig{}
	endpoint.SetDefaults()
	p.Endpoint = endpoint
	p.RefreshInterval = types.Duration(15 * time.Second)
	p.Prefix = "traefik"
	p.ExposedByDefault = true
	p.DefaultRule = DefaultTemplateRule
	p.tlsChan = make(chan *api.LeafCert)
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

// Provide allows the consul catalog provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	if p.ConnectAware {
		pool.GoCtx(p.registerConnectService)
		pool.GoCtx(p.watchConnectTls)
	}

	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, "consulcatalog"))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			var (
				err       error
				tlsConfig *api.LeafCert
			)

			p.client, err = createClient(p.Endpoint)
			if err != nil {
				return fmt.Errorf("error create consul client, %v", err)
			}

			ticker := time.NewTicker(time.Duration(p.RefreshInterval))

			// If we are running in connect aware mode then we need to
			// make sure that we obtain the certificates before starting
			// the service watcher, otherwise a connect enabled service
			// that gets resolved before the certificates are available
			// will cause an error condition.
			if p.ConnectAware {
				tlsConfig = <-p.tlsChan
			}

			for {
				select {
				case <-ticker.C:
					data, err := p.getConsulServicesData(routineCtx)
					if err != nil {
						logger.Errorf("error get consul catalog data, %v", err)
						return err
					}

					configuration := p.buildConfiguration(routineCtx, data, tlsConfig)
					configurationChan <- dynamic.Message{
						ProviderName:  "consulcatalog",
						Configuration: configuration,
					}
				case tlsConfig = <-p.tlsChan:
					// nothing much to do, next ticker cycle will propagate
					// the updates.
				case <-routineCtx.Done():
					ticker.Stop()
					return nil
				}
			}
		}

		notify := func(err error, time time.Duration) {
			logger.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}

		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxLog), notify)
		if err != nil {
			logger.Errorf("Cannot connect to consul catalog server %+v", err)
		}
	})

	return nil
}

func (p *Provider) getConsulServicesData(ctx context.Context) ([]itemData, error) {
	consulServiceNames, err := p.fetchServices(ctx)
	if err != nil {
		return nil, err
	}

	var data []itemData
	for name := range consulServiceNames {
		consulServices, healthServices, err := p.fetchService(ctx, name)
		if err != nil {
			return nil, err
		}

		for i, consulService := range consulServices {
			address := consulService.ServiceAddress
			if address == "" {
				address = consulService.Address
			}

			item := itemData{
				ID:      consulService.ServiceID,
				Node:    consulService.Node,
				Name:    consulService.ServiceName,
				Address: address,
				Port:    strconv.Itoa(consulService.ServicePort),
				Labels:  tagsToNeutralLabels(consulService.ServiceTags, p.Prefix),
				Tags:    consulService.ServiceTags,
				Status:  healthServices[i].Checks.AggregatedStatus(),
			}

			extraConf, err := p.getConfiguration(item)
			if err != nil {
				log.FromContext(ctx).Errorf("Skip item %s: %v", item.Name, err)
				continue
			}

			extraConf.ConnectEnabled = isConnectEnabled(consulService)
			item.ExtraConf = extraConf

			data = append(data, item)
		}
	}
	return data, nil
}

func isConnectEnabled(service *api.CatalogService) bool {
	if service.ServiceProxy == nil {
		return false
	}

	return service.ServiceProxy.DestinationServiceID != ""
}

func (p *Provider) fetchService(ctx context.Context, name string) ([]*api.CatalogService, []*api.ServiceEntry, error) {
	var tagFilter string
	if !p.ExposedByDefault {
		tagFilter = p.Prefix + ".enable=true"
	}

	opts := &api.QueryOptions{AllowStale: p.Stale, RequireConsistent: p.RequireConsistent, UseCache: p.Cache}

	consulServices, _, err := p.client.Catalog().Service(name, tagFilter, opts)
	if err != nil {
		return nil, nil, err
	}

	healthServices, _, err := p.client.Health().Service(name, tagFilter, false, opts)
	return consulServices, healthServices, err
}

func (p *Provider) fetchServices(ctx context.Context) (map[string][]string, error) {
	opts := &api.QueryOptions{AllowStale: p.Stale, RequireConsistent: p.RequireConsistent, UseCache: p.Cache}
	serviceNames, _, err := p.client.Catalog().Services(opts)
	return serviceNames, err
}

func (p *Provider) registerConnectService(ctx context.Context) {
	if !p.ConnectNative {
		return
	}

	ctxLog := log.With(ctx, log.Str(log.ProviderName, "consulcatalog"))
	logger := log.FromContext(ctxLog)

	if p.ServiceName == "" {
		p.ServiceName = "traefik"
	}

	client, err := createClient(p.Endpoint)
	if err != nil {
		logger.WithError(err).Error("failed to create consul client")
		return
	}

	serviceId := uuid.New().String()
	operation := func() error {
		regReq := &api.AgentServiceRegistration{
			ID:   serviceId,
			Kind: api.ServiceKindTypical,
			Name: p.ServiceName,
			Port: p.ServicePort,
			Connect: &api.AgentServiceConnect{
				Native: true,
			},
		}

		err = client.Agent().ServiceRegister(regReq)
		if err != nil {
			return fmt.Errorf("failed to register service in consul catalog. %s", err)
		}

		return nil
	}

	notify := func(err error, time time.Duration) {
		logger.Errorf("Failed to register traefik as Connect Native service in consul catalog. %s", err)
	}

	err = backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), context.Background()), notify)
	if err != nil {
		logger.WithError(err).Error("failed to register traefik in consul catalog as connect native service")
		return
	}

	<-ctx.Done()
	err = client.Agent().ServiceDeregister(serviceId)
	if err != nil {
		logger.WithError(err).Error("failed to deregister traefik from consul catalog")
	}
}

func (p *Provider) watchConnectTls(ctx context.Context) {
	ctxLog := log.With(ctx, log.Str(log.ProviderName, "consulcatalog"))
	logger := log.FromContext(ctxLog)

	operation := func() error {
		client, err := createClient(p.Endpoint)
		if err != nil {
			return fmt.Errorf("failed to create consul client. %s", err)
		}

		qopts := &api.QueryOptions{
			AllowStale:        p.Stale,
			RequireConsistent: p.RequireConsistent,
			UseCache:          p.Cache,
		}

		ticker := time.NewTicker(time.Duration(p.RefreshInterval))

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				resp, _, err := client.Agent().ConnectCALeaf(p.ServiceName, qopts.WithContext(ctx))
				if err != nil {
					return fmt.Errorf("failed to fetch TLS leaf certificates. %s", err)
				}

				p.tlsChan <- resp
			}
		}
	}

	notify := func(err error, time time.Duration) {
		logger.WithError(err).Errorf("failed to retrieve leaf certificates from consul. retrying in %s", time)
	}

	err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxLog), notify)
	if err != nil {
		logger.WithError(err).Errorf("Cannot read Connect TLS certificates from consul agent")
	}
}

func createClient(cfg *EndpointConfig) (*api.Client, error) {
	config := api.Config{
		Address:    cfg.Address,
		Scheme:     cfg.Scheme,
		Datacenter: cfg.DataCenter,
		WaitTime:   time.Duration(cfg.EndpointWaitTime),
		Token:      cfg.Token,
	}

	if cfg.HTTPAuth != nil {
		config.HttpAuth = &api.HttpBasicAuth{
			Username: cfg.HTTPAuth.Username,
			Password: cfg.HTTPAuth.Password,
		}
	}

	if cfg.TLS != nil {
		config.TLSConfig = api.TLSConfig{
			Address:            cfg.Address,
			CAFile:             cfg.TLS.CA,
			CertFile:           cfg.TLS.Cert,
			KeyFile:            cfg.TLS.Key,
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		}
	}

	return api.NewClient(&config)
}
