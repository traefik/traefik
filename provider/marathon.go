package provider

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"crypto/tls"
	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
	"net/http"
)

// Marathon holds configuration of the Marathon provider.
type Marathon struct {
	BaseProvider   `mapstructure:",squash"`
	Endpoint       string
	Domain         string
	Basic          *MarathonBasic
	TLS            *tls.Config
	marathonClient marathon.Marathon
}

// MarathonBasic holds basic authentication specific configurations
type MarathonBasic struct {
	HTTPBasicAuthUser string
	HTTPBasicPassword string
}

type lightMarathonClient interface {
	Applications(url.Values) (*marathon.Applications, error)
	AllTasks(v url.Values) (*marathon.Tasks, error)
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Marathon) Provide(configurationChan chan<- types.ConfigMessage) error {
	config := marathon.NewDefaultConfig()
	config.URL = provider.Endpoint
	config.EventsTransport = marathon.EventsTransportSSE
	if provider.Basic != nil {
		config.HTTPBasicAuthUser = provider.Basic.HTTPBasicAuthUser
		config.HTTPBasicPassword = provider.Basic.HTTPBasicPassword
	}
	config.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: provider.TLS,
		},
	}
	client, err := marathon.NewClient(config)
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return err
	}
	provider.marathonClient = client
	update := make(marathon.EventsChannel, 5)
	if provider.Watch {
		if err := client.AddEventsListener(update, marathon.EVENTS_APPLICATIONS); err != nil {
			log.Errorf("Failed to register for events, %s", err)
		} else {
			go func() {
				for {
					event := <-update
					log.Debug("Marathon event receveived", event)
					configuration := provider.loadMarathonConfig()
					if configuration != nil {
						configurationChan <- types.ConfigMessage{
							ProviderName:  "marathon",
							Configuration: configuration,
						}
					}
				}
			}()
		}
	}

	configuration := provider.loadMarathonConfig()
	configurationChan <- types.ConfigMessage{
		ProviderName:  "marathon",
		Configuration: configuration,
	}
	return nil
}

func (provider *Marathon) loadMarathonConfig() *types.Configuration {
	var MarathonFuncMap = template.FuncMap{
		"getBackend":         provider.getBackend,
		"getPort":            provider.getPort,
		"getWeight":          provider.getWeight,
		"getDomain":          provider.getDomain,
		"getProtocol":        provider.getProtocol,
		"getPassHostHeader":  provider.getPassHostHeader,
		"getEntryPoints":     provider.getEntryPoints,
		"getFrontendValue":   provider.getFrontendValue,
		"getFrontendRule":    provider.getFrontendRule,
		"getFrontendBackend": provider.getFrontendBackend,
		"replace":            replace,
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
		return taskFilter(task, applications)
	}, tasks.Tasks).([]marathon.Task)

	//filter apps
	filteredApps := fun.Filter(func(app marathon.Application) bool {
		return applicationFilter(app, filteredTasks)
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

func taskFilter(task marathon.Task, applications *marathon.Applications) bool {
	if len(task.Ports) == 0 {
		log.Debug("Filtering marathon task without port %s", task.AppID)
		return false
	}
	application, err := getApplication(task, applications.Apps)
	if err != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return false
	}
	if application.Labels["traefik.enable"] == "false" {
		log.Debugf("Filtering disabled marathon task %s", task.AppID)
		return false
	}

	//filter indeterminable task port
	portIndexLabel := application.Labels["traefik.portIndex"]
	portValueLabel := application.Labels["traefik.port"]
	if portIndexLabel != "" && portValueLabel != "" {
		log.Debugf("Filtering marathon task %s specifying both traefik.portIndex and traefik.port labels", task.AppID)
		return false
	}
	if portIndexLabel == "" && portValueLabel == "" && len(application.Ports) > 1 {
		log.Debugf("Filtering marathon task %s with more than 1 port and no traefik.portIndex or traefik.port label", task.AppID)
		return false
	}
	if portIndexLabel != "" {
		index, err := strconv.Atoi(application.Labels["traefik.portIndex"])
		if err != nil || index < 0 || index > len(application.Ports)-1 {
			log.Debugf("Filtering marathon task %s with unexpected value for traefik.portIndex label", task.AppID)
			return false
		}
	}
	if portValueLabel != "" {
		port, err := strconv.Atoi(application.Labels["traefik.port"])
		if err != nil {
			log.Debugf("Filtering marathon task %s with unexpected value for traefik.port label", task.AppID)
			return false
		}

		var foundPort bool
		for _, exposedPort := range task.Ports {
			if port == exposedPort {
				foundPort = true
				break
			}
		}

		if !foundPort {
			log.Debugf("Filtering marathon task %s without a matching port for traefik.port label", task.AppID)
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
		} else {
			log.Debugf("Filtering marathon task %s with bad healthcheck", task.AppID)
			return false
		}
	}
	return true
}

func applicationFilter(app marathon.Application, filteredTasks []marathon.Task) bool {
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

func (provider *Marathon) getLabel(application marathon.Application, label string) (string, error) {
	for key, value := range application.Labels {
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

	if portIndexLabel, err := provider.getLabel(application, "traefik.portIndex"); err == nil {
		if index, err := strconv.Atoi(portIndexLabel); err == nil {
			return strconv.Itoa(task.Ports[index])
		}
	}
	if portValueLabel, err := provider.getLabel(application, "traefik.port"); err == nil {
		return portValueLabel
	}

	for _, port := range task.Ports {
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

func (provider *Marathon) getPassHostHeader(application marathon.Application) string {
	if passHostHeader, err := provider.getLabel(application, "traefik.frontend.passHostHeader"); err == nil {
		return passHostHeader
	}
	return "false"
}

func (provider *Marathon) getEntryPoints(application marathon.Application) []string {
	if entryPoints, err := provider.getLabel(application, "traefik.frontend.entryPoints"); err == nil {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

// getFrontendValue returns the frontend value for the specified application, using
// it's label. It returns a default one if the label is not present.
func (provider *Marathon) getFrontendValue(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.frontend.value"); err == nil {
		return label
	}
	return getEscapedName(application.ID) + "." + provider.Domain
}

// getFrontendRule returns the frontend rule for the specified application, using
// it's label. It returns a default one (Host) if the label is not present.
func (provider *Marathon) getFrontendRule(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.frontend.rule"); err == nil {
		return label
	}
	return "Host"
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
