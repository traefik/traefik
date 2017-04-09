package provider

import (
	"errors"
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
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
)

var _ Provider = (*Marathon)(nil)

// Marathon holds configuration of the Marathon provider.
type Marathon struct {
	BaseProvider
	Endpoint                string         `description:"Marathon server endpoint. You can also specify multiple endpoint for Marathon"`
	Domain                  string         `description:"Default domain used"`
	ExposedByDefault        bool           `description:"Expose Marathon apps by default"`
	GroupsAsSubDomains      bool           `description:"Convert Marathon groups to subdomains"`
	DCOSToken               string         `description:"DCOSToken for DCOS environment, This will override the Authorization header"`
	MarathonLBCompatibility bool           `description:"Add compatibility with marathon-lb labels"`
	TLS                     *ClientTLS     `description:"Enable Docker TLS support"`
	DialerTimeout           flaeg.Duration `description:"Set a non-default connection timeout for Marathon"`
	KeepAlive               flaeg.Duration `description:"Set a non-default TCP Keep Alive time in seconds"`
	Basic                   *MarathonBasic
	marathonClient          marathon.Marathon
}

// MarathonBasic holds basic authentication specific configurations
type MarathonBasic struct {
	HTTPBasicAuthUser string
	HTTPBasicPassword string
}

type lightMarathonClient interface {
	AllTasks(v url.Values) (*marathon.Tasks, error)
	Applications(url.Values) (*marathon.Applications, error)
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Marathon) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	provider.Constraints = append(provider.Constraints, constraints...)
	operation := func() error {
		config := marathon.NewDefaultConfig()
		config.URL = provider.Endpoint
		config.EventsTransport = marathon.EventsTransportSSE
		if provider.Basic != nil {
			config.HTTPBasicAuthUser = provider.Basic.HTTPBasicAuthUser
			config.HTTPBasicPassword = provider.Basic.HTTPBasicPassword
		}
		if len(provider.DCOSToken) > 0 {
			config.DCOSToken = provider.DCOSToken
		}
		TLSConfig, err := provider.TLS.CreateTLSConfig()
		if err != nil {
			return err
		}
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					KeepAlive: time.Duration(provider.KeepAlive),
					Timeout:   time.Duration(provider.DialerTimeout),
				}).DialContext,
				TLSClientConfig: TLSConfig,
			},
		}
		client, err := marathon.NewClient(config)
		if err != nil {
			log.Errorf("Failed to create a client for marathon, error: %s", err)
			return err
		}
		provider.marathonClient = client

		if provider.Watch {
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
						log.Debug("Marathon event receveived", event)
						configuration := provider.loadMarathonConfig()
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
		configuration := provider.loadMarathonConfig()
		configurationChan <- types.ConfigMessage{
			ProviderName:  "marathon",
			Configuration: configuration,
		}
		return nil
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("Marathon connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		log.Errorf("Cannot connect to Marathon server %+v", err)
	}
	return nil
}

