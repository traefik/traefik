package marathon

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/Sirupsen/logrus"
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
	traceMaxScanTokenSize = 1024 * 1024
	marathonEventIDs      = marathon.EventIDApplications |
		marathon.EventIDAddHealthCheck |
		marathon.EventIDDeploymentSuccess |
		marathon.EventIDDeploymentFailed |
		marathon.EventIDDeploymentInfo |
		marathon.EventIDDeploymentStepSuccess |
		marathon.EventIDDeploymentStepFailed
)

// TaskState denotes the Mesos state a task can have.
type TaskState string

const (
	taskStateRunning TaskState = "TASK_RUNNING"
	taskStateStaging TaskState = "TASK_STAGING"
)

var _ provider.Provider = (*Provider)(nil)

// Regexp used to extract the name of the service and the name of the property for this service
// All properties are under the format traefik.<servicename>.frontend.*= except the port/portIndex/weight/protocol/backend directly after traefik.<servicename>.
var servicesPropertiesRegexp = regexp.MustCompile(`^traefik\.(?P<service_name>.+?)\.(?P<property_name>port|portIndex|weight|protocol|backend|frontend\.(.*))$`)

// Provider holds configuration of the provider.
type Provider struct {
	provider.BaseProvider
	Endpoint                string           `description:"Marathon server endpoint. You can also specify multiple endpoint for Marathon"`
	Domain                  string           `description:"Default domain used"`
	ExposedByDefault        bool             `description:"Expose Marathon apps by default"`
	GroupsAsSubDomains      bool             `description:"Convert Marathon groups to subdomains"`
	DCOSToken               string           `description:"DCOSToken for DCOS environment, This will override the Authorization header"`
	MarathonLBCompatibility bool             `description:"Add compatibility with marathon-lb labels"`
	TLS                     *types.ClientTLS `description:"Enable Docker TLS support"`
	DialerTimeout           flaeg.Duration   `description:"Set a non-default connection timeout for Marathon"`
	KeepAlive               flaeg.Duration   `description:"Set a non-default TCP Keep Alive time in seconds"`
	ForceTaskHostname       bool             `description:"Force to use the task's hostname."`
	Basic                   *Basic           `description:"Enable basic authentication"`
	RespectReadinessChecks  bool             `description:"Filter out tasks with non-successful readiness checks during deployments"`
	readyChecker            *readinessChecker
	marathonClient          marathon.Marathon
}

// Basic holds basic authentication specific configurations
type Basic struct {
	HTTPBasicAuthUser string `description:"Basic authentication User"`
	HTTPBasicPassword string `description:"Basic authentication Password"`
}

type lightMarathonClient interface {
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
		if p.Trace {
			config.LogOutput = log.CustomWriterLevel(logrus.DebugLevel, traceMaxScanTokenSize)
		}
		if p.Basic != nil {
			config.HTTPBasicAuthUser = p.Basic.HTTPBasicAuthUser
			config.HTTPBasicPassword = p.Basic.HTTPBasicPassword
		}
		var rc *readinessChecker
		if p.RespectReadinessChecks {
			log.Debug("Enabling Marathon readiness checker")
			rc = defaultReadinessChecker(p.Trace)
		}
		p.readyChecker = rc

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
			update, err := client.AddEventsListener(marathonEventIDs)
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
						log.Debugf("Received provider event %s", event)
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
		"getSubDomain":                p.getSubDomain,
		"getProtocol":                 p.getProtocol,
		"getPassHostHeader":           p.getPassHostHeader,
		"getPriority":                 p.getPriority,
		"getEntryPoints":              p.getEntryPoints,
		"getFrontendRule":             p.getFrontendRule,
		"getFrontendName":             p.getFrontendName,
		"hasCircuitBreakerLabels":     p.hasCircuitBreakerLabels,
		"hasLoadBalancerLabels":       p.hasLoadBalancerLabels,
		"hasMaxConnLabels":            p.hasMaxConnLabels,
		"getMaxConnExtractorFunc":     p.getMaxConnExtractorFunc,
		"getMaxConnAmount":            p.getMaxConnAmount,
		"getLoadBalancerMethod":       p.getLoadBalancerMethod,
		"getCircuitBreakerExpression": p.getCircuitBreakerExpression,
		"getSticky":                   p.getSticky,
		"hasHealthCheckLabels":        p.hasHealthCheckLabels,
		"getHealthCheckPath":          p.getHealthCheckPath,
		"getHealthCheckInterval":      p.getHealthCheckInterval,
		"hasServices":                 p.hasServices,
		"getServiceNames":             p.getServiceNames,
		"getServiceNameSuffix":        p.getServiceNameSuffix,
		"getBasicAuth":                p.getBasicAuth,
	}

	v := url.Values{}
	v.Add("embed", "apps.tasks")
	v.Add("embed", "apps.deployments")
	v.Add("embed", "apps.readiness")
	applications, err := p.marathonClient.Applications(v)
	if err != nil {
		log.Errorf("Failed to retrieve Marathon applications: %s", err)
		return nil
	}

	filteredApps := fun.Filter(p.applicationFilter, applications.Apps).([]marathon.Application)
	for i, app := range filteredApps {
		filteredApps[i].Tasks = fun.Filter(func(task *marathon.Task) bool {
			filtered := p.taskFilter(*task, app)
			if filtered {
				p.logIllegalServices(*task, app)
			}
			return filtered
		}, app.Tasks).([]*marathon.Task)
	}

	templateObjects := struct {
		Applications []marathon.Application
		Domain       string
	}{
		filteredApps,
		p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/marathon.tmpl", MarathonFuncMap, templateObjects)
	if err != nil {
		log.Errorf("failed to render Marathon configuration template: %s", err)
	}
	return configuration
}

