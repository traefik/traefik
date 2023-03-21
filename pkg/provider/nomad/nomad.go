package nomad

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/nomad/api"
	ptypes "github.com/traefik/paerser/types"
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

	// defaultTemplateRule is the default template for the default rule.
	defaultTemplateRule = "Host(`{{ normalize .Name }}`)"

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

// ProviderBuilder is responsible for constructing namespaced instances of the Nomad provider.
type ProviderBuilder struct {
	Configuration `yaml:",inline" export:"true"`

	// Deprecated: Use Namespaces option instead
	Namespace  string   `description:"Sets the Nomad namespace used to discover services." json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	Namespaces []string `description:"Sets the Nomad namespaces used to discover services." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty"`
}

// BuildProviders builds Nomad provider instances for the given namespaces configuration.
func (p *ProviderBuilder) BuildProviders() []*Provider {
	if p.Namespace != "" {
		log.WithoutContext().Warnf("Namespace option is deprecated, please use the Namespaces option instead.")
	}

	if len(p.Namespaces) == 0 {
		return []*Provider{{
			Configuration: p.Configuration,
			name:          providerName,
			// p.Namespace could be empty
			namespace: p.Namespace,
		}}
	}

	var providers []*Provider
	for _, namespace := range p.Namespaces {
		providers = append(providers, &Provider{
			Configuration: p.Configuration,
			name:          providerName + "-" + namespace,
			namespace:     namespace,
		})
	}

	return providers
}

// Configuration represents the Nomad provider configuration.
type Configuration struct {
	DefaultRule      string          `description:"Default rule." json:"defaultRule,omitempty" toml:"defaultRule,omitempty" yaml:"defaultRule,omitempty"`
	Constraints      string          `description:"Constraints is an expression that Traefik matches against the Nomad service's tags to determine whether to create route(s) for that service." json:"constraints,omitempty" toml:"constraints,omitempty" yaml:"constraints,omitempty" export:"true"`
	Endpoint         *EndpointConfig `description:"Nomad endpoint settings" json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
	Prefix           string          `description:"Prefix for nomad service tags." json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
	Stale            bool            `description:"Use stale consistency for catalog reads." json:"stale,omitempty" toml:"stale,omitempty" yaml:"stale,omitempty" export:"true"`
	ExposedByDefault bool            `description:"Expose Nomad services by default." json:"exposedByDefault,omitempty" toml:"exposedByDefault,omitempty" yaml:"exposedByDefault,omitempty" export:"true"`
	RefreshInterval  ptypes.Duration `description:"Interval for polling Nomad API." json:"refreshInterval,omitempty" toml:"refreshInterval,omitempty" yaml:"refreshInterval,omitempty" export:"true"`
}

// SetDefaults sets the default values for the Nomad Traefik Provider Configuration.
func (c *Configuration) SetDefaults() {
	defConfig := api.DefaultConfig()
	c.Endpoint = &EndpointConfig{
		Address: defConfig.Address,
		Region:  defConfig.Region,
		Token:   defConfig.SecretID,
	}

	if defConfig.TLSConfig != nil && (defConfig.TLSConfig.Insecure || defConfig.TLSConfig.CACert != "" || defConfig.TLSConfig.ClientCert != "" || defConfig.TLSConfig.ClientKey != "") {
		c.Endpoint.TLS = &types.ClientTLS{
			CA:                 defConfig.TLSConfig.CACert,
			Cert:               defConfig.TLSConfig.ClientCert,
			Key:                defConfig.TLSConfig.ClientKey,
			InsecureSkipVerify: defConfig.TLSConfig.Insecure,
		}
	}

	c.Prefix = defaultPrefix
	c.ExposedByDefault = true
	c.RefreshInterval = ptypes.Duration(15 * time.Second)
	c.DefaultRule = defaultTemplateRule
}

// Provider holds configuration along with the namespace it will discover services in.
type Provider struct {
	Configuration

	name           string
	namespace      string
	client         *api.Client        // client for Nomad API
	defaultRuleTpl *template.Template // default routing rule
}

type EndpointConfig struct {
	// Address is the Nomad endpoint address, if empty it defaults to NOMAD_ADDR or "http://127.0.0.1:4646".
	Address string `description:"The address of the Nomad server, including scheme and port." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	// Region is the Nomad region, if empty it defaults to NOMAD_REGION.
	Region string `description:"Nomad region to use. If not provided, the local agent region is used." json:"region,omitempty" toml:"region,omitempty" yaml:"region,omitempty"`
	// Token is the ACL token to connect with Nomad, if empty it defaults to NOMAD_TOKEN.
	Token            string           `description:"Token is used to provide a per-request ACL token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	TLS              *types.ClientTLS `description:"Configure TLS." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	EndpointWaitTime ptypes.Duration  `description:"WaitTime limits how long a Watch will block. If not provided, the agent default values will be used" json:"endpointWaitTime,omitempty" toml:"endpointWaitTime,omitempty" yaml:"endpointWaitTime,omitempty" export:"true"`
}

