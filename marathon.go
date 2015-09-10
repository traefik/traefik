package main
import (
	"github.com/gambol99/go-marathon"
	"log"
	"github.com/leekchan/gtf"
	"bytes"
	"github.com/BurntSushi/toml"
	"text/template"
	"strings"
	"strconv"
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
			if (key == "træfik.host") {
				return value
			}
		}
		return strings.Replace(application.ID, "/", "", 1)
	},
	"getWeight": func(application marathon.Application) string {
		for key, value := range application.Labels {
			if (key == "træfik.weight") {
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
		log.Println("Failed to create a client for marathon, error: %s", err)
		return
	} else {
		provider.marathonClient = client
		update := make(marathon.EventsChannel, 5)
		if (provider.Watch) {
			if err := client.AddEventsListener(update, marathon.EVENTS_APPLICATIONS); err != nil {
				log.Println("Failed to register for subscriptions, %s", err)
			} else {
				go func() {
					for {
						event := <-update
						log.Println("Marathon event receveived", event)
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
		log.Println("Failed to create a client for marathon, error: %s", err)
		return nil
	}
	tasks, err := provider.marathonClient.AllTasks()
	if (err != nil) {
		log.Println("Failed to create a client for marathon, error: %s", err)
		return nil
	}

	templateObjects := struct {
		Applications []marathon.Application
		Tasks        []marathon.Task
		Domain       string
	}{
		applications.Apps,
		tasks.Tasks,
		provider.Domain,
	}

	gtf.Inject(MarathonFuncMap)
	tmpl, err := template.New(provider.Filename).Funcs(MarathonFuncMap).ParseFiles(provider.Filename)
	if err != nil {
		log.Println("Error reading file:", err)
		return nil
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		log.Println("Error with docker template:", err)
		return nil
	}

	if _, err := toml.Decode(buffer.String(), configuration); err != nil {
		log.Println("Error creating marathon configuration:", err)
		return nil
	}

	return configuration
}