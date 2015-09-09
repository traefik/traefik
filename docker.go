package main
import (
	"github.com/fsouza/go-dockerclient"
	"github.com/leekchan/gtf"
	"bytes"
	"github.com/BurntSushi/toml"
	"log"
	"text/template"
	"strings"
)
var DockerFuncMap = template.FuncMap{
	"getBackend": func(container docker.Container) string {
		for key, value := range container.Config.Labels {
			if (key == "træfik.backend") {
				return value
			}
		}
		return container.Config.Hostname
	},
	"getPort": func(container docker.Container) string {
		for key, value := range container.Config.Labels {
			if (key == "træfik.port") {
				return value
			}
		}
		for key, _ := range container.NetworkSettings.Ports {
			return key.Port()
		}
		return ""
	},
	"getHost": getHost,
}
type DockerProvider struct {
	Watch        bool
	Endpoint     string
	dockerClient *docker.Client
}

func (provider *DockerProvider) Provide(configurationChan chan <- *Configuration) {
	provider.dockerClient, _ = docker.NewClient(provider.Endpoint)
	dockerEvents := make(chan *docker.APIEvents)
	if (provider.Watch) {
		provider.dockerClient.AddEventListener(dockerEvents)
	}
	go func() {
		for {
			event := <-dockerEvents
			log.Println("Event receveived", event)
			configuration := provider.loadDockerConfig()
			if (configuration != nil) {
				configurationChan <- configuration
			}
		}
	}()

	configuration := provider.loadDockerConfig()
	configurationChan <- configuration
}

func (provider *DockerProvider) loadDockerConfig() *Configuration {
	configuration := new(Configuration)
	containerList, _ := provider.dockerClient.ListContainers(docker.ListContainersOptions{})
	containersInspected := []docker.Container{}
	hosts := map[string][]docker.Container{}
	for _, container := range containerList {
		containerInspected, _ := provider.dockerClient.InspectContainer(container.ID)
		containersInspected = append(containersInspected, *containerInspected)
		hosts[getHost(*containerInspected)] = append(hosts[getHost(*containerInspected)], *containerInspected)
	}

	templateObjects := struct {
		Containers []docker.Container
		Hosts map[string][]docker.Container
	}{
		containersInspected,
		hosts,
	}
	gtf.Inject(DockerFuncMap)
	tmpl, err := template.New("docker.tmpl").Funcs(DockerFuncMap).ParseFiles("docker.tmpl")
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
		log.Println("Error creating docker configuration:", err)
		return nil
	}
	return configuration
}

func getHost(container docker.Container) string {
	for key, value := range container.Config.Labels {
		if (key == "træfik.host") {
			return value
		}
	}
	return strings.TrimPrefix(container.Name, "/")
}