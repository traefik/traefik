package consul

import (
	"bytes"
	"errors"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
)

const (
	// DefaultWatchWaitTime is the duration to wait when polling consul
	DefaultWatchWaitTime = 15 * time.Second
)

var _ provider.Provider = (*CatalogProvider)(nil)

// CatalogProvider holds configurations of the Consul catalog provider.
type CatalogProvider struct {
	provider.BaseProvider `mapstructure:",squash"`
	Endpoint              string `description:"Consul server endpoint"`
	Domain                string `description:"Default domain used"`
	ExposedByDefault      bool   `description:"Expose Consul services by default"`
	Prefix                string `description:"Prefix used for Consul catalog tags"`
	FrontEndRule          string `description:"Frontend rule used for Consul services"`
	client                *api.Client
	frontEndRuleTemplate  *template.Template
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

func getChangedServiceKeys(currState map[string]Service, prevState map[string]Service) ([]string, []string) {
	currKeySet := fun.Set(fun.Keys(currState).([]string)).(map[string]bool)
	prevKeySet := fun.Set(fun.Keys(prevState).([]string)).(map[string]bool)

	addedKeys := fun.Difference(currKeySet, prevKeySet).(map[string]bool)
	removedKeys := fun.Difference(prevKeySet, currKeySet).(map[string]bool)

	return fun.Keys(addedKeys).([]string), fun.Keys(removedKeys).([]string)
}

func getChangedServiceNodeKeys(currState map[string]Service, prevState map[string]Service) ([]string, []string) {
	var addedNodeKeys []string
	var removedNodeKeys []string
	for key, value := range currState {
		if prevValue, ok := prevState[key]; ok {
			addedKeys, removedKeys := getChangedHealthyKeys(value.Nodes, prevValue.Nodes)
			addedNodeKeys = append(addedKeys)
			removedNodeKeys = append(removedKeys)
		}
	}
	return addedNodeKeys, removedNodeKeys
}

func getChangedHealthyKeys(currState []string, prevState []string) ([]string, []string) {
	currKeySet := fun.Set(currState).(map[string]bool)
	prevKeySet := fun.Set(prevState).(map[string]bool)

	addedKeys := fun.Difference(currKeySet, prevKeySet).(map[string]bool)
	removedKeys := fun.Difference(prevKeySet, currKeySet).(map[string]bool)

	return fun.Keys(addedKeys).([]string), fun.Keys(removedKeys).([]string)
}

func (p *CatalogProvider) watchHealthState(stopCh <-chan struct{}, watchCh chan<- map[string][]string) {
	health := p.client.Health()
	catalog := p.client.Catalog()

	safe.Go(func() {
		// variable to hold previous state
		var flashback []string

		options := &api.QueryOptions{WaitTime: DefaultWatchWaitTime}

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			// Listening to changes that leads to `passing` state or degrades from it.
			healthyState, meta, err := health.State("passing", options)
			if err != nil {
				log.WithError(err).Error("Failed to retrieve health checks")
				return
			}

			var current []string
			if healthyState != nil {
				for _, healthy := range healthyState {
					current = append(current, healthy.ServiceID)
				}

			}

			// If LastIndex didn't change then it means `Get` returned
			// because of the WaitTime and the key didn't changed.
			if options.WaitIndex == meta.LastIndex {
				continue
			}

			options.WaitIndex = meta.LastIndex

			// The response should be unified with watchCatalogServices
			data, _, err := catalog.Services(&api.QueryOptions{})
			if err != nil {
				log.Errorf("Failed to list services: %s", err)
				return
			}

			if data != nil {
				// A critical note is that the return of a blocking request is no guarantee of a change.
				// It is possible that there was an idempotent write that does not affect the result of the query.
				// Thus it is required to do extra check for changes...
				addedKeys, removedKeys := getChangedHealthyKeys(current, flashback)

				if len(addedKeys) > 0 {
					log.WithField("DiscoveredServices", addedKeys).Debug("Health State change detected.")
					watchCh <- data
					flashback = current
				}

				if len(removedKeys) > 0 {
					log.WithField("MissingServices", removedKeys).Debug("Health State change detected.")
					watchCh <- data
					flashback = current
				}
			}
		}
	})
}

// Service represent a Consul service.
type Service struct {
	Name  string
	Tags  []string
	Nodes []string
}

