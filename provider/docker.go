package provider

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/emilevauge/traefik/autogen"
	"github.com/emilevauge/traefik/types"
	"github.com/fsouza/go-dockerclient"
)

// Docker holds configurations of the Docker provider.
type Docker struct {
	Watch    bool
	Endpoint string
	Filename string
	Domain   string
	TLS      bool
	CA       string
	Cert     string
	Key      string
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Docker) Provide(configurationChan chan<- types.ConfigMessage) error {

	var dockerClient *docker.Client
	var err error

	if provider.TLS {
		dockerClient, err = docker.NewTLSClient(provider.Endpoint, provider.Cert, provider.Key, provider.CA)
	} else {
		dockerClient, err = docker.NewClient(provider.Endpoint)
	}
	if err != nil {
		log.Errorf("Failed to create a client for docker, error: %s", err)
		return err
	}
	err = dockerClient.Ping()
	if err != nil {
		log.Errorf("Docker connection error %+v", err)
		return err
	}
	log.Debug("Docker connection established")
	if provider.Watch {
		dockerEvents := make(chan *docker.APIEvents)
		dockerClient.AddEventListener(dockerEvents)
		log.Debug("Docker listening")
		go func() {
			operation := func() error {
				for {
					event := <-dockerEvents
					if event == nil {
						return errors.New("Docker event nil")
						//							log.Fatalf("Docker connection error")
					}
					if event.Status == "start" || event.Status == "die" {
						log.Debugf("Docker event receveived %+v", event)
						configuration := provider.loadDockerConfig(dockerClient)
						if configuration != nil {
							configurationChan <- types.ConfigMessage{"docker", configuration}
						}
					}
				}
			}
			notify := func(err error, time time.Duration) {
				log.Errorf("Docker connection error %+v, retrying in %s", err, time)
			}
			err := backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
			if err != nil {
				log.Fatalf("Cannot connect to docker server %+v", err)
			}
		}()
	}

	configuration := provider.loadDockerConfig(dockerClient)
	configurationChan <- types.ConfigMessage{"docker", configuration}
	return nil
}

