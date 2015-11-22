package provider

import (
	"errors"
	"net/url"
	"strconv"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/emilevauge/traefik/types"
	"github.com/gambol99/go-marathon"
)

// Marathon holds configuration of the Marathon provider.
type Marathon struct {
	baseProvider
	Endpoint         string
	Domain           string
	NetworkInterface string
	marathonClient   lightMarathonClient
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
	config.EventsInterface = provider.NetworkInterface
	client, err := marathon.NewClient(config)
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return err
	}
	provider.marathonClient = client
	update := make(marathon.EventsChannel, 5)
	if provider.Watch {
		if err := client.AddEventsListener(update, marathon.EVENTS_APPLICATIONS); err != nil {
			log.Errorf("Failed to register for subscriptions, %s", err)
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
		"getPort":           provider.getPort,
		"getWeight":         provider.getWeight,
		"getDomain":         provider.getDomain,
		"getProtocol":       provider.getProtocol,
		"getPassHostHeader": provider.getPassHostHeader,
		"getFrontendValue":  provider.getFrontendValue,
		"getFrontendRule":   provider.getFrontendRule,
		"replace":           replace,
	}

	applications, err := provider.marathonClient.Applications(nil)
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	tasks, err := provider.marathonClient.AllTasks((url.Values{"status": []string{"running"}}))
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
	application, errApp := getApplication(task, applications.Apps)
	if errApp != nil {
		log.Errorf("Unable to get marathon application from task %s", task.AppID)
		return false
	}
	_, err := strconv.Atoi(application.Labels["traefik.port"])
	if len(application.Ports) > 1 && err != nil {
		log.Debugf("Filtering marathon task %s with more than 1 port and no traefik.port label", task.AppID)
		return false
	}
	if application.Labels["traefik.enable"] == "false" {
		log.Debugf("Filtering disabled marathon task %s", task.AppID)
		return false
	}
	//filter healthchecks
	if application.HasHealthChecks() {
		if task.HasHealthCheckResults() {
			for _, healthcheck := range task.HealthCheckResult {
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

func (provider *Marathon) getPort(task marathon.Task) string {
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
