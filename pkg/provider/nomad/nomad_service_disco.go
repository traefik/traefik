package nomad

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/nomad/api"
	types2 "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/job"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/constraints"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/types"
)

const (
	// providerName is the name of this provider.
	providerName = "nomad"

	// DefaultTemplateRule The default template for the default rule.
	DefaultTemplateRule = "Host(`{{ normalize .Name }}`)"

	// defaultPrefix is the default prefix used in tag values indicating the service
	// should be consumed and exposed via traefik.
	defaultPrefix = "traefik"
)

var _ provider.Provider = (*Provider)(nil)

type item struct {
	ID         string   // service ID
	Name       string   // service name
	Namespace  string   // service namespace
	Node       string   // node ID
	Datacenter string   // region
	Address    string   // service address
	Port       int      // service port
	Tags       []string // service tags

	ExtraConf configuration // global options
}

// Provider holds configurations of the provider.
type Provider struct {
	DefaultRule      string          `description:"Default rule." json:"defaultRule,omitempty" toml:"defaultRule,omitempty" yaml:"defaultRule,omitempty"`
	Constraints      string          `description:"Constraints is an expression that Traefik matches against the Nomad service's tags to determine whether to create route(s) for that service." json:"constraints,omitempty" toml:"constraints,omitempty" yaml:"constraints,omitempty" export:"true"`
	Endpoint         *EndpointConfig `description:"Nomad endpoint settings" json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
	Prefix           string          `description:"Prefix for nomad service tags. Default 'traefik'" json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
	Stale            bool            `description:"Use stale consistency for catalog reads." json:"stale,omitempty" toml:"stale,omitempty" yaml:"stale,omitempty" export:"true"`
	Namespace        string          `description:"Sets the Nomad namespace used to discover services." json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty" export:"true"`
	ExposedByDefault bool            `description:"Expose Nomad services by default." json:"exposedByDefault,omitempty" toml:"exposedByDefault,omitempty" yaml:"exposedByDefault,omitempty" export:"true"`
	RefreshInterval  types2.Duration `description:"Interval for polling Nomad API. Default 15s" json:"refreshInterval,omitempty" toml:"refreshInterval,omitempty" yaml:"refreshInterval,omitempty" export:"true"`

	nClient           *api.Client        // client for Nomad API
	defaultRuleTpl    *template.Template // default routing rule
	watchServicesChan chan struct{}      // poll notifications on notifications
}

type EndpointConfig struct {
	Address          string           `description:"The address of the Nomad server, including scheme and port." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Region           string           `description:"Nomad region to use. If not provided, the local agent region is used." json:"region,omitempty" toml:"region,omitempty" yaml:"region,omitempty"`
	Token            string           `description:"Token is used to provide a per-request ACL token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	TLS              *types.ClientTLS `description:"Configure TLS." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	EndpointWaitTime types2.Duration  `description:"WaitTime limits how long a Watch will block. If not provided, the agent default values will be used" json:"endpointWaitTime,omitempty" toml:"endpointWaitTime,omitempty" yaml:"endpointWaitTime,omitempty" export:"true"`
}

// SetDefaults sets the default values for the Nomad Traefik Provider.
func (p *Provider) SetDefaults() {
	p.Endpoint = &EndpointConfig{
		Address: "",  // empty -> defer to NOMAD_ADDR or "http://localhost:4646"
		Region:  "",  // empty -> defer to NOMAD_REGION or "global"
		Token:   "",  // empty -> defer to NOMAD_TOKEN
		TLS:     nil, // empty -> default to http
	}
	p.Prefix = defaultPrefix
	p.ExposedByDefault = true
	p.RefreshInterval = types2.Duration(15 * time.Second)
	p.DefaultRule = DefaultTemplateRule
}

// Init the Nomad Traefik Provider.
func (p *Provider) Init() error {
	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %w", err)
	}
	p.defaultRuleTpl = defaultRuleTpl
	p.watchServicesChan = make(chan struct{}, 2)
	return nil
}

// Provide allows the Nomad Traefik Provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	var err error
	p.nClient, err = createClient(p.Namespace, p.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to create nomad API client: %w", err)
	}

	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, providerName))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			ctx, cancel := context.WithCancel(ctxLog)
			defer cancel()

			// load initial configuration
			if loadErr := p.loadConfiguration(ctx, configurationChan); loadErr != nil {
				return fmt.Errorf("failed to load initial nomad services: %w", loadErr)
			}

			go func() {
				// issue periodic refreshes in the background
				// (Nomad does not support Watch style observations)
				notifications(ctx, time.Duration(p.RefreshInterval), p.watchServicesChan)
			}()

			// enter loop where we wait for and respond to notifications
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-p.watchServicesChan:
				}
				// load services due to refresh
				if loadErr := p.loadConfiguration(ctx, configurationChan); loadErr != nil {
					return fmt.Errorf("failed to refresh nomad services: %w", loadErr)
				}
			}
		}

		failure := func(err error, d time.Duration) {
			logger.Errorf("Provider connection error %+v, retrying in %s", err, d)
		}

		if retryErr := backoff.RetryNotify(
			safe.OperationWithRecover(operation),
			backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxLog),
			failure,
		); retryErr != nil {
			logger.Errorf("Cannot connect to Nomad server %+v", retryErr)
		}
	})

	return nil
}

