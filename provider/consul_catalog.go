package provider

import (
	"errors"
	"strings"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
)

const (
	// DefaultWatchWaitTime is the duration to wait when polling consul
	DefaultWatchWaitTime = 15 * time.Second
	// DefaultConsulCatalogTagPrefix is a prefix for additional service/node configurations
	DefaultConsulCatalogTagPrefix = "traefik"
)

// ConsulCatalog holds configurations of the Consul catalog provider.
type ConsulCatalog struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string
	Domain       string
	client       *api.Client
	Prefix       string
}

type serviceUpdate struct {
	ServiceName string
	Attributes  []string
}

type catalogUpdate struct {
	Service *serviceUpdate
	Nodes   []*api.ServiceEntry
}

func (provider *ConsulCatalog) watchServices(stopCh <-chan struct{}) <-chan map[string][]string {
	watchCh := make(chan map[string][]string)

	catalog := provider.client.Catalog()

	safe.Go(func() {
		defer close(watchCh)

		opts := &api.QueryOptions{WaitTime: DefaultWatchWaitTime}

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			data, meta, err := catalog.Services(opts)
			if err != nil {
				log.WithError(err).Errorf("Failed to list services")
				return
			}

			// If LastIndex didn't change then it means `Get` returned
			// because of the WaitTime and the key didn't changed.
			if opts.WaitIndex == meta.LastIndex {
				continue
			}
			opts.WaitIndex = meta.LastIndex

			if data != nil {
				watchCh <- data
			}
		}
	})

	return watchCh
}

func (provider *ConsulCatalog) healthyNodes(service string) (catalogUpdate, error) {
	health := provider.client.Health()
	opts := &api.QueryOptions{}
	data, _, err := health.Service(service, "", true, opts)
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch details of " + service)
		return catalogUpdate{}, err
	}

	set := map[string]bool{}
	tags := []string{}
	for _, node := range data {
		for _, tag := range node.Service.Tags {
			if _, ok := set[tag]; ok == false {
				set[tag] = true
				tags = append(tags, tag)
			}
		}
	}

	return catalogUpdate{
		Service: &serviceUpdate{
			ServiceName: service,
			Attributes:  tags,
		},
		Nodes: data,
	}, nil
}

func (provider *ConsulCatalog) getEntryPoints(list string) []string {
	return strings.Split(list, ",")
}

func (provider *ConsulCatalog) getBackend(node *api.ServiceEntry) string {
	return strings.ToLower(node.Service.Service)
}

func (provider *ConsulCatalog) getFrontendRule(service serviceUpdate) string {
	customFrontendRule := provider.getAttribute("frontend.rule", service.Attributes, "")
	if customFrontendRule != "" {
		return customFrontendRule
	}
	return "Host:" + service.ServiceName + "." + provider.Domain
}

func (provider *ConsulCatalog) getAttribute(name string, tags []string, defaultValue string) string {
	for _, tag := range tags {
		if strings.Index(tag, DefaultConsulCatalogTagPrefix+".") == 0 {
			if kv := strings.SplitN(tag[len(DefaultConsulCatalogTagPrefix+"."):], "=", 2); len(kv) == 2 && kv[0] == name {
				return kv[1]
			}
		}
	}
	return defaultValue
}

func (provider *ConsulCatalog) buildConfig(catalog []catalogUpdate) *types.Configuration {
	var FuncMap = template.FuncMap{
		"getBackend":      provider.getBackend,
		"getFrontendRule": provider.getFrontendRule,
		"getAttribute":    provider.getAttribute,
		"getEntryPoints":  provider.getEntryPoints,
		"replace":         replace,
	}

	allNodes := []*api.ServiceEntry{}
	services := []*serviceUpdate{}
	for _, info := range catalog {
		if len(info.Nodes) > 0 {
			services = append(services, info.Service)
			allNodes = append(allNodes, info.Nodes...)
		}
	}

	templateObjects := struct {
		Services []*serviceUpdate
		Nodes    []*api.ServiceEntry
	}{
		Services: services,
		Nodes:    allNodes,
	}

	configuration, err := provider.getConfiguration("templates/consul_catalog.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.WithError(err).Error("Failed to create config")
	}

	return configuration
}

func (provider *ConsulCatalog) getNodes(index map[string][]string) ([]catalogUpdate, error) {
	visited := make(map[string]bool)

	nodes := []catalogUpdate{}
	for service := range index {
		name := strings.ToLower(service)
		if !strings.Contains(name, " ") && !visited[name] {
			visited[name] = true
			log.WithFields(log.Fields{
				"service": name,
			}).Debug("Fetching service")
			healthy, err := provider.healthyNodes(name)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, healthy)
		}
	}
	return nodes, nil
}

func (provider *ConsulCatalog) watch(configurationChan chan<- types.ConfigMessage) error {
	stopCh := make(chan struct{})
	serviceCatalog := provider.watchServices(stopCh)

	defer close(stopCh)

	for {
		select {
		case index, ok := <-serviceCatalog:
			if !ok {
				return errors.New("Consul service list nil")
			}
			log.Debug("List of services changed")
			nodes, err := provider.getNodes(index)
			if err != nil {
				return err
			}
			configuration := provider.buildConfig(nodes)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "consul_catalog",
				Configuration: configuration,
			}
		}
	}
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *ConsulCatalog) Provide(configurationChan chan<- types.ConfigMessage) error {
	config := api.DefaultConfig()
	config.Address = provider.Endpoint
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}
	provider.client = client

	safe.Go(func() {
		notify := func(err error, time time.Duration) {
			log.Errorf("Consul connection error %+v, retrying in %s", err, time)
		}
		worker := func() error {
			return provider.watch(configurationChan)
		}
		err := backoff.RetryNotify(worker, backoff.NewExponentialBackOff(), notify)
		if err != nil {
			log.Fatalf("Cannot connect to consul server %+v", err)
		}
	})

	return err
}