func (provider *Marathon) loadMarathonConfig() *types.Configuration {
	var MarathonFuncMap = template.FuncMap{
		"getBackend":                  provider.getBackend,
		"getBackendServer":            provider.getBackendServer,
		"getPort":                     provider.getPort,
		"getWeight":                   provider.getWeight,
		"getDomain":                   provider.getDomain,
		"getProtocol":                 provider.getProtocol,
		"getPassHostHeader":           provider.getPassHostHeader,
		"getPriority":                 provider.getPriority,
		"getEntryPoints":              provider.getEntryPoints,
		"getFrontendRule":             provider.getFrontendRule,
		"getFrontendBackend":          provider.getFrontendBackend,
		"hasCircuitBreakerLabels":     provider.hasCircuitBreakerLabels,
		"hasLoadBalancerLabels":       provider.hasLoadBalancerLabels,
		"hasMaxConnLabels":            provider.hasMaxConnLabels,
		"getMaxConnExtractorFunc":     provider.getMaxConnExtractorFunc,
		"getMaxConnAmount":            provider.getMaxConnAmount,
		"getLoadBalancerMethod":       provider.getLoadBalancerMethod,
		"getCircuitBreakerExpression": provider.getCircuitBreakerExpression,
		"getSticky":                   provider.getSticky,
	}

	applications, err := provider.marathonClient.Applications(nil)
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	tasks, err := provider.marathonClient.AllTasks(&marathon.AllTasksOpts{Status: "running"})
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	//filter tasks
	filteredTasks := fun.Filter(func(task marathon.Task) bool {
		return provider.taskFilter(task, applications, provider.ExposedByDefault)
	}, tasks.Tasks).([]marathon.Task)

	//filter apps
	filteredApps := fun.Filter(func(app marathon.Application) bool {
		return provider.applicationFilter(app, filteredTasks)
	}, applications.Apps).([]marathon.Application)

	templateObjects := struct {
		Applications []marathon.Application
		Tasks        []marathon.Task
		Domain       string
	}{
		filteredApps,
		filteredTasks,
		provider.Domain,
	}

	configuration, err := provider.getConfiguration("templates/marathon.tmpl", MarathonFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func (provider *Marathon) taskFilter(task marathon.Task, applications *marathon.Applications, exposedByDefaultFlag bool) bool {
	application, err := getApplication(task, applications.Apps)
	if err != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return false
	}
	ports := processPorts(application, task)
	if len(ports) == 0 {
		log.Debug("Filtering marathon task without port %s", task.AppID)
		return false
	}
	label, _ := provider.getLabel(application, "traefik.tags")
	constraintTags := strings.Split(label, ",")
	if provider.MarathonLBCompatibility {
		if label, err := provider.getLabel(application, "HAPROXY_GROUP"); err == nil {
			constraintTags = append(constraintTags, label)
		}
	}
	if ok, failingConstraint := provider.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Application %v pruned by '%v' constraint", application.ID, failingConstraint.String())
		}
		return false
	}

	if !isApplicationEnabled(application, exposedByDefaultFlag) {
		log.Debugf("Filtering disabled marathon task %s", task.AppID)
		return false
	}

	//filter indeterminable task port
	portIndexLabel := (*application.Labels)["traefik.portIndex"]
	portValueLabel := (*application.Labels)["traefik.port"]
	if portIndexLabel != "" && portValueLabel != "" {
		log.Debugf("Filtering marathon task %s specifying both traefik.portIndex and traefik.port labels", task.AppID)
		return false
	}
	if portIndexLabel != "" {
		index, err := strconv.Atoi((*application.Labels)["traefik.portIndex"])
		if err != nil || index < 0 || index > len(ports)-1 {
			log.Debugf("Filtering marathon task %s with unexpected value for traefik.portIndex label", task.AppID)
			return false
		}
	}
	if portValueLabel != "" {
		_, err := strconv.Atoi((*application.Labels)["traefik.port"])
		if err != nil {
			log.Debugf("Filtering marathon task %s with unexpected value for traefik.port label", task.AppID)
			return false
		}
	}

	//filter healthchecks
	if application.HasHealthChecks() {
		if task.HasHealthCheckResults() {
			for _, healthcheck := range task.HealthCheckResults {
				// found one bad healthcheck, return false
				if !healthcheck.Alive {
					log.Debugf("Filtering marathon task %s with bad healthcheck", task.AppID)
					return false
				}
			}
		}
	}
	return true
}

func (provider *Marathon) applicationFilter(app marathon.Application, filteredTasks []marathon.Task) bool {
	label, _ := provider.getLabel(app, "traefik.tags")
	constraintTags := strings.Split(label, ",")
	if provider.MarathonLBCompatibility {
		if label, err := provider.getLabel(app, "HAPROXY_GROUP"); err == nil {
			constraintTags = append(constraintTags, label)
		}
	}
	if ok, failingConstraint := provider.MatchConstraints(constraintTags); !ok {
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

func (provider *Marathon) getLabel(application marathon.Application, label string) (string, error) {
	for key, value := range *application.Labels {
		if key == label {
			return value, nil
		}
	}
	return "", errors.New("Label not found:" + label)
}

func (provider *Marathon) getPort(task marathon.Task, applications []marathon.Application) string {
	application, err := getApplication(task, applications)
	if err != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return ""
	}
	ports := processPorts(application, task)
	if portIndexLabel, err := provider.getLabel(application, "traefik.portIndex"); err == nil {
		if index, err := strconv.Atoi(portIndexLabel); err == nil {
			return strconv.Itoa(ports[index])
		}
	}
	if portValueLabel, err := provider.getLabel(application, "traefik.port"); err == nil {
		return portValueLabel
	}

	for _, port := range ports {
		return strconv.Itoa(port)
	}
	return ""
}

func (provider *Marathon) getWeight(task marathon.Task, applications []marathon.Application) string {
	application, errApp := getApplication(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return "0"
	}
	if label, err := provider.getLabel(application, "traefik.weight"); err == nil {
		return label
	}
	return "0"
}

func (provider *Marathon) getDomain(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.domain"); err == nil {
		return label
	}
	return provider.Domain
}

func (provider *Marathon) getProtocol(task marathon.Task, applications []marathon.Application) string {
	application, errApp := getApplication(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return "http"
	}
	if label, err := provider.getLabel(application, "traefik.protocol"); err == nil {
		return label
	}
	return "http"
}

