package main
import (
	"github.com/gambol99/go-marathon"
	"github.com/leekchan/gtf"
	"bytes"
	"github.com/BurntSushi/toml"
	"text/template"
	"strings"
	"strconv"
	"github.com/BurntSushi/ty/fun"
)

type MarathonProvider struct {
	Watch            bool
	Endpoint         string
	marathonClient   marathon.Marathon
	Domain           string
	Filename         string
	NetworkInterface string
}

var MarathonFuncMap = template.FuncMap{
	"getPort": func(task marathon.Task) string {
		for _, port := range task.Ports {
			return strconv.Itoa(port)
		}
		return ""
	},
	"getHost": func(application marathon.Application) string {
		for key, value := range application.Labels {
			if (key == "traefik.host") {
				return value
			}
		}
		return strings.Replace(application.ID, "/", "", 1)
	},
	"getWeight": func(application marathon.Application) string {
		for key, value := range application.Labels {
			if (key == "traefik.weight") {
				return value
			}
		}
		return "0"
	},
	"replace": func(s1 string, s2 string, s3 string) string {
		return strings.Replace(s3, s1, s2, -1)
	},
}

func (provider *MarathonProvider) Provide(configurationChan chan <- *Configuration) {
	config := marathon.NewDefaultConfig()
	config.URL = provider.Endpoint
	config.EventsInterface = provider.NetworkInterface
	if client, err := marathon.NewClient(config); err != nil {
		log.Error("Failed to create a client for marathon, error: %s", err)
		return
	} else {
		provider.marathonClient = client
		update := make(marathon.EventsChannel, 5)
		if (provider.Watch) {
			if err := client.AddEventsListener(update, marathon.EVENTS_APPLICATIONS); err != nil {
				log.Error("Failed to register for subscriptions, %s", err)
			} else {
				go func() {
					for {
						event := <-update
						log.Debug("Marathon event receveived", event)
						configuration := provider.loadMarathonConfig()
						if (configuration != nil) {
							configurationChan <- configuration
						}
					}
				}()
			}
		}

		configuration := provider.loadMarathonConfig()
		configurationChan <- configuration
	}
}

func (provider *MarathonProvider) loadMarathonConfig() *Configuration {
	configuration := new(Configuration)

	applications, err := provider.marathonClient.Applications(nil)
	if (err != nil) {
		log.Error("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	tasks, err := provider.marathonClient.AllTasks()
	if (err != nil) {
		log.Error("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	//filter tasks
	filteredTasks := fun.Filter(func(task marathon.Task) bool {
		if (len(task.Ports) == 0) {
			log.Debug("Filtering marathon task without port", task.AppID)
			return false
		}
		application := getApplication(task, applications.Apps)
		_, err := strconv.Atoi(application.Labels["traefik.port"])
		if (len(application.Ports) > 1 &&  err != nil) {
			log.Debug("Filtering marathon task with more than 1 port and no traefik.port label", task.AppID)
			return false
		}
		if (application.Labels["traefik.enable"] == "false") {
			log.Debug("Filtering disabled marathon task", task.AppID)
			return false
		}
		return true
	}, tasks.Tasks).([]marathon.Task)

	//filter apps
	filteredApps := fun.Filter(func(app marathon.Application) bool {
		//get ports from app tasks
		if (!fun.Exists(func(task marathon.Task) bool {
			if (task.AppID == app.ID) {
				return true
			}
			return false
		}, filteredTasks)) {
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

	gtf.Inject(MarathonFuncMap)
	tmpl := template.New(provider.Filename).Funcs(DockerFuncMap)
	if(len(provider.Filename) > 0){
		_, err := tmpl.ParseFiles(provider.Filename)
		if err != nil {
			log.Error("Error reading file", err)
			return nil
		}
	}else{
		buf, err := Asset("providerTemplates/marathon.tmpl")
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
		log.Error("Error with docker template:", err)
		return nil
	}

	if _, err := toml.Decode(buffer.String(), configuration); err != nil {
		log.Error("Error creating marathon configuration:", err)
		return nil
	}

	return configuration
}

func getApplication(task marathon.Task, apps []marathon.Application) *marathon.Application {
	for _, application := range apps {
		if (application.ID == task.AppID) {
			return &application
		}
	}
	return nil
}