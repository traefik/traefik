package marathon

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
)

const (
	labelPort      = "traefik.port"
	labelPortIndex = "traefik.portIndex"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configuration of the provider.
type Provider struct {
	provider.BaseProvider
	Endpoint                string              `description:"Marathon server endpoint. You can also specify multiple endpoint for Marathon"`
	Domain                  string              `description:"Default domain used"`
	ExposedByDefault        bool                `description:"Expose Marathon apps by default"`
	GroupsAsSubDomains      bool                `description:"Convert Marathon groups to subdomains"`
	DCOSToken               string              `description:"DCOSToken for DCOS environment, This will override the Authorization header"`
	MarathonLBCompatibility bool                `description:"Add compatibility with marathon-lb labels"`
	TLS                     *provider.ClientTLS `description:"Enable Docker TLS support"`
	DialerTimeout           flaeg.Duration      `description:"Set a non-default connection timeout for Marathon"`
	KeepAlive               flaeg.Duration      `description:"Set a non-default TCP Keep Alive time in seconds"`
	ForceTaskHostname       bool                `description:"Force to use the task's hostname."`
	Basic                   *Basic
	marathonClient          marathon.Marathon
}

// Basic holds basic authentication specific configurations
type Basic struct {
	HTTPBasicAuthUser string
	HTTPBasicPassword string
}

type lightMarathonClient interface {
	AllTasks(v url.Values) (*marathon.Tasks, error)
	Applications(url.Values) (*marathon.Applications, error)
}

// Provide allows the marathon provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	p.Constraints = append(p.Constraints, constraints...)
	operation := func() error {
		config := marathon.NewDefaultConfig()
		config.URL = p.Endpoint
		config.EventsTransport = marathon.EventsTransportSSE
		if p.Basic != nil {
			config.HTTPBasicAuthUser = p.Basic.HTTPBasicAuthUser
			config.HTTPBasicPassword = p.Basic.HTTPBasicPassword
		}
		if len(p.DCOSToken) > 0 {
			config.DCOSToken = p.DCOSToken
		}
		TLSConfig, err := p.TLS.CreateTLSConfig()
		if err != nil {
			return err
		}
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					KeepAlive: time.Duration(p.KeepAlive),
					Timeout:   time.Duration(p.DialerTimeout),
				}).DialContext,
				TLSClientConfig: TLSConfig,
			},
		}
		client, err := marathon.NewClient(config)
		if err != nil {
			log.Errorf("Failed to create a client for marathon, error: %s", err)
			return err
		}
		p.marathonClient = client

		if p.Watch {
			update, err := client.AddEventsListener(marathon.EventIDApplications)
			if err != nil {
				log.Errorf("Failed to register for events, %s", err)
				return err
			}
			pool.Go(func(stop chan bool) {
				defer close(update)
				for {
					select {
					case <-stop:
						return
					case event := <-update:
						log.Debug("Provider event receveived", event)
						configuration := p.loadMarathonConfig()
						if configuration != nil {
							configurationChan <- types.ConfigMessage{
								ProviderName:  "marathon",
								Configuration: configuration,
							}
						}
					}
				}
			})
		}
		configuration := p.loadMarathonConfig()
		configurationChan <- types.ConfigMessage{
			ProviderName:  "marathon",
			Configuration: configuration,
		}
		return nil
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("Provider connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		log.Errorf("Cannot connect to Provider server %+v", err)
	}
	return nil
}

