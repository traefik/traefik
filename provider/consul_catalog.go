package provider

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/Sirupsen/logrus"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
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

var _ Provider = (*ConsulCatalog)(nil)

// ConsulCatalog holds configurations of the Consul catalog provider.
type ConsulCatalog struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string `description:"Consul server endpoint"`
	Domain       string `description:"Default domain used"`
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

type nodeSorter []*api.ServiceEntry

func (a nodeSorter) Len() int {
	return len(a)
}

func (a nodeSorter) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a nodeSorter) Less(i int, j int) bool {
	lentr := a[i]
	rentr := a[j]

	ls := strings.ToLower(lentr.Service.Service)
	lr := strings.ToLower(rentr.Service.Service)

	if ls != lr {
		return ls < lr
	}
	if lentr.Service.Address != rentr.Service.Address {
		return lentr.Service.Address < rentr.Service.Address
	}
	if lentr.Node.Address != rentr.Node.Address {
		return lentr.Node.Address < rentr.Node.Address
	}
	return lentr.Service.Port < rentr.Service.Port
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

	nodes := fun.Filter(func(node *api.ServiceEntry) bool {
		constraintTags := provider.getContraintTags(node.Service.Tags)
		ok, failingConstraint := provider.MatchConstraints(constraintTags)
		if ok == false && failingConstraint != nil {
			log.Debugf("Service %v pruned by '%v' constraint", service, failingConstraint.String())
		}
		return ok
	}, data).([]*api.ServiceEntry)

	//Merge tags of nodes matching constraints, in a single slice.
	tags := fun.Foldl(func(node *api.ServiceEntry, set []string) []string {
		return fun.Keys(fun.Union(
			fun.Set(set),
			fun.Set(node.Service.Tags),
		).(map[string]bool)).([]string)
	}, []string{}, nodes).([]string)

	return catalogUpdate{
		Service: &serviceUpdate{
			ServiceName: service,
			Attributes:  tags,
		},
		Nodes: nodes,
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

func (provider *ConsulCatalog) getBackendAddress(node *api.ServiceEntry) string {
	if node.Service.Address != "" {
		return node.Service.Address
	}
	return node.Node.Address
}

func (provider *ConsulCatalog) getBackendName(node *api.ServiceEntry, index int) string {
	serviceName := strings.ToLower(node.Service.Service) + "--" + node.Service.Address + "--" + strconv.Itoa(node.Service.Port)

	for _, tag := range node.Service.Tags {
		serviceName += "--" + normalize(tag)
	}

	serviceName = strings.Replace(serviceName, ".", "-", -1)
	serviceName = strings.Replace(serviceName, "=", "-", -1)

	// unique int at the end
	serviceName += "--" + strconv.Itoa(index)
	return serviceName
}

func (provider *ConsulCatalog) getAttribute(name string, tags []string, defaultValue string) string {
	for _, tag := range tags {
		if strings.Index(strings.ToLower(tag), DefaultConsulCatalogTagPrefix+".") == 0 {
			if kv := strings.SplitN(tag[len(DefaultConsulCatalogTagPrefix+"."):], "=", 2); len(kv) == 2 && strings.ToLower(kv[0]) == strings.ToLower(name) {
				return kv[1]
			}
		}
	}
	return defaultValue
}

func (provider *ConsulCatalog) getContraintTags(tags []string) []string {
	var list []string

	for _, tag := range tags {
		if strings.Index(strings.ToLower(tag), DefaultConsulCatalogTagPrefix+".tags=") == 0 {
			splitedTags := strings.Split(tag[len(DefaultConsulCatalogTagPrefix+".tags="):], ",")
			list = append(list, splitedTags...)
		}
	}

	return list
}

func (provider *ConsulCatalog) buildConfig(catalog []catalogUpdate) *types.Configuration {
	var FuncMap = template.FuncMap{
		"getBackend":           provider.getBackend,
		"getFrontendRule":      provider.getFrontendRule,
		"getBackendName":       provider.getBackendName,
		"getBackendAddress":    provider.getBackendAddress,
		"getAttribute":         provider.getAttribute,
		"getEntryPoints":       provider.getEntryPoints,
		"hasMaxconnAttributes": provider.hasMaxconnAttributes,
	}

	allNodes := []*api.ServiceEntry{}
	services := []*serviceUpdate{}
	for _, info := range catalog {
		for _, node := range info.Nodes {
			isEnabled := provider.getAttribute("enable", node.Service.Tags, "true")
			if isEnabled != "false" && len(info.Nodes) > 0 {
				services = append(services, info.Service)
				allNodes = append(allNodes, info.Nodes...)
				break
			}

		}
	}
	// Ensure a stable ordering of nodes so that identical configurations may be detected
	sort.Sort(nodeSorter(allNodes))

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

func (provider *ConsulCatalog) hasMaxconnAttributes(attributes []string) bool {
	amount := provider.getAttribute("backend.maxconn.amount", attributes, "")
	extractorfunc := provider.getAttribute("backend.maxconn.extractorfunc", attributes, "")
	if amount != "" && extractorfunc != "" {
		return true
	}
	return false
}

func (provider *ConsulCatalog) getNodes(index map[string][]string) ([]catalogUpdate, error) {
	visited := make(map[string]bool)

	nodes := []catalogUpdate{}
	for service := range index {
		name := strings.ToLower(service)
		if !strings.Contains(name, " ") && !visited[name] {
			visited[name] = true
			log.WithFields(logrus.Fields{
				"service": name,
			}).Debug("Fetching service")
			healthy, err := provider.healthyNodes(name)
			if err != nil {
				return nil, err
			}
			// healthy.Nodes can be empty if constraints do not match, without throwing error
			if healthy.Service != nil && len(healthy.Nodes) > 0 {
				nodes = append(nodes, healthy)
			}
		}
	}
	return nodes, nil
}

func (provider *ConsulCatalog) watch(configurationChan chan<- types.ConfigMessage, stop chan bool) error {
	stopCh := make(chan struct{})
	serviceCatalog := provider.watchServices(stopCh)

	defer close(stopCh)

	for {
		select {
		case <-stop:
			return nil
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
func (provider *ConsulCatalog) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	config := api.DefaultConfig()
	config.Address = provider.Endpoint
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}
	provider.client = client
	provider.Constraints = append(provider.Constraints, constraints...)

	pool.Go(func(stop chan bool) {
		notify := func(err error, time time.Duration) {
			log.Errorf("Consul connection error %+v, retrying in %s", err, time)
		}
		operation := func() error {
			return provider.watch(configurationChan, stop)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to consul server %+v", err)
		}
	})

	return err
}