func (p *Provider) loadConfiguration(ctx context.Context, configurationC chan<- dynamic.Message) error {
	items, err := p.getNomadServiceData(ctx)
	if err != nil {
		return err
	}
	configurationC <- dynamic.Message{
		ProviderName:  providerName,
		Configuration: p.buildConfig(ctx, items),
	}

	return nil
}

func createClient(namespace string, endpoint *EndpointConfig) (*api.Client, error) {
	config := api.Config{
		Address:   endpoint.Address,
		Region:    endpoint.Region,
		WaitTime:  time.Duration(endpoint.EndpointWaitTime),
		Namespace: namespace,
	}

	if endpoint.TLS != nil {
		address := strings.TrimPrefix(endpoint.Address, "http")
		address = strings.TrimPrefix(address, "https")

		config.TLSConfig = &api.TLSConfig{
			CACert:        endpoint.TLS.CA,
			CAPath:        "",
			CACertPEM:     nil,
			ClientCert:    endpoint.TLS.Cert,
			ClientCertPEM: nil,
			ClientKey:     endpoint.TLS.Key,
			ClientKeyPEM:  nil,
			TLSServerName: address,
			Insecure:      endpoint.TLS.InsecureSkipVerify,
		}
	}

	return api.NewClient(&config)
}

// notifications will send on c after each interval, until ctx is Done.
func notifications(ctx context.Context, interval time.Duration, c chan<- struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case <-ctx.Done():
				return
			case c <- struct{}{}:
			default:
				// c is full, discard event
			}
		}
	}
}

// configuration contains information from the service's tags that are globals
// (not specific to the dynamic configuration).
type configuration struct {
	Enable bool // <prefix>.enable
}

// globalConfig returns a configuration with settings not specific to a particular
// dynamic configuration (i.e. "<prefix>.enable")
func (p *Provider) globalConfig(tags []string) configuration {
	c := configuration{Enable: p.ExposedByDefault}
	labels := tagsToLabels(tags, p.Prefix)
	if v, exists := labels["traefik.enable"]; exists {
		c.Enable = v == "true"
	}
	return c
}

func (p *Provider) getNomadServiceData(ctx context.Context) ([]item, error) {
	// first, get list of service stubs
	opts := &api.QueryOptions{AllowStale: p.Stale}
	opts = opts.WithContext(ctx)

	stubs, _, listErr := p.nClient.Services().List(opts)
	if listErr != nil {
		return nil, listErr
	}

	var items []item

	for _, stub := range stubs {
		for _, service := range stub.Services {
			name, tags := service.ServiceName, service.Tags
			logger := log.FromContext(log.With(ctx, log.Str("serviceName", name)))
			globalCfg := p.globalConfig(tags)
			if !globalCfg.Enable {
				logger.Debug("Filter Nomad service that is not enabled")
				continue
			}

			matches, matchErr := constraints.MatchTags(tags, p.Constraints)
			if matchErr != nil {
				logger.Errorf("Error matching constraint expressions: %v", matchErr)
				continue
			}

			if !matches {
				logger.Debugf("Filter Nomad service not matching constraints: %q", p.Constraints)
				continue
			}

			instances, fetchErr := p.fetchService(ctx, name)
			if fetchErr != nil {
				return nil, fetchErr
			}

			for _, i := range instances {
				items = append(items, item{
					ID:         i.ID,
					Name:       i.ServiceName,
					Namespace:  i.Namespace,
					Node:       i.NodeID,
					Datacenter: i.Datacenter,
					Address:    i.Address,
					Port:       i.Port,
					Tags:       i.Tags,
					ExtraConf:  p.globalConfig(i.Tags),
				})
			}
		}
	}

	return items, nil
}

func enableTag(prefix string) string {
	return fmt.Sprintf("%s.enable=true", prefix)
}

// fetchService queries Nomad API for services matching name, that also have the
// <prefix>.enable=true set in its tags.
//
// todo: Nomad currently (v1.3.0) does not support health checks, and as such does
//  not yet return health status information. When it does, refactor this section
//  to include health status.
func (p *Provider) fetchService(ctx context.Context, name string) ([]*api.ServiceRegistration, error) {
	var tagFilter string
	if !p.ExposedByDefault {
		tagFilter = fmt.Sprintf(`Tags contains %q`, enableTag(p.Prefix))
	}

	opts := &api.QueryOptions{AllowStale: p.Stale, Filter: tagFilter}
	opts = opts.WithContext(ctx)

	services, _, err := p.nClient.Services().Get(name, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %w", err)
	}
	return services, nil
}