func (p *Provider) loadMarathonConfig() *types.Configuration {
	var MarathonFuncMap = template.FuncMap{
		"getBackend":                  p.getBackend,
		"getBackendServer":            p.getBackendServer,
		"getPort":                     p.getPort,
		"getWeight":                   p.getWeight,
		"getDomain":                   p.getDomain,
		"getProtocol":                 p.getProtocol,
		"getPassHostHeader":           p.getPassHostHeader,
		"getPriority":                 p.getPriority,
		"getEntryPoints":              p.getEntryPoints,
		"getFrontendRule":             p.getFrontendRule,
		"getFrontendBackend":          p.getFrontendBackend,
		"hasCircuitBreakerLabels":     p.hasCircuitBreakerLabels,
		"hasLoadBalancerLabels":       p.hasLoadBalancerLabels,
		"hasMaxConnLabels":            p.hasMaxConnLabels,
		"getMaxConnExtractorFunc":     p.getMaxConnExtractorFunc,
		"getMaxConnAmount":            p.getMaxConnAmount,
		"getLoadBalancerMethod":       p.getLoadBalancerMethod,
		"getCircuitBreakerExpression": p.getCircuitBreakerExpression,
		"getSticky":                   p.getSticky,
	}

	applications, err := p.marathonClient.Applications(nil)
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	tasks, err := p.marathonClient.AllTasks(&marathon.AllTasksOpts{Status: "running"})
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	//filter tasks
	filteredTasks := fun.Filter(func(task marathon.Task) bool {
		return p.taskFilter(task, applications, p.ExposedByDefault)
	}, tasks.Tasks).([]marathon.Task)

	//filter apps
	filteredApps := fun.Filter(func(app marathon.Application) bool {
		return p.applicationFilter(app, filteredTasks)
	}, applications.Apps).([]marathon.Application)

	templateObjects := struct {
		Applications []marathon.Application
		Tasks        []marathon.Task
		Domain       string
	}{
		filteredApps,
		filteredTasks,
		p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/marathon.tmpl", MarathonFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func (p *Provider) taskFilter(task marathon.Task, applications *marathon.Applications, exposedByDefaultFlag bool) bool {
	application, err := getApplication(task, applications.Apps)
	if err != nil {
		log.Errorf("Unable to get Marathon application %s for task %s", task.AppID, task.ID)
		return false
	}
	if _, err = processPorts(application, task); err != nil {
		log.Errorf("Filtering Marathon task %s from application %s without port: %s", task.ID, application.ID, err)
		return false
	}

	// Filter illegal port label specification.
	_, hasPortIndexLabel := p.getLabel(application, labelPortIndex)
	_, hasPortLabel := p.getLabel(application, labelPort)
	if hasPortIndexLabel && hasPortLabel {
		log.Debugf("Filtering Marathon task %s from application %s specifying both traefik.portIndex and traefik.port labels", task.ID, application.ID)
		return false
	}

	// Filter by constraints.
	label, _ := p.getLabel(application, "traefik.tags")
	constraintTags := strings.Split(label, ",")
	if p.MarathonLBCompatibility {
		if label, ok := p.getLabel(application, "HAPROXY_GROUP"); ok {
			constraintTags = append(constraintTags, label)
		}
	}
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Filtering Marathon task %s from application %s pruned by '%v' constraint", task.ID, application.ID, failingConstraint.String())
		}
		return false
	}

	// Filter disabled application.
	if !isApplicationEnabled(application, exposedByDefaultFlag) {
		log.Debugf("Filtering disabled Marathon task %s from application %s", task.ID, application.ID)
		return false
	}

	// Filter task with existing, bad health check results.
	if application.HasHealthChecks() {
		if task.HasHealthCheckResults() {
			for _, healthcheck := range task.HealthCheckResults {
				if !healthcheck.Alive {
					log.Debugf("Filtering Marathon task %s from application %s with bad health check", task.ID, application.ID)
					return false
				}
			}
		}
	}

	return true
}

func (p *Provider) applicationFilter(app marathon.Application, filteredTasks []marathon.Task) bool {
	label, _ := p.getLabel(app, "traefik.tags")
	constraintTags := strings.Split(label, ",")
	if p.MarathonLBCompatibility {
		if label, ok := p.getLabel(app, "HAPROXY_GROUP"); ok {
			constraintTags = append(constraintTags, label)
		}
	}
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Application %v pruned by '%v' constraint", app.ID, failingConstraint.String())
		}
		return false
	}

	return fun.Exists(func(task marathon.Task) bool {
		return task.AppID == app.ID
	}, filteredTasks)
}

func getApplication(task marathon.Task, apps []marathon.Application) (marathon.Application, error) {
	for _, application := range apps {
		if application.ID == task.AppID {
			return application, nil
		}
	}
	return marathon.Application{}, errors.New("Application not found: " + task.AppID)
}

func isApplicationEnabled(application marathon.Application, exposedByDefault bool) bool {
	return exposedByDefault && (*application.Labels)["traefik.enable"] != "false" || (*application.Labels)["traefik.enable"] == "true"
}

