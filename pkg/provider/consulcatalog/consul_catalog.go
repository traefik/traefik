package consulcatalog

import (
	"context"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/job"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/constraints"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/types"
)

// DefaultTemplateRule The default template for the default rule.
const DefaultTemplateRule = "Host(`{{ normalize .Name }}`)"

var _ provider.Provider = (*Provider)(nil)

type itemData struct {
	ID         string
	Node       string
	Datacenter string
	Name       string
	Namespace  string
	Address    string
	Port       string
	Status     string
	Labels     map[string]string
	Tags       []string
	ExtraConf  configuration
}

// Provider holds configurations of the provider.
type Provider struct {
	Constraints       string          `description:"Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container." json:"constraints,omitempty" toml:"constraints,omitempty" yaml:"constraints,omitempty" export:"true"`
	Endpoint          *EndpointConfig `description:"Consul endpoint settings" json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
	Prefix            string          `description:"Prefix for consul service tags. Default 'traefik'" json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
	RefreshInterval   ptypes.Duration `description:"Interval for check Consul API. Default 15s" json:"refreshInterval,omitempty" toml:"refreshInterval,omitempty" yaml:"refreshInterval,omitempty" export:"true"`
	RequireConsistent bool            `description:"Forces the read to be fully consistent." json:"requireConsistent,omitempty" toml:"requireConsistent,omitempty" yaml:"requireConsistent,omitempty" export:"true"`
	Stale             bool            `description:"Use stale consistency for catalog reads." json:"stale,omitempty" toml:"stale,omitempty" yaml:"stale,omitempty" export:"true"`
	Cache             bool            `description:"Use local agent caching for catalog reads." json:"cache,omitempty" toml:"cache,omitempty" yaml:"cache,omitempty" export:"true"`
	ExposedByDefault  bool            `description:"Expose containers by default." json:"exposedByDefault,omitempty" toml:"exposedByDefault,omitempty" yaml:"exposedByDefault,omitempty" export:"true"`
	DefaultRule       string          `description:"Default rule." json:"defaultRule,omitempty" toml:"defaultRule,omitempty" yaml:"defaultRule,omitempty"`
	ConnectAware      bool            `description:"Enable Consul Connect support." json:"connectAware,omitempty" toml:"connectAware,omitempty" yaml:"connectAware,omitempty" export:"true"`
	ConnectByDefault  bool            `description:"Consider every service as Connect capable by default." json:"connectByDefault,omitempty" toml:"connectByDefault,omitempty" yaml:"connectByDefault,omitempty" export:"true"`
	ServiceName       string          `description:"Name of the Traefik service in Consul Catalog (needs to be registered via the orchestrator or manually)." json:"serviceName,omitempty" toml:"serviceName,omitempty" yaml:"serviceName,omitempty" export:"true"`
	Namespace         string          `description:"Sets the namespace used to discover services (Consul Enterprise only)." json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty" export:"true"`

	client         *api.Client
	defaultRuleTpl *template.Template
	certChan       chan *connectCert
}

