package provider

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/context"

	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/utils"
	"github.com/containous/traefik/version"
	"github.com/docker/engine-api/client"
	dockertypes "github.com/docker/engine-api/types"
	eventtypes "github.com/docker/engine-api/types/events"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/go-connections/sockets"
	"github.com/vdemeester/docker-events"
)

// DockerAPIVersion is a constant holding the version of the Docker API traefik will use
const DockerAPIVersion string = "1.21"

// Docker holds configurations of the Docker provider.
type Docker struct {
	BaseProvider     `mapstructure:",squash"`
	Endpoint         string     `description:"Docker server endpoint. Can be a tcp or a unix socket endpoint"`
	Domain           string     `description:"Default domain used"`
	TLS              *ClientTLS `description:"Enable Docker TLS support"`
	ExposedByDefault bool       `description:"Expose containers by default"`
}

func (provider *Docker) createClient() (client.APIClient, error) {
	var httpClient *http.Client
	httpHeaders := map[string]string{
		"User-Agent": "Traefik " + version.Version,
	}
	if provider.TLS != nil {
		config, err := provider.TLS.CreateTLSConfig()
		if err != nil {
			return nil, err
		}
		tr := &http.Transport{
			TLSClientConfig: config,
		}
		proto, addr, _, err := client.ParseHost(provider.Endpoint)
		if err != nil {
			return nil, err
		}

		sockets.ConfigureTransport(tr, proto, addr)

		httpClient = &http.Client{
			Transport: tr,
		}

	}
	return client.NewClient(provider.Endpoint, DockerAPIVersion, httpClient, httpHeaders)
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Docker) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints []types.Constraint) error {
	provider.Constraints = append(provider.Constraints, constraints...)
	// TODO register this routine in pool, and watch for stop channel
	safe.Go(func() {
		operation := func() error {
			var err error

			dockerClient, err := provider.createClient()
			if err != nil {
				log.Errorf("Failed to create a client for docker, error: %s", err)
				return err
			}

			ctx := context.Background()
			version, err := dockerClient.ServerVersion(ctx)
			log.Debugf("Docker connection established with docker %s (API %s)", version.Version, version.APIVersion)
			containers, err := listContainers(ctx, dockerClient)
			if err != nil {
				log.Errorf("Failed to list containers for docker, error %s", err)
				return err
			}
			configuration := provider.loadDockerConfig(containers)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "docker",
				Configuration: configuration,
			}
			if provider.Watch {
				ctx, cancel := context.WithCancel(ctx)
				pool.Go(func(stop chan bool) {
					for {
						select {
						case <-stop:
							cancel()
							return
						}
					}
				})
				f := filters.NewArgs()
				f.Add("type", "container")
				options := dockertypes.EventsOptions{
					Filters: f,
				}
				eventHandler := events.NewHandler(events.ByAction)
				startStopHandle := func(m eventtypes.Message) {
					log.Debugf("Docker event received %+v", m)
					containers, err := listContainers(ctx, dockerClient)
					if err != nil {
						log.Errorf("Failed to list containers for docker, error %s", err)
						// Call cancel to get out of the monitor
						cancel()
						return
					}
					configuration := provider.loadDockerConfig(containers)
					if configuration != nil {
						configurationChan <- types.ConfigMessage{
							ProviderName:  "docker",
							Configuration: configuration,
						}
					}
				}
				eventHandler.Handle("start", startStopHandle)
				eventHandler.Handle("die", startStopHandle)

				errChan := events.MonitorWithHandler(ctx, dockerClient, options, eventHandler)
				if err := <-errChan; err != nil {
					return err
				}
			}
			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Docker connection error %+v, retrying in %s", err, time)
		}
		err := utils.RetryNotifyJob(operation, backoff.NewExponentialBackOff(), notify)
		if err != nil {
			log.Errorf("Cannot connect to docker server %+v", err)
		}
	})

	return nil
}

func (provider *Docker) loadDockerConfig(containersInspected []dockertypes.ContainerJSON) *types.Configuration {
	var DockerFuncMap = template.FuncMap{
		"getBackend":        provider.getBackend,
		"getIPAddress":      provider.getIPAddress,
		"getPort":           provider.getPort,
		"getWeight":         provider.getWeight,
		"getDomain":         provider.getDomain,
		"getProtocol":       provider.getProtocol,
		"getPassHostHeader": provider.getPassHostHeader,
		"getPriority":       provider.getPriority,
		"getEntryPoints":    provider.getEntryPoints,
		"getFrontendRule":   provider.getFrontendRule,
		"replace":           replace,
	}

	// filter containers
	filteredContainers := fun.Filter(func(container dockertypes.ContainerJSON) bool {
		return provider.containerFilter(container, provider.ExposedByDefault)
	}, containersInspected).([]dockertypes.ContainerJSON)

	frontends := map[string][]dockertypes.ContainerJSON{}
	for _, container := range filteredContainers {
		frontendName := provider.getFrontendName(container)
		frontends[frontendName] = append(frontends[frontendName], container)
	}

	templateObjects := struct {
		Containers []dockertypes.ContainerJSON
		Frontends  map[string][]dockertypes.ContainerJSON
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

func (provider *Docker) containerFilter(container dockertypes.ContainerJSON, exposedByDefaultFlag bool) bool {
	_, err := strconv.Atoi(container.Config.Labels["traefik.port"])
	if len(container.NetworkSettings.Ports) == 0 && err != nil {
		log.Debugf("Filtering container without port and no traefik.port label %s", container.Name)
		return false
	}
	if len(container.NetworkSettings.Ports) > 1 && err != nil {
		log.Debugf("Filtering container with more than 1 port and no traefik.port label %s", container.Name)
		return false
	}

	if !isContainerEnabled(container, exposedByDefaultFlag) {
		log.Debugf("Filtering disabled container %s", container.Name)
		return false
	}

	constraintTags := strings.Split(container.Config.Labels["traefik.tags"], ",")
	if ok, failingConstraint := provider.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Container %v pruned by '%v' constraint", container.Name, failingConstraint.String())
		}
		return false
	}

	return true
}