func (p *CatalogProvider) watchCatalogServices(stopCh <-chan struct{}, watchCh chan<- map[string][]string) {
	catalog := p.client.Catalog()

	safe.Go(func() {
		current := make(map[string]Service)
		// variable to hold previous state
		var flashback map[string]Service

		options := &api.QueryOptions{WaitTime: DefaultWatchWaitTime}

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			data, meta, err := catalog.Services(options)
			if err != nil {
				log.Errorf("Failed to list services: %s", err)
				return
			}

			if options.WaitIndex == meta.LastIndex {
				continue
			}

			options.WaitIndex = meta.LastIndex

			if data != nil {

				for key, value := range data {
					nodes, _, err := catalog.Service(key, "", &api.QueryOptions{})
					if err != nil {
						log.Errorf("Failed to get detail of service %s: %s", key, err)
						return
					}
					nodesID := getServiceIds(nodes)
					if service, ok := current[key]; ok {
						service.Tags = value
						service.Nodes = nodesID
					} else {
						service := Service{
							Name:  key,
							Tags:  value,
							Nodes: nodesID,
						}
						current[key] = service
					}
				}
				// A critical note is that the return of a blocking request is no guarantee of a change.
				// It is possible that there was an idempotent write that does not affect the result of the query.
				// Thus it is required to do extra check for changes...
				addedServiceKeys, removedServiceKeys := getChangedServiceKeys(current, flashback)

				addedServiceNodeKeys, removedServiceNodeKeys := getChangedServiceNodeKeys(current, flashback)

				if len(removedServiceKeys) > 0 || len(removedServiceNodeKeys) > 0 || len(addedServiceKeys) > 0 || len(addedServiceNodeKeys) > 0 {
					log.WithField("MissingServices", removedServiceKeys).WithField("DiscoveredServices", addedServiceKeys).Debug("Catalog Services change detected.")
					watchCh <- data
					flashback = make(map[string]Service, len(current))
					for key, value := range current {
						flashback[key] = value
					}
				}
			}
		}
	})
}
func getServiceIds(services []*api.CatalogService) []string {
	var serviceIds []string
	for _, service := range services {
		serviceIds = append(serviceIds, service.ServiceID)
	}
	return serviceIds
}