// EndpointConfig holds configurations of the endpoint.
type EndpointConfig struct {
	Address          string                  `description:"The address of the Consul server" json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Scheme           string                  `description:"The URI scheme for the Consul server" json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty"`
	DataCenter       string                  `description:"Data center to use. If not provided, the default agent data center is used" json:"datacenter,omitempty" toml:"datacenter,omitempty" yaml:"datacenter,omitempty"`
	Token            string                  `description:"Token is used to provide a per-request ACL token which overrides the agent's default token" json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty"`
	TLS              *types.ClientTLS        `description:"Enable TLS support." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	HTTPAuth         *EndpointHTTPAuthConfig `description:"Auth info to use for http access" json:"httpAuth,omitempty" toml:"httpAuth,omitempty" yaml:"httpAuth,omitempty" export:"true"`
	EndpointWaitTime ptypes.Duration         `description:"WaitTime limits how long a Watch will block. If not provided, the agent default values will be used" json:"endpointWaitTime,omitempty" toml:"endpointWaitTime,omitempty" yaml:"endpointWaitTime,omitempty" export:"true"`
}

// EndpointHTTPAuthConfig holds configurations of the authentication.
type EndpointHTTPAuthConfig struct {
	Username string `description:"Basic Auth username" json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty"`
	Password string `description:"Basic Auth password" json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	endpoint := &EndpointConfig{}
	p.Endpoint = endpoint
	p.RefreshInterval = ptypes.Duration(15 * time.Second)
	p.Prefix = "traefik"
	p.ExposedByDefault = true
	p.DefaultRule = DefaultTemplateRule
	p.ServiceName = "traefik"
}

// Init the provider.
func (p *Provider) Init() error {
	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %w", err)
	}

	p.defaultRuleTpl = defaultRuleTpl
	p.certChan = make(chan *connectCert)
	return nil
}

// Provide allows the consul catalog provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	var err error
	p.client, err = createClient(p.Endpoint)
	if err != nil {
		return fmt.Errorf("unable to create consul client: %w", err)
	}

	if p.ConnectAware {
		leafWatcher, rootWatcher, err := p.createConnectTLSWatchers()
		if err != nil {
			return fmt.Errorf("unable to create consul watch plans: %w", err)
		}
		pool.GoCtx(func(routineCtx context.Context) {
			p.watchConnectTLS(routineCtx, leafWatcher, rootWatcher)
		})
	}

	var certInfo *connectCert

	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, "consulcatalog"))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			var err error

			// If we are running in connect aware mode then we need to
			// make sure that we obtain the certificates before starting
			// the service watcher, otherwise a connect enabled service
			// that gets resolved before the certificates are available
			// will cause an error condition.
			if p.ConnectAware && !certInfo.isReady() {
				logger.Infof("Waiting for Connect certificate before building first configuration")
				select {
				case <-routineCtx.Done():
					return nil
				case certInfo = <-p.certChan:
					if certInfo.err != nil {
						return backoff.Permanent(err)
					}
				}
			}

			// get configuration at the provider's startup.
			err = p.loadConfiguration(ctxLog, certInfo, configurationChan)
			if err != nil {
				return fmt.Errorf("failed to get consul catalog data: %w", err)
			}

			// Periodic refreshes.
			ticker := time.NewTicker(time.Duration(p.RefreshInterval))
			defer ticker.Stop()

			for {
				select {
				case <-routineCtx.Done():
					return nil
				case <-ticker.C:
				case certInfo = <-p.certChan:
					if certInfo.err != nil {
						return backoff.Permanent(err)
					}
				}
				err = p.loadConfiguration(ctxLog, certInfo, configurationChan)
				if err != nil {
					return fmt.Errorf("failed to refresh consul catalog data: %w", err)
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

func (p *Provider) loadConfiguration(ctx context.Context, certInfo *connectCert, configurationChan chan<- dynamic.Message) error {
	data, err := p.getConsulServicesData(ctx)
	if err != nil {
		return err
	}

	configurationChan <- dynamic.Message{
		ProviderName:  "consulcatalog",
		Configuration: p.buildConfiguration(ctx, data, certInfo),
	}

	return nil
}

func (p *Provider) getConsulServicesData(ctx context.Context) ([]itemData, error) {
	// The query option "Filter" is not supported by /catalog/services.
	// https://www.consul.io/api/catalog.html#list-services
	opts := &api.QueryOptions{AllowStale: p.Stale, RequireConsistent: p.RequireConsistent, UseCache: p.Cache, Namespace: p.Namespace}
	opts = opts.WithContext(ctx)

	serviceNames, _, err := p.client.Catalog().Services(opts)
	if err != nil {
		return nil, err
	}

	var data []itemData
	for name, tags := range serviceNames {
		logger := log.FromContext(log.With(ctx, log.Str("serviceName", name)))

		svcCfg, err := p.getConfiguration(tagsToNeutralLabels(tags, p.Prefix))
		if err != nil {
			logger.Errorf("Skip service: %v", err)
			continue
		}

		if !svcCfg.Enable {
			logger.Debug("Filtering disabled item")
			continue
		}

		matches, err := constraints.MatchTags(tags, p.Constraints)
		if err != nil {
			logger.Errorf("Error matching constraint expressions: %v", err)
			continue
		}

		if !matches {
			logger.Debugf("Container pruned by constraint expressions: %q", p.Constraints)
			continue
		}

		if !p.ConnectAware && svcCfg.ConsulCatalog.Connect {
			logger.Debugf("Filtering out Connect aware item, Connect support is not enabled")
			continue
		}

		consulServices, statuses, err := p.fetchService(ctx, name, svcCfg.ConsulCatalog.Connect)
		if err != nil {
			return nil, err
		}

		for _, consulService := range consulServices {
			address := consulService.ServiceAddress
			if address == "" {
				address = consulService.Address
			}

			namespace := consulService.Namespace
			if namespace == "" {
				namespace = "default"
			}

			status, exists := statuses[consulService.ID+consulService.ServiceID]
			if !exists {
				status = api.HealthAny
			}

			item := itemData{
				ID:         consulService.ServiceID,
				Node:       consulService.Node,
				Datacenter: consulService.Datacenter,
				Namespace:  namespace,
				Name:       name,
				Address:    address,
				Port:       strconv.Itoa(consulService.ServicePort),
				Labels:     tagsToNeutralLabels(consulService.ServiceTags, p.Prefix),
				Tags:       consulService.ServiceTags,
				Status:     status,
			}

			extraConf, err := p.getConfiguration(item.Labels)
			if err != nil {
				log.FromContext(ctx).Errorf("Skip item %s: %v", item.Name, err)
				continue
			}
			item.ExtraConf = extraConf

			data = append(data, item)
		}
	}

	return data, nil
}

func (p *Provider) fetchService(ctx context.Context, name string, connectEnabled bool) ([]*api.CatalogService, map[string]string, error) {
	var tagFilter string
	if !p.ExposedByDefault {
		tagFilter = p.Prefix + ".enable=true"
	}

	opts := &api.QueryOptions{AllowStale: p.Stale, RequireConsistent: p.RequireConsistent, UseCache: p.Cache, Namespace: p.Namespace}
	opts = opts.WithContext(ctx)

	catalogFunc := p.client.Catalog().Service
	healthFunc := p.client.Health().Service
	if connectEnabled {
		catalogFunc = p.client.Catalog().Connect
		healthFunc = p.client.Health().Connect
	}

	consulServices, _, err := catalogFunc(name, tagFilter, opts)
	if err != nil {
		return nil, nil, err
	}

	healthServices, _, err := healthFunc(name, tagFilter, false, opts)
	if err != nil {
		return nil, nil, err
	}

	// Index status by service and node so it can be retrieved from a CatalogService even if the health and services
	// are not in sync.
	statuses := make(map[string]string)
	for _, health := range healthServices {
		if health.Service == nil || health.Node == nil {
			continue
		}

		statuses[health.Node.ID+health.Service.ID] = health.Checks.AggregatedStatus()
	}

	return consulServices, statuses, err
}

func rootsWatchHandler(ctx context.Context, dest chan<- []string) func(watch.BlockingParamVal, interface{}) {
	return func(_ watch.BlockingParamVal, raw interface{}) {
		if raw == nil {
			log.FromContext(ctx).Errorf("Root certificate watcher called with nil")
			return
		}

		v, ok := raw.(*api.CARootList)
		if !ok || v == nil {
			log.FromContext(ctx).Errorf("Invalid result for root certificate watcher")
			return
		}

		roots := make([]string, 0, len(v.Roots))
		for _, root := range v.Roots {
			roots = append(roots, root.RootCertPEM)
		}

		dest <- roots
	}
}

type keyPair struct {
	cert string
	key  string
}

func leafWatcherHandler(ctx context.Context, dest chan<- keyPair) func(watch.BlockingParamVal, interface{}) {
	return func(_ watch.BlockingParamVal, raw interface{}) {
		if raw == nil {
			log.FromContext(ctx).Errorf("Leaf certificate watcher called with nil")
			return
		}

		v, ok := raw.(*api.LeafCert)
		if !ok || v == nil {
			log.FromContext(ctx).Errorf("Invalid result for leaf certificate watcher")
			return
		}

		dest <- keyPair{
			cert: v.CertPEM,
			key:  v.PrivateKeyPEM,
		}
	}
}

func (p *Provider) createConnectTLSWatchers() (*watch.Plan, *watch.Plan, error) {
	leafWatcher, err := watch.Parse(map[string]interface{}{
		"type":    "connect_leaf",
		"service": p.ServiceName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create leaf cert watcher plan: %w", err)
	}

	rootWatcher, err := watch.Parse(map[string]interface{}{
		"type": "connect_roots",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create root cert watcher plan: %w", err)
	}

	return leafWatcher, rootWatcher, nil
}

// watchConnectTLS watches for updates of the root certificate or the leaf
// certificate, and transmits them to the caller via p.certChan. Any error is also
// propagated up through p.certChan, in connectCert.err.
func (p *Provider) watchConnectTLS(ctx context.Context, leafWatcher *watch.Plan, rootWatcher *watch.Plan) {
	ctxLog := log.With(ctx, log.Str(log.ProviderName, "consulcatalog"))
	logger := log.FromContext(ctxLog)

	leafChan := make(chan keyPair)
	rootChan := make(chan []string)

	leafWatcher.HybridHandler = leafWatcherHandler(ctx, leafChan)
	rootWatcher.HybridHandler = rootsWatchHandler(ctx, rootChan)

	logOpts := &hclog.LoggerOptions{
		Name:       "consulcatalog",
		Level:      hclog.LevelFromString(logrus.GetLevel().String()),
		JSONFormat: true,
	}

	hclogger := hclog.New(logOpts)

	go func() {
		err := leafWatcher.RunWithClientAndHclog(p.client, hclogger)
		if err != nil {
			p.certChan <- &connectCert{err: err}
		}
	}()

	go func() {
		err := rootWatcher.RunWithClientAndHclog(p.client, hclogger)
		if err != nil {
			p.certChan <- &connectCert{err: err}
		}
	}()

	var (
		certInfo  *connectCert
		leafCerts keyPair
		rootCerts []string
	)

	for {
		select {
		case <-ctx.Done():
			leafWatcher.Stop()
			rootWatcher.Stop()
			return
		case rootCerts = <-rootChan:
		case leafCerts = <-leafChan:
		}
		newCertInfo := &connectCert{
			root: rootCerts,
			leaf: leafCerts,
		}
		if newCertInfo.isReady() && !newCertInfo.equals(certInfo) {
			logger.Debugf("Updating connect certs for service %s", p.ServiceName)
			certInfo = newCertInfo
			p.certChan <- newCertInfo
		}
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
