package provider

import (
	"errors"
	"strings"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/emilevauge/traefik/types"
	"github.com/hashicorp/consul/api"
)

const (
	// DefaultWatchWaitTime is the duration to wait when polling consul
	DefaultWatchWaitTime = 15 * time.Second
)

// ConsulCatalog holds configurations of the Consul catalog provider.
type ConsulCatalog struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string
	Domain       string
	client       *api.Client
}

type catalogUpdate struct {
	Service string
	Nodes   []*api.ServiceEntry
}

func (provider *ConsulCatalog) watchServices(stopCh <-chan struct{}) <-chan map[string][]string {
	watchCh := make(chan map[string][]string)

	catalog := provider.client.Catalog()

	go func() {
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
	}()

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

	return catalogUpdate{
		Service: service,
		Nodes:   data,
	}, nil
}

func (provider *ConsulCatalog) getBackend(node *api.ServiceEntry) string {
	return strings.ToLower(node.Service.Service)
}

func (provider *ConsulCatalog) getFrontendValue(service string) string {
	return service + "." + provider.Domain
}

func (provider *ConsulCatalog) buildConfig(catalog []catalogUpdate) *types.Configuration {
	var FuncMap = template.FuncMap{
		"getBackend":       provider.getBackend,
		"getFrontendValue": provider.getFrontendValue,
		"replace":          replace,
	}

	allNodes := []*api.ServiceEntry{}
	serviceNames := []string{}
	for _, info := range catalog {
		if len(info.Nodes) > 0 {
			serviceNames = append(serviceNames, info.Service)
			allNodes = append(allNodes, info.Nodes...)
		}
	}

	templateObjects := struct {
		Services []string
		Nodes    []*api.ServiceEntry
	}{
		Services: serviceNames,
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

	go func() {
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
	}()

	return err
}