func (p *CatalogProvider) healthyNodes(service string) (catalogUpdate, error) {
	health := p.client.Health()
	opts := &api.QueryOptions{}
	data, _, err := health.Service(service, "", true, opts)
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch details of %s", service)
		return catalogUpdate{}, err
	}

	nodes := fun.Filter(func(node *api.ServiceEntry) bool {
		return p.nodeFilter(service, node)
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

func (p *CatalogProvider) nodeFilter(service string, node *api.ServiceEntry) bool {
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

func (p *CatalogProvider) isServiceEnabled(node *api.ServiceEntry) bool {
	enable, err := strconv.ParseBool(p.getAttribute("enable", node.Service.Tags, strconv.FormatBool(p.ExposedByDefault)))
	if err != nil {
		log.Debugf("Invalid value for enable, set to %b", p.ExposedByDefault)
		return p.ExposedByDefault
	}
	return enable
}

func (p *CatalogProvider) getPrefixedName(name string) string {
	if len(p.Prefix) > 0 {
		return p.Prefix + "." + name
	}
	return name
}

func (p *CatalogProvider) getEntryPoints(list string) []string {
	return strings.Split(list, ",")
}

func (p *CatalogProvider) getBackend(node *api.ServiceEntry) string {
	return strings.ToLower(node.Service.Service)
}

func (p *CatalogProvider) getFrontendRule(service serviceUpdate) string {
	customFrontendRule := p.getAttribute("frontend.rule", service.Attributes, "")
	if customFrontendRule == "" {
		customFrontendRule = p.FrontEndRule
	}

	t := p.frontEndRuleTemplate
	t, err := t.Parse(customFrontendRule)
	if err != nil {
		log.Errorf("failed to parse Consul Catalog custom frontend rule: %s", err)
		return ""
	}

	templateObjects := struct {
		ServiceName string
		Domain      string
		Attributes  []string
	}{
		ServiceName: service.ServiceName,
		Domain:      p.Domain,
		Attributes:  service.Attributes,
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, templateObjects)
	if err != nil {
		log.Errorf("failed to execute Consul Catalog custom frontend rule template: %s", err)
		return ""
	}

	return buffer.String()
}

func (p *CatalogProvider) getBackendAddress(node *api.ServiceEntry) string {
	if node.Service.Address != "" {
		return node.Service.Address
	}
	return node.Node.Address
}

func (p *CatalogProvider) getBackendName(node *api.ServiceEntry, index int) string {
	serviceName := strings.ToLower(node.Service.Service) + "--" + node.Service.Address + "--" + strconv.Itoa(node.Service.Port)

	for _, tag := range node.Service.Tags {
		serviceName += "--" + provider.Normalize(tag)
	}

	serviceName = strings.Replace(serviceName, ".", "-", -1)
	serviceName = strings.Replace(serviceName, "=", "-", -1)

	// unique int at the end
	serviceName += "--" + strconv.Itoa(index)
	return serviceName
}

func (p *CatalogProvider) getAttribute(name string, tags []string, defaultValue string) string {
	return p.getTag(p.getPrefixedName(name), tags, defaultValue)
}

func (p *CatalogProvider) getBasicAuth(tags []string) []string {
	list := p.getAttribute("frontend.auth.basic", tags, "")
	if list != "" {
		return strings.Split(list, ",")
	}
	return []string{}
}

func (p *CatalogProvider) hasTag(name string, tags []string) bool {
	// Very-very unlikely that a Consul tag would ever start with '=!='
	tag := p.getTag(name, tags, "=!=")
	return tag != "=!="
}

func (p *CatalogProvider) getTag(name string, tags []string, defaultValue string) string {
	for _, tag := range tags {
		// Given the nature of Consul tags, which could be either singular markers, or key=value pairs, we check if the consul tag starts with 'name'
		if strings.Index(strings.ToLower(tag), strings.ToLower(name)) == 0 {
			// In case, where a tag might be a key=value, try to split it by the first '='
			// - If the first element (which would always be there, even if the tag is a singular marker without '=' in it
			if kv := strings.SplitN(tag, "=", 2); strings.ToLower(kv[0]) == strings.ToLower(name) {
				// If the returned result is a key=value pair, return the 'value' component
				if len(kv) == 2 {
					return kv[1]
				}
				// If the returned result is a singular marker, return the 'key' component
				return kv[0]
			}
		}
	}
	return defaultValue
}

func (p *CatalogProvider) getConstraintTags(tags []string) []string {
	var list []string

	for _, tag := range tags {
		// We look for a Consul tag named 'traefik.tags' (unless different 'prefix' is configured)
		if strings.Index(strings.ToLower(tag), p.getPrefixedName("tags=")) == 0 {
			// If 'traefik.tags=' tag is found, take the tag value and split by ',' adding the result to the list to be returned
			splitedTags := strings.Split(tag[len(p.getPrefixedName("tags=")):], ",")
			list = append(list, splitedTags...)
		}
	}

	return list
}

func (p *CatalogProvider) buildConfig(catalog []catalogUpdate) *types.Configuration {
	var FuncMap = template.FuncMap{
		"getBackend":           p.getBackend,
		"getFrontendRule":      p.getFrontendRule,
		"getBackendName":       p.getBackendName,
		"getBackendAddress":    p.getBackendAddress,
		"getAttribute":         p.getAttribute,
		"getBasicAuth":         p.getBasicAuth,
		"getTag":               p.getTag,
		"hasTag":               p.hasTag,
		"getEntryPoints":       p.getEntryPoints,
		"hasMaxconnAttributes": p.hasMaxconnAttributes,
	}

	allNodes := []*api.ServiceEntry{}
	services := []*serviceUpdate{}
	for _, info := range catalog {
		if len(info.Nodes) > 0 {
			services = append(services, info.Service)
			allNodes = append(allNodes, info.Nodes...)
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

	configuration, err := p.GetConfiguration("templates/consul_catalog.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.WithError(err).Error("Failed to create config")
	}

	return configuration
}

func (p *CatalogProvider) hasMaxconnAttributes(attributes []string) bool {
	amount := p.getAttribute("backend.maxconn.amount", attributes, "")
	extractorfunc := p.getAttribute("backend.maxconn.extractorfunc", attributes, "")
	if amount != "" && extractorfunc != "" {
		return true
	}
	return false
}

func (p *CatalogProvider) getNodes(index map[string][]string) ([]catalogUpdate, error) {
	visited := make(map[string]bool)

	nodes := []catalogUpdate{}
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

func (p *CatalogProvider) watch(configurationChan chan<- types.ConfigMessage, stop chan bool) error {
	stopCh := make(chan struct{})
	watchCh := make(chan map[string][]string)

	p.watchHealthState(stopCh, watchCh)
	p.watchCatalogServices(stopCh, watchCh)

	defer close(stopCh)
	defer close(watchCh)

	for {
		select {
		case <-stop:
			return nil
		case index, ok := <-watchCh:
			if !ok {
				return errors.New("Consul service list nil")
			}
			log.Debug("List of services changed")
			nodes, err := p.getNodes(index)
			if err != nil {
				return err
			}
			configuration := p.buildConfig(nodes)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "consul_catalog",
				Configuration: configuration,
			}
		}
	}
}

func (p *CatalogProvider) setupFrontEndTemplate() {
	var FuncMap = template.FuncMap{
		"getAttribute": p.getAttribute,
		"getTag":       p.getTag,
		"hasTag":       p.hasTag,
	}
	t := template.New("consul catalog frontend rule").Funcs(FuncMap)
	p.frontEndRuleTemplate = t
}

// Provide allows the consul catalog provider to provide configurations to traefik
// using the given configuration channel.
func (p *CatalogProvider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	config := api.DefaultConfig()
	config.Address = p.Endpoint
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}
	p.client = client
	p.Constraints = append(p.Constraints, constraints...)
	p.setupFrontEndTemplate()

	pool.Go(func(stop chan bool) {
		notify := func(err error, time time.Duration) {
			log.Errorf("Consul connection error %+v, retrying in %s", err, time)
		}
		operation := func() error {
			return p.watch(configurationChan, stop)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to consul server %+v", err)
		}
	})

	return err
}
