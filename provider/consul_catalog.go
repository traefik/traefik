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

type ConsulKV interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
	List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
}

// ConsulCatalog holds configurations of the Consul catalog provider.
type ConsulCatalog struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string
	Domain       string
	client       *api.Client
	kv           ConsulKV
	Prefix       string
}

type catalogUpdate struct {
	Service string
	Nodes   []*api.ServiceEntry
}

// listen on catalog change in the current consul datacenter
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

// listen on K/V changes under /<prefix>
func (provider *ConsulCatalog) watchKVStore(stopCh <-chan struct{}) <-chan map[string][]string {
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

			_, meta, err_kv := provider.kv.List(provider.Prefix, opts)
			if err_kv != nil {
				log.WithError(err_kv).Errorf("Failed to list services")
				return
			}

			// If LastIndex didn't change then it means `Get` returned
			// because of the WaitTime and the key didn't changed.
			if opts.WaitIndex == meta.LastIndex {
				continue
			}
			opts.WaitIndex = meta.LastIndex

			data, _, err_catalog := catalog.Services(nil)
			if err_catalog != nil {
				log.WithError(err_catalog).Errorf("Failed to list keys at " + provider.Prefix)
				return
			}
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

// node_name is not mandatory
func (provider *ConsulCatalog) getKV(defaultValue string, service_name string, node_name string, property string) string {
	value := defaultValue

	key := provider.Prefix + "/" + service_name + "/"
	if node_name != "" {
		key += node_name + "/"
	}
	key += property

	pair, _, err := provider.kv.Get(key, nil)
	if err != nil {
		log.Error(err)
	}
	if err == nil && pair != nil {
		value = string(pair.Value)
	}

	return value
}

func (provider *ConsulCatalog) buildConfig(catalog []catalogUpdate) *types.Configuration {
	var FuncMap = template.FuncMap{
		"getBackend":       provider.getBackend,
		"getFrontendValue": provider.getFrontendValue,
		"replace":          replace,
		"GetKV":            provider.getKV,
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

func (provider *ConsulCatalog) handleConsulChange(configurationChan chan<- types.ConfigMessage, index map[string][]string, ok bool) error {
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
	return nil
}

func (provider *ConsulCatalog) watch(configurationChan chan<- types.ConfigMessage) error {
	stopCh := make(chan struct{})
	serviceCatalog := provider.watchServices(stopCh)
	kvStore := provider.watchKVStore(stopCh)

	defer close(stopCh)

	for {
		var err error

		select {
		case index, ok := <-serviceCatalog:
			err = provider.handleConsulChange(configurationChan, index, ok)
		case index, ok := <-kvStore:
			err = provider.handleConsulChange(configurationChan, index, ok)
		}

		if err != nil {
			return err
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
	provider.kv = client.KV()

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