func (provider *Docker) getFrontendName(container dockertypes.ContainerJSON) string {
	// Replace '.' with '-' in quoted keys because of this issue https://github.com/BurntSushi/toml/issues/78
	return normalize(provider.getFrontendRule(container))
}

// GetFrontendRule returns the frontend rule for the specified container, using
// it's label. It returns a default one (Host) if the label is not present.
func (provider *Docker) getFrontendRule(container dockertypes.ContainerJSON) string {
	if label, err := getLabel(container, "traefik.frontend.rule"); err == nil {
		return label
	}
	return "Host:" + provider.getSubDomain(container.Name) + "." + provider.Domain
}

func (provider *Docker) getBackend(container dockertypes.ContainerJSON) string {
	if label, err := getLabel(container, "traefik.backend"); err == nil {
		return label
	}
	return normalize(container.Name)
}

func (provider *Docker) getIPAddress(container dockertypes.ContainerJSON) string {
	if label, err := getLabel(container, "traefik.docker.network"); err == nil && label != "" {
		networks := container.NetworkSettings.Networks
		if networks != nil {
			network := networks[label]
			if network != nil {
				return network.IPAddress
			}
		}
	}

	// If net==host, quick n' dirty, we return 127.0.0.1
	// This will work locally, but will fail with swarm.
	if container.HostConfig != nil && "host" == container.HostConfig.NetworkMode {
		return "127.0.0.1"
	}

	for _, network := range container.NetworkSettings.Networks {
		return network.IPAddress
	}
	return ""
}

func (provider *Docker) getPort(container dockertypes.ContainerJSON) string {
	if label, err := getLabel(container, "traefik.port"); err == nil {
		return label
	}
	for key := range container.NetworkSettings.Ports {
		return key.Port()
	}
	return ""
}

func (provider *Docker) getWeight(container dockertypes.ContainerJSON) string {
	if label, err := getLabel(container, "traefik.weight"); err == nil {
		return label
	}
	return "1"
}

func (provider *Docker) getDomain(container dockertypes.ContainerJSON) string {
	if label, err := getLabel(container, "traefik.domain"); err == nil {
		return label
	}
	return provider.Domain
}

func (provider *Docker) getProtocol(container dockertypes.ContainerJSON) string {
	if label, err := getLabel(container, "traefik.protocol"); err == nil {
		return label
	}
	return "http"
}

func (provider *Docker) getPassHostHeader(container dockertypes.ContainerJSON) string {
	if passHostHeader, err := getLabel(container, "traefik.frontend.passHostHeader"); err == nil {
		return passHostHeader
	}
	return "true"
}

func (provider *Docker) getPriority(container dockertypes.ContainerJSON) string {
	if priority, err := getLabel(container, "traefik.frontend.priority"); err == nil {
		return priority
	}
	return "0"
}

func (provider *Docker) getEntryPoints(container dockertypes.ContainerJSON) []string {
	if entryPoints, err := getLabel(container, "traefik.frontend.entryPoints"); err == nil {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

func isContainerEnabled(container dockertypes.ContainerJSON, exposedByDefault bool) bool {
	return exposedByDefault && container.Config.Labels["traefik.enable"] != "false" || container.Config.Labels["traefik.enable"] == "true"
}

func getLabel(container dockertypes.ContainerJSON, label string) (string, error) {
	for key, value := range container.Config.Labels {
		if key == label {
			return value, nil
		}
	}
	return "", errors.New("Label not found:" + label)
}

func getLabels(container dockertypes.ContainerJSON, labels []string) (map[string]string, error) {
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

func listContainers(ctx context.Context, dockerClient client.ContainerAPIClient) ([]dockertypes.ContainerJSON, error) {
	containerList, err := dockerClient.ContainerList(ctx, dockertypes.ContainerListOptions{})
	if err != nil {
		return []dockertypes.ContainerJSON{}, err
	}
	containersInspected := []dockertypes.ContainerJSON{}

	// get inspect containers
	for _, container := range containerList {
		containerInspected, err := dockerClient.ContainerInspect(ctx, container.ID)
		if err != nil {
			log.Warnf("Failed to inspect container %s, error: %s", container.ID, err)
		} else {
			containersInspected = append(containersInspected, containerInspected)
		}
	}
	return containersInspected, nil
}

// Escape beginning slash "/", convert all others to dash "-"
func (provider *Docker) getSubDomain(name string) string {
	return strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1)
}
