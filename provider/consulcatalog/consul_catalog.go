package consulcatalog

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
)

const (
	// DefaultWatchWaitTime is the duration to wait when polling consul
	DefaultWatchWaitTime = 15 * time.Second
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the Consul catalog provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`
	Endpoint              string           `description:"Consul server endpoint"`
	Domain                string           `description:"Default domain used"`
	Stale                 bool             `description:"Use stale consistency for catalog reads" export:"true"`
	ExposedByDefault      bool             `description:"Expose Consul services by default" export:"true"`
	Prefix                string           `description:"Prefix used for Consul catalog tags" export:"true"`
	FrontEndRule          string           `description:"Frontend rule used for Consul services" export:"true"`
	TLS                   *types.ClientTLS `description:"Enable TLS support" export:"true"`
	client                *api.Client
	frontEndRuleTemplate  *template.Template
}

// Service represent a Consul service.
type Service struct {
	Name      string
	Tags      []string
	Nodes     []string
	Addresses []string
	Ports     []int
}

type serviceUpdate struct {
	ServiceName       string
	ParentServiceName string
	Attributes        []string
	TraefikLabels     map[string]string
}

type frontendSegment struct {
	Name   string
	Labels map[string]string
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
	lEntry := a[i]
	rEntry := a[j]

	ls := strings.ToLower(lEntry.Service.Service)
	lr := strings.ToLower(rEntry.Service.Service)

	if ls != lr {
		return ls < lr
	}
	if lEntry.Service.Address != rEntry.Service.Address {
		return lEntry.Service.Address < rEntry.Service.Address
	}
	if lEntry.Node.Address != rEntry.Node.Address {
		return lEntry.Node.Address < rEntry.Node.Address
	}
	return lEntry.Service.Port < rEntry.Service.Port
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	err := p.BaseProvider.Init(constraints)
	if err != nil {
		return err
	}

	client, err := p.createClient()
	if err != nil {
		return err
	}

	p.client = client
	p.setupFrontEndRuleTemplate()

	return nil
}

// Provide allows the consul catalog provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	pool.Go(func(stop chan bool) {
		notify := func(err error, time time.Duration) {
			log.Errorf("Consul connection error %+v, retrying in %s", err, time)
		}
		operation := func() error {
			return p.watch(configurationChan, stop)
		}
		errRetry := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if errRetry != nil {
			log.Errorf("Cannot connect to consul server %+v", errRetry)
		}
	})
	return nil
}

func (p *Provider) createClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = p.Endpoint
	if p.TLS != nil {
		tlsConfig, err := p.TLS.CreateTLSConfig()
		if err != nil {
			return nil, err
		}

		config.Scheme = "https"
		config.Transport.TLSClientConfig = tlsConfig
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (p *Provider) watch(configurationChan chan<- types.ConfigMessage, stop chan bool) error {
	stopCh := make(chan struct{})
	watchCh := make(chan map[string][]string)
	errorCh := make(chan error)

	var errorOnce sync.Once
	notifyError := func(err error) {
		errorOnce.Do(func() {
			errorCh <- err
		})
	}

	p.watchHealthState(stopCh, watchCh, notifyError)
	p.watchCatalogServices(stopCh, watchCh, notifyError)

	defer close(stopCh)
	defer close(watchCh)

	safe.Go(func() {
		for index := range watchCh {
			log.Debug("List of services changed")
			nodes, err := p.getNodes(index)
			if err != nil {
				notifyError(err)
			}
			configuration := p.buildConfiguration(nodes)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "consul_catalog",
				Configuration: configuration,
			}
		}
	})

	for {
		select {
		case <-stop:
			return nil
		case err := <-errorCh:
			return err
		}
	}
}

func (p *Provider) watchCatalogServices(stopCh <-chan struct{}, watchCh chan<- map[string][]string, notifyError func(error)) {
	catalog := p.client.Catalog()

	safe.Go(func() {
		// variable to hold previous state
		var flashback map[string]Service

		options := &api.QueryOptions{WaitTime: DefaultWatchWaitTime, AllowStale: p.Stale}

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			data, meta, err := catalog.Services(options)
			if err != nil {
				log.Errorf("Failed to list services: %v", err)
				notifyError(err)
				return
			}

			if options.WaitIndex == meta.LastIndex {
				continue
			}

			options.WaitIndex = meta.LastIndex

			if data != nil {
				current := make(map[string]Service)
				for key, value := range data {
					nodes, _, err := catalog.Service(key, "", &api.QueryOptions{AllowStale: p.Stale})
					if err != nil {
						log.Errorf("Failed to get detail of service %s: %v", key, err)
						notifyError(err)
						return
					}

					nodesID := getServiceIds(nodes)
					ports := getServicePorts(nodes)
					addresses := getServiceAddresses(nodes)

					if service, ok := current[key]; ok {
						service.Tags = value
						service.Nodes = nodesID
						service.Ports = ports
					} else {
						service := Service{
							Name:      key,
							Tags:      value,
							Nodes:     nodesID,
							Addresses: addresses,
							Ports:     ports,
						}
						current[key] = service
					}
				}

				// A critical note is that the return of a blocking request is no guarantee of a change.
				// It is possible that there was an idempotent write that does not affect the result of the query.
				// Thus it is required to do extra check for changes...
				if hasChanged(current, flashback) {
					watchCh <- data
					flashback = current
				}
			}
		}
	})
}

func (p *Provider) watchHealthState(stopCh <-chan struct{}, watchCh chan<- map[string][]string, notifyError func(error)) {
	health := p.client.Health()
	catalog := p.client.Catalog()

	safe.Go(func() {
		// variable to hold previous state
		var flashback map[string][]string
		var flashbackMaintenance []string

		options := &api.QueryOptions{WaitTime: DefaultWatchWaitTime, AllowStale: p.Stale}

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			// Listening to changes that leads to `passing` state or degrades from it.
			healthyState, meta, err := health.State("any", options)
			if err != nil {
				log.WithError(err).Error("Failed to retrieve health checks")
				notifyError(err)
				return
			}

			var current = make(map[string][]string)
			var currentFailing = make(map[string]*api.HealthCheck)
			var maintenance []string
			if healthyState != nil {
				for _, healthy := range healthyState {
					key := fmt.Sprintf("%s-%s", healthy.Node, healthy.ServiceID)
					_, failing := currentFailing[key]
					if healthy.Status == "passing" && !failing {
						current[key] = append(current[key], healthy.Node)
					} else if strings.HasPrefix(healthy.CheckID, "_service_maintenance") || strings.HasPrefix(healthy.CheckID, "_node_maintenance") {
						maintenance = append(maintenance, healthy.CheckID)
					} else {
						currentFailing[key] = healthy
						if _, ok := current[key]; ok {
							delete(current, key)
						}
					}
				}
			}

			// If LastIndex didn't change then it means `Get` returned
			// because of the WaitTime and the key didn't changed.
			if options.WaitIndex == meta.LastIndex {
				continue
			}

			options.WaitIndex = meta.LastIndex

			// The response should be unified with watchCatalogServices
			data, _, err := catalog.Services(&api.QueryOptions{AllowStale: p.Stale})
			if err != nil {
				log.Errorf("Failed to list services: %v", err)
				notifyError(err)
				return
			}

			if data != nil {
				// A critical note is that the return of a blocking request is no guarantee of a change.
				// It is possible that there was an idempotent write that does not affect the result of the query.
				// Thus it is required to do extra check for changes...
				addedKeys, removedKeys, changedKeys := getChangedHealth(current, flashback)

				if len(addedKeys) > 0 || len(removedKeys) > 0 || len(changedKeys) > 0 {
					log.WithField("DiscoveredServices", addedKeys).
						WithField("MissingServices", removedKeys).
						WithField("ChangedServices", changedKeys).
						Debug("Health State change detected.")

					watchCh <- data
					flashback = current
					flashbackMaintenance = maintenance
				} else {
					addedKeysMaintenance, removedMaintenance := getChangedStringKeys(maintenance, flashbackMaintenance)

					if len(addedKeysMaintenance) > 0 || len(removedMaintenance) > 0 {
						log.WithField("MaintenanceMode", maintenance).Debug("Maintenance change detected.")
						watchCh <- data
						flashback = current
						flashbackMaintenance = maintenance
					}
				}
			}
		}
	})
}

func (p *Provider) getNodes(index map[string][]string) ([]catalogUpdate, error) {
	visited := make(map[string]bool)

	var nodes []catalogUpdate
	for service := range index {
		name := strings.ToLower(service)
		if !strings.Contains(name, " ") && !visited[name] {
			visited[name] = true
			log.WithField("service", name).Debug("Fetching service")
			healthy, err := p.healthyNodes(name)
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

func hasChanged(current map[string]Service, previous map[string]Service) bool {
	if len(current) != len(previous) {
		return true
	}
	addedServiceKeys, removedServiceKeys := getChangedServiceKeys(current, previous)
	return len(removedServiceKeys) > 0 || len(addedServiceKeys) > 0 || hasServiceChanged(current, previous)
}

func getChangedServiceKeys(current map[string]Service, previous map[string]Service) ([]string, []string) {
	currKeySet := fun.Set(fun.Keys(current).([]string)).(map[string]bool)
	prevKeySet := fun.Set(fun.Keys(previous).([]string)).(map[string]bool)

	addedKeys := fun.Difference(currKeySet, prevKeySet).(map[string]bool)
	removedKeys := fun.Difference(prevKeySet, currKeySet).(map[string]bool)

	return fun.Keys(addedKeys).([]string), fun.Keys(removedKeys).([]string)
}

func hasServiceChanged(current map[string]Service, previous map[string]Service) bool {
	for key, value := range current {
		if prevValue, ok := previous[key]; ok {
			addedNodesKeys, removedNodesKeys := getChangedStringKeys(value.Nodes, prevValue.Nodes)
			if len(addedNodesKeys) > 0 || len(removedNodesKeys) > 0 {
				return true
			}
			addedTagsKeys, removedTagsKeys := getChangedStringKeys(value.Tags, prevValue.Tags)
			if len(addedTagsKeys) > 0 || len(removedTagsKeys) > 0 {
				return true
			}
			addedAddressesKeys, removedAddressesKeys := getChangedStringKeys(value.Addresses, prevValue.Addresses)
			if len(addedAddressesKeys) > 0 || len(removedAddressesKeys) > 0 {
				return true
			}
			addedPortsKeys, removedPortsKeys := getChangedIntKeys(value.Ports, prevValue.Ports)
			if len(addedPortsKeys) > 0 || len(removedPortsKeys) > 0 {
				return true
			}
		}
	}
	return false
}

func getChangedStringKeys(currState []string, prevState []string) ([]string, []string) {
	currKeySet := fun.Set(currState).(map[string]bool)
	prevKeySet := fun.Set(prevState).(map[string]bool)

	addedKeys := fun.Difference(currKeySet, prevKeySet).(map[string]bool)
	removedKeys := fun.Difference(prevKeySet, currKeySet).(map[string]bool)

	return fun.Keys(addedKeys).([]string), fun.Keys(removedKeys).([]string)
}

func getChangedHealth(current map[string][]string, previous map[string][]string) ([]string, []string, []string) {
	currKeySet := fun.Set(fun.Keys(current).([]string)).(map[string]bool)
	prevKeySet := fun.Set(fun.Keys(previous).([]string)).(map[string]bool)

	addedKeys := fun.Difference(currKeySet, prevKeySet).(map[string]bool)
	removedKeys := fun.Difference(prevKeySet, currKeySet).(map[string]bool)

	var changedKeys []string

	for key, value := range current {
		if prevValue, ok := previous[key]; ok {
			addedNodesKeys, removedNodesKeys := getChangedStringKeys(value, prevValue)
			if len(addedNodesKeys) > 0 || len(removedNodesKeys) > 0 {
				changedKeys = append(changedKeys, key)
			}
		}
	}

	return fun.Keys(addedKeys).([]string), fun.Keys(removedKeys).([]string), changedKeys
}

func getChangedIntKeys(currState []int, prevState []int) ([]int, []int) {
	currKeySet := fun.Set(currState).(map[int]bool)
	prevKeySet := fun.Set(prevState).(map[int]bool)

	addedKeys := fun.Difference(currKeySet, prevKeySet).(map[int]bool)
	removedKeys := fun.Difference(prevKeySet, currKeySet).(map[int]bool)

	return fun.Keys(addedKeys).([]int), fun.Keys(removedKeys).([]int)
}

func getServiceIds(services []*api.CatalogService) []string {
	var serviceIds []string
	for _, service := range services {
		serviceIds = append(serviceIds, service.ID)
	}
	return serviceIds
}

func getServicePorts(services []*api.CatalogService) []int {
	var servicePorts []int
	for _, service := range services {
		servicePorts = append(servicePorts, service.ServicePort)
	}
	return servicePorts
}

func getServiceAddresses(services []*api.CatalogService) []string {
	var serviceAddresses []string
	for _, service := range services {
		serviceAddresses = append(serviceAddresses, service.ServiceAddress)
	}
	return serviceAddresses
}

func (p *Provider) healthyNodes(service string) (catalogUpdate, error) {
	health := p.client.Health()
	data, _, err := health.Service(service, "", true, &api.QueryOptions{AllowStale: p.Stale})
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch details of %s", service)
		return catalogUpdate{}, err
	}

	nodes := fun.Filter(func(node *api.ServiceEntry) bool {
		return p.nodeFilter(service, node)
	}, data).([]*api.ServiceEntry)

	// Merge tags of nodes matching constraints, in a single slice.
	tags := fun.Foldl(func(node *api.ServiceEntry, set []string) []string {
		return fun.Keys(fun.Union(
			fun.Set(set),
			fun.Set(node.Service.Tags),
		).(map[string]bool)).([]string)
	}, []string{}, nodes).([]string)

	labels := tagsToNeutralLabels(tags, p.Prefix)

	return catalogUpdate{
		Service: &serviceUpdate{
			ServiceName:   service,
			Attributes:    tags,
			TraefikLabels: labels,
		},
		Nodes: nodes,
	}, nil
}

func (p *Provider) nodeFilter(service string, node *api.ServiceEntry) bool {
	// Filter disabled application.
	if !p.isServiceEnabled(node) {
		log.Debugf("Filtering disabled Consul service %s", service)
		return false
	}

	// Filter by constraints.
	constraintTags := p.getConstraintTags(node.Service.Tags)
	ok, failingConstraint := p.MatchConstraints(constraintTags)
	if !ok && failingConstraint != nil {
		log.Debugf("Service %v pruned by '%v' constraint", service, failingConstraint.String())
		return false
	}
	return true
}

func (p *Provider) isServiceEnabled(node *api.ServiceEntry) bool {
	rawValue := getTag(p.getPrefixedName(label.SuffixEnable), node.Service.Tags, "")

	if len(rawValue) == 0 {
		return p.ExposedByDefault
	}

	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		log.Errorf("Invalid value for %s: %s", label.SuffixEnable, rawValue)
		return p.ExposedByDefault
	}
	return value
}

func (p *Provider) getConstraintTags(tags []string) []string {
	var values []string

	prefix := p.getPrefixedName("tags=")
	for _, tag := range tags {
		// We look for a Consul tag named 'traefik.tags' (unless different 'prefix' is configured)
		if strings.HasPrefix(strings.ToLower(tag), prefix) {
			// If 'traefik.tags=' tag is found, take the tag value and split by ',' adding the result to the list to be returned
			splitedTags := label.SplitAndTrimString(tag[len(prefix):], ",")
			values = append(values, splitedTags...)
		}
	}

	return values
}

func (p *Provider) generateFrontends(service *serviceUpdate) []*serviceUpdate {
	frontends := make([]*serviceUpdate, 0)
	// to support <prefix>.frontend.xxx
	frontends = append(frontends, &serviceUpdate{
		ServiceName:       service.ServiceName,
		ParentServiceName: service.ServiceName,
		Attributes:        service.Attributes,
		TraefikLabels:     service.TraefikLabels,
	})

	// loop over children of <prefix>.frontends.*
	for _, frontend := range getSegments(label.Prefix+"frontends", label.Prefix, service.TraefikLabels) {
		frontends = append(frontends, &serviceUpdate{
			ServiceName:       service.ServiceName + "-" + frontend.Name,
			ParentServiceName: service.ServiceName,
			Attributes:        service.Attributes,
			TraefikLabels:     frontend.Labels,
		})
	}

	return frontends
}

func getSegments(path string, prefix string, tree map[string]string) []*frontendSegment {
	segments := make([]*frontendSegment, 0)
	// find segment names
	segmentNames := make(map[string]bool)
	for key := range tree {
		if strings.HasPrefix(key, path+".") {
			segmentNames[strings.SplitN(strings.TrimPrefix(key, path+"."), ".", 2)[0]] = true
		}
	}

	// get labels for each segment found
	for segment := range segmentNames {
		labels := make(map[string]string)
		for key, value := range tree {
			if strings.HasPrefix(key, path+"."+segment) {
				labels[prefix+"frontend"+strings.TrimPrefix(key, path+"."+segment)] = value
			}
		}
		segments = append(segments, &frontendSegment{
			Name:   segment,
			Labels: labels,
		})
	}

	return segments
}
