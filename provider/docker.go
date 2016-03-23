package provider

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/types"
	"github.com/fsouza/go-dockerclient"
)

// Docker holds configurations of the Docker provider.
type Docker struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string
	Domain       string
	TLS          *DockerTLS
}

// DockerTLS holds TLS specific configurations
type DockerTLS struct {
	CA                 string
	Cert               string
	Key                string
	InsecureSkipVerify bool
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Docker) Provide(configurationChan chan<- types.ConfigMessage) error {
	go func() {
		operation := func() error {
			var dockerClient *docker.Client
			var err error

			if provider.TLS != nil {
				dockerClient, err = docker.NewTLSClient(provider.Endpoint,
					provider.TLS.Cert, provider.TLS.Key, provider.TLS.CA)
				if err == nil {
					dockerClient.TLSConfig.InsecureSkipVerify = provider.TLS.InsecureSkipVerify
				}
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
			configuration := provider.loadDockerConfig(listContainers(dockerClient))
			configurationChan <- types.ConfigMessage{
				ProviderName:  "docker",
				Configuration: configuration,
			}
			if provider.Watch {
				dockerEvents := make(chan *docker.APIEvents)
				dockerClient.AddEventListener(dockerEvents)
				log.Debug("Docker listening")
				for {
					event := <-dockerEvents
					if event == nil {
						return errors.New("Docker event nil")
						//							log.Fatalf("Docker connection error")
					}
					if event.Status == "start" || event.Status == "die" {
						log.Debugf("Docker event receveived %+v", event)
						configuration := provider.loadDockerConfig(listContainers(dockerClient))
						if configuration != nil {
							configurationChan <- types.ConfigMessage{
								ProviderName:  "docker",
								Configuration: configuration,
							}
						}
					}
				}
			}
			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Docker connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
		if err != nil {
			log.Fatalf("Cannot connect to docker server %+v", err)
		}
	}()

	return nil
}

func (provider *Docker) loadDockerConfig(containersInspected []docker.Container) *types.Configuration {
	var DockerFuncMap = template.FuncMap{
		"getBackend":        provider.getBackend,
		"getPort":           provider.getPort,
		"getWeight":         provider.getWeight,
		"getDomain":         provider.getDomain,
		"getProtocol":       provider.getProtocol,
		"getPassHostHeader": provider.getPassHostHeader,
		"getEntryPoints":    provider.getEntryPoints,
		"getFrontendValue":  provider.getFrontendValue,
		"getFrontendRule":   provider.getFrontendRule,
		"replace":           replace,
	}

	// filter containers
	filteredContainers := fun.Filter(containerFilter, containersInspected).([]docker.Container)

	frontends := map[string][]docker.Container{}
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

	configuration, err := provider.getConfiguration("templates/docker.tmpl", DockerFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func containerFilter(container docker.Container) bool {
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

	labels, err := getLabels(container, []string{"traefik.frontend.rule", "traefik.frontend.value"})
	if len(labels) != 0 && err != nil {
		log.Debugf("Filtering bad labeled container %s", container.Name)
		return false
	}

	return true
}

func (provider *Docker) getFrontendName(container docker.Container) string {
	// Replace '.' with '-' in quoted keys because of this issue https://github.com/BurntSushi/toml/issues/78
	frontendName := fmt.Sprintf("%s-%s", provider.getFrontendRule(container), provider.getFrontendValue(container))
	frontendName = strings.Replace(frontendName, "[", "", -1)
	frontendName = strings.Replace(frontendName, "]", "", -1)

	return strings.Replace(frontendName, ".", "-", -1)
}

// GetFrontendValue returns the frontend value for the specified container, using
// it's label. It returns a default one if the label is not present.
func (provider *Docker) getFrontendValue(container docker.Container) string {
	if label, err := getLabel(container, "traefik.frontend.value"); err == nil {
		return label
	}
	return getEscapedName(container.Name) + "." + provider.Domain
}

// GetFrontendRule returns the frontend rule for the specified container, using
// it's label. It returns a default one (Host) if the label is not present.
func (provider *Docker) getFrontendRule(container docker.Container) string {
	if label, err := getLabel(container, "traefik.frontend.rule"); err == nil {
		return label
	}
	return "Host"
}

func (provider *Docker) getBackend(container docker.Container) string {
	if label, err := getLabel(container, "traefik.backend"); err == nil {
		return label
	}
	return getEscapedName(container.Name)
}

func (provider *Docker) getPort(container docker.Container) string {
	if label, err := getLabel(container, "traefik.port"); err == nil {
		return label
	}
	for key := range container.NetworkSettings.Ports {
		return key.Port()
	}
	return ""
}

func (provider *Docker) getWeight(container docker.Container) string {
	if label, err := getLabel(container, "traefik.weight"); err == nil {
		return label
	}
	return "0"
}

func (provider *Docker) getDomain(container docker.Container) string {
	if label, err := getLabel(container, "traefik.domain"); err == nil {
		return label
	}
	return provider.Domain
}

func (provider *Docker) getProtocol(container docker.Container) string {
	if label, err := getLabel(container, "traefik.protocol"); err == nil {
		return label
	}
	return "http"
}

func (provider *Docker) getPassHostHeader(container docker.Container) string {
	if passHostHeader, err := getLabel(container, "traefik.frontend.passHostHeader"); err == nil {
		return passHostHeader
	}
	return "false"
}

func (provider *Docker) getEntryPoints(container docker.Container) []string {
	if entryPoints, err := getLabel(container, "traefik.frontend.entryPoints"); err == nil {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

func getLabel(container docker.Container, label string) (string, error) {
	for key, value := range container.Config.Labels {
		if key == label {
			return value, nil
		}
	}
	return "", errors.New("Label not found:" + label)
}

func getLabels(container docker.Container, labels []string) (map[string]string, error) {
	var globalErr error
	foundLabels := map[string]string{}
	for _, label := range labels {
		foundLabel, err := getLabel(container, label)
		// Error out only if one of them is defined.
		if err != nil {
			globalErr = errors.New("Label not found: " + label)
			continue
		}
		foundLabels[label] = foundLabel

	}
	return foundLabels, globalErr
}

func listContainers(dockerClient *docker.Client) []docker.Container {
	containerList, _ := dockerClient.ListContainers(docker.ListContainersOptions{})
	containersInspected := []docker.Container{}

	// get inspect containers
	for _, container := range containerList {
		containerInspected, _ := dockerClient.InspectContainer(container.ID)
		containersInspected = append(containersInspected, *containerInspected)
	}
	return containersInspected
}
