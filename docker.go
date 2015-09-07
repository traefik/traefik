package main
import (
	"github.com/fsouza/go-dockerclient"
	"github.com/leekchan/gtf"
	"bytes"
	"github.com/BurntSushi/toml"
	"log"
)

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
	for _, container := range containerList {
		containerInspected, _ := provider.dockerClient.InspectContainer(container.ID)
		containersInspected = append(containersInspected, *containerInspected)
	}
	containers := struct {
		Containers []docker.Container
	}{
		containersInspected,
	}
	tmpl, err := gtf.New("docker.tmpl").ParseFiles("docker.tmpl")
	if err != nil {
		log.Println("Error reading file:", err)
		return nil
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, containers)
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