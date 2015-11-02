package provider

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/emilevauge/traefik/autogen"
	"github.com/emilevauge/traefik/types"
	"github.com/gambol99/go-marathon"
)

type Marathon struct {
	Watch            bool
	Endpoint         string
	marathonClient   marathon.Marathon
	Domain           string
	Filename         string
	NetworkInterface string
}

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
						configurationChan <- types.ConfigMessage{"marathon", configuration}
					}
				}
			}()
		}
	}

	configuration := provider.loadMarathonConfig()
	configurationChan <- types.ConfigMessage{"marathon", configuration}
	return nil
}

func (provider *Marathon) loadMarathonConfig() *types.Configuration {
	var MarathonFuncMap = template.FuncMap{
		"getPort": func(task marathon.Task) string {
			for _, port := range task.Ports {
				return strconv.Itoa(port)
			}
			return ""
		},
		"getWeight": func(task marathon.Task, applications []marathon.Application) string {
			application, errApp := getApplication(task, applications)
			if errApp != nil {
				log.Errorf("Unable to get marathon application from task %s", task.AppID)
				return "0"
			}
			if label, err := provider.getLabel(application, "traefik.weight"); err == nil {
				return label
			}
			return "0"
		},
		"getDomain": func(application marathon.Application) string {
			if label, err := provider.getLabel(application, "traefik.domain"); err == nil {
				return label
			}
			return provider.Domain
		},
		"replace": func(s1 string, s2 string, s3 string) string {
			return strings.Replace(s3, s1, s2, -1)
		},
		"getProtocol": func(task marathon.Task, applications []marathon.Application) string {
			application, errApp := getApplication(task, applications)
			if errApp != nil {
				log.Errorf("Unable to get marathon application from task %s", task.AppID)
				return "http"
			}
			if label, err := provider.getLabel(application, "traefik.protocol"); err == nil {
				return label
			}
			return "http"
		},
		"getPassHostHeader": func(application marathon.Application) string {
			if passHostHeader, err := provider.getLabel(application, "traefik.frontend.passHostHeader"); err == nil {
				return passHostHeader
			}
			return "false"
		},
		"getFrontendValue": provider.GetFrontendValue,
		"getFrontendRule":  provider.GetFrontendRule,
	}
	configuration := new(types.Configuration)

	applications, err := provider.marathonClient.Applications(nil)
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	tasks, err := provider.marathonClient.AllTasks()
	if err != nil {
		log.Errorf("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	//filter tasks
	filteredTasks := fun.Filter(func(task marathon.Task) bool {
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
	}, tasks.Tasks).([]marathon.Task)

	//filter apps
	filteredApps := fun.Filter(func(app marathon.Application) bool {
		//get ports from app tasks
		if !fun.Exists(func(task marathon.Task) bool {
			if task.AppID == app.ID {
				return true
			}
			return false
		}, filteredTasks) {
			return false
		}
		return true
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

	tmpl := template.New(provider.Filename).Funcs(MarathonFuncMap)
	if len(provider.Filename) > 0 {
		_, err := tmpl.ParseFiles(provider.Filename)
		if err != nil {
			log.Error("Error reading file", err)
			return nil
		}
	} else {
		buf, err := autogen.Asset("templates/marathon.tmpl")
		if err != nil {
			log.Error("Error reading file", err)
		}
		_, err = tmpl.Parse(string(buf))
		if err != nil {
			log.Error("Error reading file", err)
			return nil
		}
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		log.Error("Error with marathon template:", err)
		return nil
	}

	if _, err := toml.Decode(buffer.String(), configuration); err != nil {
		log.Error("Error creating marathon configuration:", err)
		return nil
	}

	return configuration
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

func (provider *Marathon) getEscapedName(name string) string {
	return strings.Replace(name, "/", "", -1)
}

func (provider *Marathon) GetFrontendValue(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.frontend.value"); err == nil {
		return label
	}
	return provider.getEscapedName(application.ID) + "." + provider.Domain
}

func (provider *Marathon) GetFrontendRule(application marathon.Application) string {
	if label, err := provider.getLabel(application, "traefik.frontend.rule"); err == nil {
		return label
	}
	return "Host"
}