func (p *Provider) getLabel(application marathon.Application, label string) (string, bool) {
	for key, value := range *application.Labels {
		if key == label {
			return value, true
		}
	}
	return "", false
}

func (p *Provider) getPort(task marathon.Task, applications []marathon.Application) string {
	application, err := getApplication(task, applications)
	if err != nil {
		log.Errorf("Unable to get Marathon application %s for task %s", application.ID, task.ID)
		return ""
	}
	port, err := processPorts(application, task)
	if err != nil {
		log.Errorf("Unable to process ports for Marathon application %s and task %s: %s", application.ID, task.ID, err)
		return ""
	}

	return strconv.Itoa(port)
}

func (p *Provider) getWeight(task marathon.Task, applications []marathon.Application) string {
	application, errApp := getApplication(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return "0"
	}
	if label, ok := p.getLabel(application, "traefik.weight"); ok {
		return label
	}
	return "0"
}

func (p *Provider) getDomain(application marathon.Application) string {
	if label, ok := p.getLabel(application, "traefik.domain"); ok {
		return label
	}
	return p.Domain
}

func (p *Provider) getProtocol(task marathon.Task, applications []marathon.Application) string {
	application, errApp := getApplication(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return "http"
	}
	if label, ok := p.getLabel(application, "traefik.protocol"); ok {
		return label
	}
	return "http"
}

func (p *Provider) getSticky(application marathon.Application) string {
	if sticky, ok := p.getLabel(application, "traefik.backend.loadbalancer.sticky"); ok {
		return sticky
	}
	return "false"
}

func (p *Provider) getPassHostHeader(application marathon.Application) string {
	if passHostHeader, ok := p.getLabel(application, "traefik.frontend.passHostHeader"); ok {
		return passHostHeader
	}
	return "true"
}

func (p *Provider) getPriority(application marathon.Application) string {
	if priority, ok := p.getLabel(application, "traefik.frontend.priority"); ok {
		return priority
	}
	return "0"
}

func (p *Provider) getEntryPoints(application marathon.Application) []string {
	if entryPoints, ok := p.getLabel(application, "traefik.frontend.entryPoints"); ok {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

// getFrontendRule returns the frontend rule for the specified application, using
// it's label. It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(application marathon.Application) string {
	if label, ok := p.getLabel(application, "traefik.frontend.rule"); ok {
		return label
	}
	if p.MarathonLBCompatibility {
		if label, ok := p.getLabel(application, "HAPROXY_0_VHOST"); ok {
			return "Host:" + label
		}
	}
	return "Host:" + p.getSubDomain(application.ID) + "." + p.Domain
}

func (p *Provider) getBackend(task marathon.Task, applications []marathon.Application) string {
	application, errApp := getApplication(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return ""
	}
	return p.getFrontendBackend(application)
}

func (p *Provider) getFrontendBackend(application marathon.Application) string {
	if label, ok := p.getLabel(application, "traefik.backend"); ok {
		return label
	}
	return provider.Replace("/", "-", application.ID)
}

func (p *Provider) getSubDomain(name string) string {
	if p.GroupsAsSubDomains {
		splitedName := strings.Split(strings.TrimPrefix(name, "/"), "/")
		provider.ReverseStringSlice(&splitedName)
		reverseName := strings.Join(splitedName, ".")
		return reverseName
	}
	return strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1)
}

func (p *Provider) hasCircuitBreakerLabels(application marathon.Application) bool {
	_, ok := p.getLabel(application, "traefik.backend.circuitbreaker.expression")
	return ok
}

func (p *Provider) hasLoadBalancerLabels(application marathon.Application) bool {
	_, errMethod := p.getLabel(application, "traefik.backend.loadbalancer.method")
	_, errSticky := p.getLabel(application, "traefik.backend.loadbalancer.sticky")
	return errMethod || errSticky
}

func (p *Provider) hasMaxConnLabels(application marathon.Application) bool {
	if _, ok := p.getLabel(application, "traefik.backend.maxconn.amount"); !ok {
		return false
	}
	_, ok := p.getLabel(application, "traefik.backend.maxconn.extractorfunc")
	return ok
}

func (p *Provider) getMaxConnAmount(application marathon.Application) int64 {
	if label, ok := p.getLabel(application, "traefik.backend.maxconn.amount"); ok {
		i, errConv := strconv.ParseInt(label, 10, 64)
		if errConv != nil {
			log.Errorf("Unable to parse traefik.backend.maxconn.amount %s", label)
			return math.MaxInt64
		}
		return i
	}
	return math.MaxInt64
}