func (provider *Docker) loadDockerConfig(dockerClient *docker.Client) *types.Configuration {
	var DockerFuncMap = template.FuncMap{
		"getBackend": func(container docker.Container) string {
			if label, err := provider.getLabel(container, "traefik.backend"); err == nil {
				return label
			}
			return provider.getEscapedName(container.Name)
		},
		"getPort": func(container docker.Container) string {
			if label, err := provider.getLabel(container, "traefik.port"); err == nil {
				return label
			}
			for key := range container.NetworkSettings.Ports {
				return key.Port()
			}
			return ""
		},
		"getWeight": func(container docker.Container) string {
			if label, err := provider.getLabel(container, "traefik.weight"); err == nil {
				return label
			}
			return "0"
		},
		"getDomain": func(container docker.Container) string {
			if label, err := provider.getLabel(container, "traefik.domain"); err == nil {
				return label
			}
			return provider.Domain
		},
		"getProtocol": func(container docker.Container) string {
			if label, err := provider.getLabel(container, "traefik.protocol"); err == nil {
				return label
			}
			return "http"
		},
		"getPassHostHeader": func(container docker.Container) string {
			if passHostHeader, err := provider.getLabel(container, "traefik.frontend.passHostHeader"); err == nil {
				return passHostHeader
			}
			return "false"
		},
		"getFrontendValue": provider.GetFrontendValue,
		"getFrontendRule":  provider.GetFrontendRule,
		"replace": func(s1 string, s2 string, s3 string) string {
			return strings.Replace(s3, s1, s2, -1)
		},
	}
	configuration := new(types.Configuration)
	containerList, _ := dockerClient.ListContainers(docker.ListContainersOptions{})
	containersInspected := []docker.Container{}
	frontends := map[string][]docker.Container{}

	// get inspect containers
	for _, container := range containerList {
		containerInspected, _ := dockerClient.InspectContainer(container.ID)
		containersInspected = append(containersInspected, *containerInspected)
	}

	// filter containers
	filteredContainers := fun.Filter(func(container docker.Container) bool {
		if len(container.NetworkSettings.Ports) == 0 {
			log.Debugf("Filtering container without port %s", container.Name)
			return false
		}
		_, err := strconv.Atoi(container.Config.Labels["traefik.port"])
		if len(container.NetworkSettings.Ports) > 1 && err != nil {
			log.Debugf("Filtering container with more than 1 port and no traefik.port label %s", container.Name)
			return false
		}
		if container.Config.Labels["traefik.enable"] == "false" {
			log.Debugf("Filtering disabled container %s", container.Name)
			return false
		}

		labels, err := provider.getLabels(container, []string{"traefik.frontend.rule", "traefik.frontend.value"})
		if len(labels) != 0 && err != nil {
			log.Debugf("Filtering bad labeled container %s", container.Name)
			return false
		}

		return true
	}, containersInspected).([]docker.Container)

	for _, container := range filteredContainers {
		frontends[provider.getFrontendName(container)] = append(frontends[provider.getFrontendName(container)], container)
	}

	templateObjects := struct {
		Containers []docker.Container
		Frontends  map[string][]docker.Container
		Domain     string
	}{
		filteredContainers,
		frontends,
		provider.Domain,
	}
	tmpl := template.New(provider.Filename).Funcs(DockerFuncMap)
	if len(provider.Filename) > 0 {
		_, err := tmpl.ParseFiles(provider.Filename)
		if err != nil {
			log.Error("Error reading file", err)
			return nil
		}
	} else {
		buf, err := autogen.Asset("templates/docker.tmpl")
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
	err := tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		log.Error("Error with docker template", err)
		return nil
	}

	if _, err := toml.Decode(buffer.String(), configuration); err != nil {
		log.Error("Error creating docker configuration ", err)
		return nil
	}
	return configuration
}

func (provider *Docker) getFrontendName(container docker.Container) string {
	// Replace '.' with '-' in quoted keys because of this issue https://github.com/BurntSushi/toml/issues/78
	frontendName := fmt.Sprintf("%s-%s", provider.GetFrontendRule(container), provider.GetFrontendValue(container))
	frontendName = strings.Replace(frontendName, "[", "", -1)
	frontendName = strings.Replace(frontendName, "]", "", -1)

	return strings.Replace(frontendName, ".", "-", -1)
}

func (provider *Docker) getEscapedName(name string) string {
	return strings.Replace(name, "/", "", -1)
}

func (provider *Docker) getLabel(container docker.Container, label string) (string, error) {
	for key, value := range container.Config.Labels {
		if key == label {
			return value, nil
		}
	}
	return "", errors.New("Label not found:" + label)
}

func (provider *Docker) getLabels(container docker.Container, labels []string) (map[string]string, error) {
	var globalErr error
	foundLabels := map[string]string{}
	for _, label := range labels {
		foundLabel, err := provider.getLabel(container, label)
		// Error out only if one of them is defined.
		if err != nil {
			globalErr = errors.New("Label not found: " + label)
			continue
		}
		foundLabels[label] = foundLabel

	}
	return foundLabels, globalErr
}

// GetFrontendValue returns the frontend value for the specified container, using
// it's label. It returns a default one if the label is not present.
func (provider *Docker) GetFrontendValue(container docker.Container) string {
	if label, err := provider.getLabel(container, "traefik.frontend.value"); err == nil {
		return label
	}
	return provider.getEscapedName(container.Name) + "." + provider.Domain
}

// GetFrontendRule returns the frontend rule for the specified container, using
// it's label. It returns a default one (Host) if the label is not present.
func (provider *Docker) GetFrontendRule(container docker.Container) string {
	if label, err := provider.getLabel(container, "traefik.frontend.rule"); err == nil {
		return label
	}
	return "Host"
}
