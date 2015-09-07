package main
import(
	"github.com/fsouza/go-dockerclient"
	"fmt"
	"github.com/leekchan/gtf"
	"bytes"
	"github.com/BurntSushi/toml"
)

type DockerProvider struct {
	dockerClient *docker.Client
}

func (provider *DockerProvider) Provide(serviceChan chan<- Service){
	endpoint := "unix:///var/run/docker.sock"
	provider.dockerClient, _ = docker.NewClient(endpoint)
	dockerEvents := make(chan *docker.APIEvents)
	provider.dockerClient.AddEventListener(dockerEvents)
	go func() {
		for {
			event := <-dockerEvents
			fmt.Println("Event receveived", event)
			service:= provider.loadDockerConfig()
			serviceChan <- *service
		}
	}()

	service:= provider.loadDockerConfig()
	fmt.Println("Sending service")
	serviceChan <- *service
}

func (provider *DockerProvider) loadDockerConfig() *Service {
	service := new(Service)
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
		panic(err)
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, containers)
	if err != nil {
		panic(err)
	}

	fmt.Println(buffer.String())

	if _, err := toml.Decode(buffer.String(), service); err != nil {
		fmt.Println(err)
		return nil
	}
	return service
}