func (p *Provider) getMaxConnExtractorFunc(application marathon.Application) string {
	if label, ok := p.getLabel(application, "traefik.backend.maxconn.extractorfunc"); ok {
		return label
	}
	return "request.host"
}

func (p *Provider) getLoadBalancerMethod(application marathon.Application) string {
	if label, ok := p.getLabel(application, "traefik.backend.loadbalancer.method"); ok {
		return label
	}
	return "wrr"
}

func (p *Provider) getCircuitBreakerExpression(application marathon.Application) string {
	if label, ok := p.getLabel(application, "traefik.backend.circuitbreaker.expression"); ok {
		return label
	}
	return "NetworkErrorRatio() > 1"
}

func processPorts(application marathon.Application, task marathon.Task) (int, error) {
	if portLabel, ok := (*application.Labels)[labelPort]; ok {
		port, err := strconv.Atoi(portLabel)
		switch {
		case err != nil:
			return 0, fmt.Errorf("failed to parse port label: %s", err)
		case port <= 0:
			return 0, fmt.Errorf("explicitly specified port %d must be larger than zero", port)
		}
		return port, nil
	}

	ports := retrieveAvailablePorts(application, task)
	if len(ports) == 0 {
		return 0, errors.New("no port found")
	}

	portIndex := 0
	portIndexLabel, ok := (*application.Labels)[labelPortIndex]
	if ok {
		var err error
		portIndex, err = parseIndex(portIndexLabel, len(ports))
		if err != nil {
			return 0, fmt.Errorf("cannot use port index to select from %d ports: %s", len(ports), err)
		}
	}
	return ports[portIndex], nil
}

func retrieveAvailablePorts(application marathon.Application, task marathon.Task) []int {
	// Using default port configuration
	if task.Ports != nil && len(task.Ports) > 0 {
		return task.Ports
	}

	// Using port definition if available
	if application.PortDefinitions != nil && len(*application.PortDefinitions) > 0 {
		var ports []int
		for _, def := range *application.PortDefinitions {
			if def.Port != nil {
				ports = append(ports, *def.Port)
			}
		}
		return ports
	}
	// If using IP-per-task using this port definition
	if application.IPAddressPerTask != nil && len(*((*application.IPAddressPerTask).Discovery).Ports) > 0 {
		var ports []int
		for _, def := range *((*application.IPAddressPerTask).Discovery).Ports {
			ports = append(ports, def.Number)
		}
		return ports
	}

	return []int{}
}

func (p *Provider) getBackendServer(task marathon.Task, applications []marathon.Application) string {
	application, err := getApplication(task, applications)
	if err != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return ""
	}

	numTaskIPAddresses := len(task.IPAddresses)
	switch {
	case application.IPAddressPerTask == nil || p.ForceTaskHostname:
		return task.Host
	case numTaskIPAddresses == 0:
		log.Errorf("Missing IP address for Marathon application %s on task %s", application.ID, task.ID)
		return ""
	case numTaskIPAddresses == 1:
		return task.IPAddresses[0].IPAddress
	default:
		ipAddressIdxStr, ok := p.getLabel(application, "traefik.ipAddressIdx")
		if !ok {
			log.Errorf("Found %d task IP addresses but missing IP address index for Marathon application %s on task %s", numTaskIPAddresses, application.ID, task.ID)
			return ""
		}

		ipAddressIdx, err := parseIndex(ipAddressIdxStr, numTaskIPAddresses)
		if err != nil {
			log.Errorf("Cannot use IP address index to select from %d task IP addresses for Marathon application %s on task %s: %s", numTaskIPAddresses, application.ID, task.ID, err)
			return ""
		}

		return task.IPAddresses[ipAddressIdx].IPAddress
	}
}

func parseIndex(index string, length int) (int, error) {
	parsed, err := strconv.Atoi(index)
	switch {
	case err != nil:
		return 0, fmt.Errorf("failed to parse index '%s': %s", index, err)
	case parsed < 0, parsed > length-1:
		return 0, fmt.Errorf("index %d must be within range (0, %d)", parsed, length-1)
	}

	return parsed, nil
}