// SetDefaults sets the default values for the Nomad Traefik Provider.
func (p *Provider) SetDefaults() {
	p.Configuration.SetDefaults()
}

// Init the Nomad Traefik Provider.
func (p *Provider) Init() error {
	if p.namespace == api.AllNamespacesNamespace {
		return errors.New("wildcard namespace not supported")
	}

	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %w", err)
	}
	p.defaultRuleTpl = defaultRuleTpl

	// In case they didn't initialize Provider with BuildProviders
	if p.name == "" {
		p.name = providerName
	}

	return nil
}

// Provide allows the Nomad Traefik Provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	var err error
	p.client, err = createClient(p.namespace, p.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to create nomad API client: %w", err)
	}

	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, p.name))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			ctx, cancel := context.WithCancel(ctxLog)
			defer cancel()

			// load initial configuration
			if err := p.loadConfiguration(ctx, configurationChan); err != nil {
				return fmt.Errorf("failed to load initial nomad services: %w", err)
			}

			// issue periodic refreshes in the background
			// (Nomad does not support Watch style observations)
			ticker := time.NewTicker(time.Duration(p.RefreshInterval))
			defer ticker.Stop()

			// enter loop where we wait for and respond to notifications
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
				}
				// load services due to refresh
				if err := p.loadConfiguration(ctx, configurationChan); err != nil {
					return fmt.Errorf("failed to refresh nomad services: %w", err)
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
		ProviderName:  p.name,
		Configuration: p.buildConfig(ctx, items),
	}

	return nil
}

func createClient(namespace string, endpoint *EndpointConfig) (*api.Client, error) {
	config := api.Config{
		Address:   endpoint.Address,
		Namespace: namespace,
		Region:    endpoint.Region,
		SecretID:  endpoint.Token,
		WaitTime:  time.Duration(endpoint.EndpointWaitTime),
	}

	if endpoint.TLS != nil {
		config.TLSConfig = &api.TLSConfig{
			CACert:     endpoint.TLS.CA,
			ClientCert: endpoint.TLS.Cert,
			ClientKey:  endpoint.TLS.Key,
			Insecure:   endpoint.TLS.InsecureSkipVerify,
		}
	}

	return api.NewClient(&config)
}

// configuration contains information from the service's tags that are globals
// (not specific to the dynamic configuration).
type configuration struct {
	Enable bool // <prefix>.enable is the corresponding label.
	Canary bool // <prefix>.nomad.canary is the corresponding label.
}

// getExtraConf returns a configuration with settings which are not part of the dynamic configuration (e.g. "<prefix>.enable").
func (p *Provider) getExtraConf(tags []string) configuration {
	labels := tagsToLabels(tags, p.Prefix)

	enabled := p.ExposedByDefault
	if v, exists := labels["traefik.enable"]; exists {
		enabled = strings.EqualFold(v, "true")
	}

	var canary bool
	if v, exists := labels["traefik.nomad.canary"]; exists {
		canary = strings.EqualFold(v, "true")
	}

	return configuration{Enable: enabled, Canary: canary}
}

func (p *Provider) getNomadServiceData(ctx context.Context) ([]item, error) {
	// first, get list of service stubs
	opts := &api.QueryOptions{AllowStale: p.Stale}
	opts = opts.WithContext(ctx)

	stubs, _, err := p.client.Services().List(opts)
	if err != nil {
		return nil, err
	}

	var items []item

	for _, stub := range stubs {
		for _, service := range stub.Services {
			logger := log.FromContext(log.With(ctx, log.Str("serviceName", service.ServiceName)))

			extraConf := p.getExtraConf(service.Tags)
			if !extraConf.Enable {
				logger.Debug("Filter Nomad service that is not enabled")
				continue
			}

			matches, err := constraints.MatchTags(service.Tags, p.Constraints)
			if err != nil {
				logger.Errorf("Error matching constraint expressions: %v", err)
				continue
			}

			if !matches {
				logger.Debugf("Filter Nomad service not matching constraints: %q", p.Constraints)
				continue
			}

			instances, err := p.fetchService(ctx, service.ServiceName)
			if err != nil {
				return nil, err
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
					ExtraConf:  p.getExtraConf(i.Tags),
				})
			}
		}
	}

	return items, nil
}

// fetchService queries Nomad API for services matching name,
// that also have the  <prefix>.enable=true set in its tags.
func (p *Provider) fetchService(ctx context.Context, name string) ([]*api.ServiceRegistration, error) {
	var tagFilter string
	if !p.ExposedByDefault {
		tagFilter = fmt.Sprintf(`Tags contains %q`, fmt.Sprintf("%s.enable=true", p.Prefix))
	}

	// TODO: Nomad currently (v1.3.0) does not support health checks,
	//  and as such does not yet return health status information.
	//  When it does, refactor this section to include health status.
	opts := &api.QueryOptions{AllowStale: p.Stale, Filter: tagFilter}
	opts = opts.WithContext(ctx)

	services, _, err := p.client.Services().Get(name, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %w", err)
	}
	return services, nil
}