func (provider *Marathon) getSticky(application marathon.Application) string {
	if sticky, err := provider.getLabel(application, "traefik.backend.loadbalancer.sticky"); err == nil {
		return sticky
	}
	return "false"
}

func (provider *Marathon) getPassHostHeader(application marathon.Application) string {
	if passHostHeader, err := provider.getLabel(application, "traefik.frontend.passHostHeader"); err == nil {
		return passHostHeader
	}
	return "true"
}

func (provider *Marathon) getPriority(application marathon.Application) string {
	if priority, err := provider.getLabel(application, "traefik.frontend.priority"); err == nil {
		return priority
	}
	return "0"
}

func (provider *Marathon) getEntryPoints(application marathon.Application) []string {
	if entryPoints, err := provider.getLabel(application, "traefik.frontend.entryPoints"); err == nil {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

// getFrontendRule returns the frontend rule for the specified application, using
// it's label. It returns a default one (Host) if the label is not present.
func (provider *Marathon) getFrontendRule(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.frontend.rule"); err == nil {
		return label
	}
	if provider.MarathonLBCompatibility {
		if label, err := provider.getLabel(application, "HAPROXY_0_VHOST"); err == nil {
			return "Host:" + label
		}
	}
	return "Host:" + provider.getSubDomain(application.ID) + "." + provider.Domain
}

func (provider *Marathon) getBackend(task marathon.Task, applications []marathon.Application) string {
	application, errApp := getApplication(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return ""
	}
	return provider.getFrontendBackend(application)
}

func (provider *Marathon) getFrontendBackend(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.backend"); err == nil {
		return label
	}
	return replace("/", "-", application.ID)
}

func (provider *Marathon) getSubDomain(name string) string {
	if provider.GroupsAsSubDomains {
		splitedName := strings.Split(strings.TrimPrefix(name, "/"), "/")
		reverseStringSlice(&splitedName)
		reverseName := strings.Join(splitedName, ".")
		return reverseName
	}
	return strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1)
}

func (provider *Marathon) hasCircuitBreakerLabels(application marathon.Application) bool {
	if _, err := provider.getLabel(application, "traefik.backend.circuitbreaker.expression"); err != nil {
		return false
	}
	return true
}

func (provider *Marathon) hasLoadBalancerLabels(application marathon.Application) bool {
	_, errMethod := provider.getLabel(application, "traefik.backend.loadbalancer.method")
	_, errSticky := provider.getLabel(application, "traefik.backend.loadbalancer.sticky")
	if errMethod != nil && errSticky != nil {
		return false
	}
	return true
}

func (provider *Marathon) hasMaxConnLabels(application marathon.Application) bool {
	if _, err := provider.getLabel(application, "traefik.backend.maxconn.amount"); err != nil {
		return false
	}
	if _, err := provider.getLabel(application, "traefik.backend.maxconn.extractorfunc"); err != nil {
		return false
	}
	return true
}

func (provider *Marathon) getMaxConnAmount(application marathon.Application) int64 {
	if label, err := provider.getLabel(application, "traefik.backend.maxconn.amount"); err == nil {
		i, errConv := strconv.ParseInt(label, 10, 64)
		if errConv != nil {
			log.Errorf("Unable to parse traefik.backend.maxconn.amount %s", label)
			return math.MaxInt64
		}
		return i
	}
	return math.MaxInt64
}

func (provider *Marathon) getMaxConnExtractorFunc(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.backend.maxconn.extractorfunc"); err == nil {
		return label
	}
	return "request.host"
}

func (provider *Marathon) getLoadBalancerMethod(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.backend.loadbalancer.method"); err == nil {
		return label
	}
	return "wrr"
}

func (provider *Marathon) getCircuitBreakerExpression(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.backend.circuitbreaker.expression"); err == nil {
		return label
	}
	return "NetworkErrorRatio() > 1"
}

func processPorts(application marathon.Application, task marathon.Task) []int {

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

func (provider *Marathon) getBackendServer(task marathon.Task, applications []marathon.Application) string {
	application, err := getApplication(task, applications)
	if err != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return ""
	}
	if len(task.IPAddresses) == 0 {
		return ""
	} else if len(task.IPAddresses) == 1 {
		return task.IPAddresses[0].IPAddress
	} else {
		ipAddressIdxStr, err := provider.getLabel(application, "traefik.ipAddressIdx")
		if err != nil {
			log.Errorf("Unable to get marathon IPAddress from task %s", task.AppID)
			return ""
		}
		ipAddressIdx, err := strconv.Atoi(ipAddressIdxStr)
		if err != nil {
			log.Errorf("Invalid marathon IPAddress from task %s", task.AppID)
			return ""
		}
		return task.IPAddresses[ipAddressIdx].IPAddress
	}
}