func (p *Provider) applicationFilter(app marathon.Application) bool {
	// Filter disabled application.
	if !isApplicationEnabled(app, p.ExposedByDefault) {
		log.Debugf("Filtering disabled Marathon application %s", app.ID)
		return false
	}

	// Filter by constraints.
	label, _ := p.getAppLabel(app, types.LabelTags)
	constraintTags := strings.Split(label, ",")
	if p.MarathonLBCompatibility {
		if label, ok := p.getAppLabel(app, "HAPROXY_GROUP"); ok {
			constraintTags = append(constraintTags, label)
		}
	}
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Filtering Marathon application %v pruned by '%v' constraint", app.ID, failingConstraint.String())
		}
		return false
	}

	return true
}

func (p *Provider) taskFilter(task marathon.Task, application marathon.Application) bool {
	if task.State != string(taskStateRunning) {
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

	if ready := p.readyChecker.Do(task, application); !ready {
		log.Infof("Filtering unready task %s from application %s", task.ID, application.ID)
		return false
	}

	return true
}

func isApplicationEnabled(application marathon.Application, exposedByDefault bool) bool {
	return exposedByDefault && (*application.Labels)[types.LabelEnable] != "false" || (*application.Labels)[types.LabelEnable] == "true"
}

// logIllegalServices logs illegal service configurations.
// While we cannot filter on the service level, they will eventually get
// rejected once the server configuration is rendered.
func (p *Provider) logIllegalServices(task marathon.Task, application marathon.Application) {
	for _, serviceName := range p.getServiceNames(application) {
		// Check for illegal/missing ports.
		if _, err := p.processPorts(application, task, serviceName); err != nil {
			log.Warnf("%s has an illegal configuration: no proper port available", identifier(application, task, serviceName))
			continue
		}

		// Check for illegal port label combinations.
		_, hasPortLabel := p.getLabel(application, types.LabelPort, serviceName)
		_, hasPortIndexLabel := p.getLabel(application, types.LabelPortIndex, serviceName)
		if hasPortLabel && hasPortIndexLabel {
			log.Warnf("%s has both port and port index specified; port will take precedence", identifier(application, task, serviceName))
		}
	}
}

//servicePropertyValues is a map of services properties
//an example value is: weight=42
type servicePropertyValues map[string]string

//serviceProperties is a map of service properties per service, which we can get with label[serviceName][propertyName]. It yields a property value.
type serviceProperties map[string]servicePropertyValues

//hasServices checks if there are service-defining labels for the given application
func (p *Provider) hasServices(application marathon.Application) bool {
	return len(extractServiceProperties(application.Labels)) > 0
}

//extractServiceProperties extracts the service labels for the given application
func extractServiceProperties(labels *map[string]string) serviceProperties {
	v := make(serviceProperties)

	if labels != nil {
		for label, value := range *labels {
			matches := servicesPropertiesRegexp.FindStringSubmatch(label)
			if matches == nil {
				continue
			}

			// According to the regex, match index 1 is "service_name" and match index 2 is the "property_name"
			serviceName := matches[1]
			propertyName := matches[2]
			if _, ok := v[serviceName]; !ok {
				v[serviceName] = make(servicePropertyValues)
			}
			v[serviceName][propertyName] = value
		}
	}

	return v
}

//getServiceProperty returns the property for a service label searching in all labels of the given application
func getServiceProperty(application marathon.Application, serviceName string, property string) (string, bool) {
	value, ok := extractServiceProperties(application.Labels)[serviceName][property]
	return value, ok
}

//getServiceNames returns a list of service names for a given application
//An empty name "" will be added if no service specific properties exist, as an indication that there are no sub-services, but only main application
func (p *Provider) getServiceNames(application marathon.Application) []string {
	labelServiceProperties := extractServiceProperties(application.Labels)
	var names []string

	for k := range labelServiceProperties {
		names = append(names, k)
	}
	if len(names) == 0 {
		names = append(names, "")
	}
	return names
}

func (p *Provider) getServiceNameSuffix(serviceName string) string {
	if len(serviceName) > 0 {
		serviceName = strings.Replace(serviceName, "/", "-", -1)
		serviceName = strings.Replace(serviceName, ".", "-", -1)
		return "-service-" + serviceName
	}
	return ""
}

//getAppLabel is a convenience function to get application label, when no serviceName is available
//it is identical to calling getLabel(application, label, "")
func (p *Provider) getAppLabel(application marathon.Application, label string) (string, bool) {
	return p.getLabel(application, label, "")
}

//getLabel returns a string value of a corresponding `label` argument
//  If serviceName is non-empty, we look for a service label. If none exists or serviceName is empty, we look for an application label.
func (p *Provider) getLabel(application marathon.Application, label string, serviceName string) (string, bool) {
	if len(serviceName) > 0 {
		property := strings.TrimPrefix(label, types.LabelPrefix)
		if value, ok := getServiceProperty(application, serviceName, property); ok {
			return value, true
		}
	}
	for key, value := range *application.Labels {
		if key == label {
			return value, true
		}
	}
	return "", false
}

func (p *Provider) getPort(task marathon.Task, application marathon.Application, serviceName string) string {
	port, err := p.processPorts(application, task, serviceName)
	if err != nil {
		log.Errorf("Unable to process ports for %s: %s", identifier(application, task, serviceName), err)
		return ""
	}

	return strconv.Itoa(port)
}

func (p *Provider) getWeight(application marathon.Application, serviceName string) string {
	if label, ok := p.getLabel(application, types.LabelWeight, serviceName); ok {
		return label
	}
	return "0"
}

func (p *Provider) getDomain(application marathon.Application) string {
	if label, ok := p.getAppLabel(application, types.LabelDomain); ok {
		return label
	}
	return p.Domain
}

func (p *Provider) getProtocol(application marathon.Application, serviceName string) string {
	if label, ok := p.getLabel(application, types.LabelProtocol, serviceName); ok {
		return label
	}
	return "http"
}

func (p *Provider) getSticky(application marathon.Application) string {
	if sticky, ok := p.getAppLabel(application, types.LabelBackendLoadbalancerSticky); ok {
		return sticky
	}
	return "false"
}

func (p *Provider) getPassHostHeader(application marathon.Application, serviceName string) string {
	if passHostHeader, ok := p.getLabel(application, types.LabelFrontendPassHostHeader, serviceName); ok {
		return passHostHeader
	}
	return "true"
}

func (p *Provider) getPriority(application marathon.Application, serviceName string) string {
	if priority, ok := p.getLabel(application, types.LabelFrontendPriority, serviceName); ok {
		return priority
	}
	return "0"
}

func (p *Provider) getEntryPoints(application marathon.Application, serviceName string) []string {
	if entryPoints, ok := p.getLabel(application, types.LabelFrontendEntryPoints, serviceName); ok {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

// getFrontendRule returns the frontend rule for the specified application, using
// its label. If service is provided, it will look for serviceName label before generic one.
// It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(application marathon.Application, serviceName string) string {
	if label, ok := p.getLabel(application, types.LabelFrontendRule, serviceName); ok {
		return label
	}
	if p.MarathonLBCompatibility {
		if label, ok := p.getAppLabel(application, "HAPROXY_0_VHOST"); ok {
			return "Host:" + label
		}
	}
	if len(serviceName) > 0 {
		return "Host:" + strings.ToLower(provider.Normalize(serviceName)) + "." + p.getSubDomain(application.ID) + "." + p.Domain
	}
	return "Host:" + p.getSubDomain(application.ID) + "." + p.Domain
}

func (p *Provider) getBackend(application marathon.Application, serviceName string) string {
	if label, ok := p.getLabel(application, types.LabelBackend, serviceName); ok {
		return label
	}
	return strings.Replace(application.ID, "/", "-", -1) + p.getServiceNameSuffix(serviceName)
}

func (p *Provider) getFrontendName(application marathon.Application, serviceName string) string {
	appName := strings.Replace(application.ID, "/", "-", -1)
	return "frontend" + appName + p.getServiceNameSuffix(serviceName)
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
	_, ok := p.getAppLabel(application, types.LabelBackendCircuitbreakerExpression)
	return ok
}

func (p *Provider) hasLoadBalancerLabels(application marathon.Application) bool {
	_, errMethod := p.getAppLabel(application, types.LabelBackendLoadbalancerMethod)
	_, errSticky := p.getAppLabel(application, types.LabelBackendLoadbalancerSticky)
	return errMethod || errSticky
}

func (p *Provider) hasMaxConnLabels(application marathon.Application) bool {
	if _, ok := p.getAppLabel(application, types.LabelBackendMaxconnAmount); !ok {
		return false
	}
	_, ok := p.getAppLabel(application, types.LabelBackendMaxconnExtractorfunc)
	return ok
}

func (p *Provider) getMaxConnAmount(application marathon.Application) int64 {
	if label, ok := p.getAppLabel(application, types.LabelBackendMaxconnAmount); ok {
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
	if label, ok := p.getAppLabel(application, types.LabelBackendMaxconnExtractorfunc); ok {
		return label
	}
	return "request.host"
}

func (p *Provider) getLoadBalancerMethod(application marathon.Application) string {
	if label, ok := p.getAppLabel(application, types.LabelBackendLoadbalancerMethod); ok {
		return label
	}
	return "wrr"
}

func (p *Provider) getCircuitBreakerExpression(application marathon.Application) string {
	if label, ok := p.getAppLabel(application, types.LabelBackendCircuitbreakerExpression); ok {
		return label
	}
	return "NetworkErrorRatio() > 1"
}

func (p *Provider) hasHealthCheckLabels(application marathon.Application) bool {
	return p.getHealthCheckPath(application) != ""
}

func (p *Provider) getHealthCheckPath(application marathon.Application) string {
	if label, ok := p.getAppLabel(application, types.LabelBackendHealthcheckPath); ok {
		return label
	}
	return ""
}

func (p *Provider) getHealthCheckInterval(application marathon.Application) string {
	if label, ok := p.getAppLabel(application, types.LabelBackendHealthcheckInterval); ok {
		return label
	}
	return ""
}

func (p *Provider) getBasicAuth(application marathon.Application, serviceName string) []string {
	if basicAuth, ok := p.getLabel(application, types.LabelFrontendAuthBasic, serviceName); ok {
		return strings.Split(basicAuth, ",")
	}

	return []string{}
}

// processPorts returns the configured port.
// An explicitly specified port is preferred. If none is specified, it selects
// one of the available port. The first such found port is returned unless an
// optional index is provided.
func (p *Provider) processPorts(application marathon.Application, task marathon.Task, serviceName string) (int, error) {
	if portLabel, ok := p.getLabel(application, types.LabelPort, serviceName); ok {
		port, err := strconv.Atoi(portLabel)
		switch {
		case err != nil:
			return 0, fmt.Errorf("failed to parse port label %q: %s", portLabel, err)
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
	if portIndexLabel, ok := p.getLabel(application, types.LabelPortIndex, serviceName); ok {
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

func (p *Provider) getBackendServer(task marathon.Task, application marathon.Application) string {
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
		ipAddressIdxStr, ok := p.getAppLabel(application, "traefik.ipAddressIdx")
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
		return 0, fmt.Errorf("failed to parse index %q: %s", index, err)
	case parsed < 0, parsed > length-1:
		return 0, fmt.Errorf("index %d must be within range (0, %d)", parsed, length-1)
	}

	return parsed, nil
}

func identifier(app marathon.Application, task marathon.Task, serviceName string) string {
	id := fmt.Sprintf("Marathon task %s from application %s", task.ID, app.ID)
	if serviceName != "" {
		id += fmt.Sprintf(" (service: %s)", serviceName)
	}
	return id
}
