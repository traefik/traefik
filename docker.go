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

type DockerProvider struct {
	Watch        bool
	Endpoint     string
	dockerClient *docker.Client
	Filename     string
	Domain       string
}

var DockerFuncMap = template.FuncMap{
	"getBackend": func(container docker.Container) string {
		for key, value := range container.Config.Labels {
			if (key == "traefik.backend") {
				return value
			}
		}
		return getHost(container)
	},
	"getPort": func(container docker.Container) string {
		for key, value := range container.Config.Labels {
			if (key == "traefik.port") {
				return value
			}
		}
		for key, _ := range container.NetworkSettings.Ports {
			return key.Port()
		}
		return ""
	},
	"getWeight": func(container docker.Container) string {
		for key, value := range container.Config.Labels {
			if (key == "traefik.weight") {
				return value
			}
		}
		return "0"
	},
	"replace": func(s1 string, s2 string, s3 string) string {
		return strings.Replace(s3, s1, s2, -1)
	},
	"getHost": getHost,
}

func (provider *DockerProvider) Provide(configurationChan chan <- *Configuration) {
	if client, err := docker.NewClient(provider.Endpoint); err != nil {
		log.Fatalf("Failed to create a client for docker, error: %s", err)
	} else {
		provider.dockerClient = client
		dockerEvents := make(chan *docker.APIEvents)
		if (provider.Watch) {
			provider.dockerClient.AddEventListener(dockerEvents)
			go func() {
				for {
					event := <-dockerEvents
					log.Println("Docker event receveived", event)
					configuration := provider.loadDockerConfig()
					if (configuration != nil) {
						configurationChan <- configuration
					}
				}
			}()
		}

		configuration := provider.loadDockerConfig()
		configurationChan <- configuration
	}
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
		Hosts      map[string][]docker.Container
		Domain     string
	}{
		containersInspected,
		hosts,
		provider.Domain,
	}
	gtf.Inject(DockerFuncMap)
	tmpl, err := template.New(provider.Filename).Funcs(DockerFuncMap).ParseFiles(provider.Filename)
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
		if (key == "traefik.host") {
			return value
		}
	}
	return strings.Replace(strings.Replace(container.Name, "/", "", -1), ".", "-", -1)
}