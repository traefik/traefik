package docker

import (
	"context"
	"errors"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	"github.com/docker/engine-api/client"
	dockertypes "github.com/docker/engine-api/types"
	dockercontainertypes "github.com/docker/engine-api/types/container"
	eventtypes "github.com/docker/engine-api/types/events"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/swarm"
	swarmtypes "github.com/docker/engine-api/types/swarm"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-connections/sockets"
	"github.com/vdemeester/docker-events"
)

const (
	// SwarmAPIVersion is a constant holding the version of the Provider API traefik will use
	SwarmAPIVersion string = "1.24"
	// SwarmDefaultWatchTime is the duration of the interval when polling docker
	SwarmDefaultWatchTime = 15 * time.Second
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the Provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	Endpoint              string              `description:"Provider server endpoint. Can be a tcp or a unix socket endpoint"`
	Domain                string              `description:"Default domain used"`
	TLS                   *provider.ClientTLS `description:"Enable Provider TLS support"`
	ExposedByDefault      bool                `description:"Expose containers by default"`
	UseBindPortIP         bool                `description:"Use the ip address from the bound port, rather than from the inner network"`
	SwarmMode             bool                `description:"Use Provider on Swarm Mode"`
}

// dockerData holds the need data to the Provider p
type dockerData struct {
	ServiceName     string
	Name            string
	Labels          map[string]string // List of labels set to container or service
	NetworkSettings networkSettings
	Health          string
}

// NetworkSettings holds the networks data to the Provider p
type networkSettings struct {
	NetworkMode dockercontainertypes.NetworkMode
	Ports       nat.PortMap
	Networks    map[string]*networkData
}

// Network holds the network data to the Provider p
type networkData struct {
	Name     string
	Addr     string
	Port     int
	Protocol string
	ID       string
}

func (p *Provider) createClient() (client.APIClient, error) {
	var httpClient *http.Client
	httpHeaders := map[string]string{
		"User-Agent": "Traefik " + version.Version,
	}
	if p.TLS != nil {
		config, err := p.TLS.CreateTLSConfig()
		if err != nil {
			return nil, err
		}
		tr := &http.Transport{
			TLSClientConfig: config,
		}
		proto, addr, _, err := client.ParseHost(p.Endpoint)
		if err != nil {
			return nil, err
		}

		sockets.ConfigureTransport(tr, proto, addr)

		httpClient = &http.Client{
			Transport: tr,
		}

	}
	var version string
	if p.SwarmMode {
		version = SwarmAPIVersion
	} else {
		version = DockerAPIVersion
	}
	return client.NewClient(p.Endpoint, version, httpClient, httpHeaders)

}

// Provide allows the p to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	p.Constraints = append(p.Constraints, constraints...)
	// TODO register this routine in pool, and watch for stop channel
	safe.Go(func() {
		operation := func() error {
			var err error

			dockerClient, err := p.createClient()
			if err != nil {
				log.Errorf("Failed to create a client for docker, error: %s", err)
				return err
			}

			ctx := context.Background()
			version, err := dockerClient.ServerVersion(ctx)
			log.Debugf("Provider connection established with docker %s (API %s)", version.Version, version.APIVersion)
			var dockerDataList []dockerData
			if p.SwarmMode {
				dockerDataList, err = p.listServices(ctx, dockerClient)
				if err != nil {
					log.Errorf("Failed to list services for docker swarm mode, error %s", err)
					return err
				}
			} else {
				dockerDataList, err = listContainers(ctx, dockerClient)
				if err != nil {
					log.Errorf("Failed to list containers for docker, error %s", err)
					return err
				}
			}

			configuration := p.loadDockerConfig(dockerDataList)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "docker",
				Configuration: configuration,
			}
			if p.Watch {
				ctx, cancel := context.WithCancel(ctx)
				if p.SwarmMode {
					// TODO: This need to be change. Linked to Swarm events docker/docker#23827
					ticker := time.NewTicker(SwarmDefaultWatchTime)
					pool.Go(func(stop chan bool) {
						for {
							select {
							case <-ticker.C:
								services, err := p.listServices(ctx, dockerClient)
								if err != nil {
									log.Errorf("Failed to list services for docker, error %s", err)
									return
								}
								configuration := p.loadDockerConfig(services)
								if configuration != nil {
									configurationChan <- types.ConfigMessage{
										ProviderName:  "docker",
										Configuration: configuration,
									}
								}

							case <-stop:
								ticker.Stop()
								cancel()
								return
							}
						}
					})

				} else {
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
						log.Debugf("Provider event received %+v", m)
						containers, err := listContainers(ctx, dockerClient)
						if err != nil {
							log.Errorf("Failed to list containers for docker, error %s", err)
							// Call cancel to get out of the monitor
							cancel()
							return
						}
						configuration := p.loadDockerConfig(containers)
						if configuration != nil {
							configurationChan <- types.ConfigMessage{
								ProviderName:  "docker",
								Configuration: configuration,
							}
						}
					}
					eventHandler.Handle("start", startStopHandle)
					eventHandler.Handle("die", startStopHandle)
					eventHandler.Handle("health_status: healthy", startStopHandle)
					eventHandler.Handle("health_status: unhealthy", startStopHandle)
					eventHandler.Handle("health_status: starting", startStopHandle)

					errChan := events.MonitorWithHandler(ctx, dockerClient, options, eventHandler)
					if err := <-errChan; err != nil {
						return err
					}
				}
			}
			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to docker server %+v", err)
		}
	})

	return nil
}

func (p *Provider) loadDockerConfig(containersInspected []dockerData) *types.Configuration {
	var DockerFuncMap = template.FuncMap{
		"getBackend":                  p.getBackend,
		"getIPAddress":                p.getIPAddress,
		"getPort":                     p.getPort,
		"getWeight":                   p.getWeight,
		"getDomain":                   p.getDomain,
		"getProtocol":                 p.getProtocol,
		"getPassHostHeader":           p.getPassHostHeader,
		"getPriority":                 p.getPriority,
		"getEntryPoints":              p.getEntryPoints,
		"getFrontendRule":             p.getFrontendRule,
		"hasCircuitBreakerLabel":      p.hasCircuitBreakerLabel,
		"getCircuitBreakerExpression": p.getCircuitBreakerExpression,
		"hasLoadBalancerLabel":        p.hasLoadBalancerLabel,
		"getLoadBalancerMethod":       p.getLoadBalancerMethod,
		"hasMaxConnLabels":            p.hasMaxConnLabels,
		"getMaxConnAmount":            p.getMaxConnAmount,
		"getMaxConnExtractorFunc":     p.getMaxConnExtractorFunc,
		"getSticky":                   p.getSticky,
		"getIsBackendLBSwarm":         p.getIsBackendLBSwarm,
		"hasServices":                 p.hasServices,
		"getServiceNames":             p.getServiceNames,
		"getServicePort":              p.getServicePort,
		"getServiceWeight":            p.getServiceWeight,
		"getServiceProtocol":          p.getServiceProtocol,
		"getServiceEntryPoints":       p.getServiceEntryPoints,
		"getServiceFrontendRule":      p.getServiceFrontendRule,
		"getServicePassHostHeader":    p.getServicePassHostHeader,
		"getServicePriority":          p.getServicePriority,
		"getServiceBackend":           p.getServiceBackend,
	}
	// filter containers
	filteredContainers := fun.Filter(func(container dockerData) bool {
		return p.containerFilter(container)
	}, containersInspected).([]dockerData)

	frontends := map[string][]dockerData{}
	backends := map[string]dockerData{}
	servers := map[string][]dockerData{}
	for _, container := range filteredContainers {
		frontendName := p.getFrontendName(container)
		frontends[frontendName] = append(frontends[frontendName], container)
		backendName := p.getBackend(container)
		backends[backendName] = container
		servers[backendName] = append(servers[backendName], container)
	}

	templateObjects := struct {
		Containers []dockerData
		Frontends  map[string][]dockerData
		Backends   map[string]dockerData
		Servers    map[string][]dockerData
		Domain     string
	}{
		filteredContainers,
		frontends,
		backends,
		servers,
		p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/docker.tmpl", DockerFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func (p *Provider) hasCircuitBreakerLabel(container dockerData) bool {
	if _, err := getLabel(container, "traefik.backend.circuitbreaker.expression"); err != nil {
		return false
	}
	return true
}

// Regexp used to extract the name of the service and the name of the property for this service
// All properties are under the format traefik.<servicename>.frontent.*= except the port/weight/protocol directly after traefik.<servicename>.
var servicesPropertiesRegexp = regexp.MustCompile(`^traefik\.(?P<service_name>.*?)\.(?P<property_name>port|weight|protocol|frontend\.(.*))$`)

// Map of services properties
// we can get it with label[serviceName][propertyName] and we got the propertyValue
type labelServiceProperties map[string]map[string]string

// Check if for the given container, we find labels that are defining services
func (p *Provider) hasServices(container dockerData) bool {
	return len(extractServicesLabels(container.Labels)) > 0
}

// Extract the service labels from container labels of dockerData struct
func extractServicesLabels(labels map[string]string) labelServiceProperties {
	v := make(labelServiceProperties)

	for index, serviceProperty := range labels {
		matches := servicesPropertiesRegexp.FindStringSubmatch(index)
		if matches != nil {
			result := make(map[string]string)
			for i, name := range servicesPropertiesRegexp.SubexpNames() {
				if i != 0 {
					result[name] = matches[i]
				}
			}
			serviceName := result["service_name"]
			if _, ok := v[serviceName]; !ok {
				v[serviceName] = make(map[string]string)
			}
			v[serviceName][result["property_name"]] = serviceProperty
		}
	}

	return v
}

// Gets the entry for a service label searching in all labels of the given container
func getContainerServiceLabel(container dockerData, serviceName string, entry string) (string, bool) {
	value, ok := extractServicesLabels(container.Labels)[serviceName][entry]
	return value, ok
}

// Gets array of service names for a given container
func (p *Provider) getServiceNames(container dockerData) []string {
	labelServiceProperties := extractServicesLabels(container.Labels)
	keys := make([]string, 0, len(labelServiceProperties))
	for k := range labelServiceProperties {
		keys = append(keys, k)
	}
	return keys
}

// Extract entrypoints from labels for a given service and a given docker container
func (p *Provider) getServiceEntryPoints(container dockerData, serviceName string) []string {
	if entryPoints, ok := getContainerServiceLabel(container, serviceName, "frontend.entryPoints"); ok {
		return strings.Split(entryPoints, ",")
	}
	return p.getEntryPoints(container)

}

// Extract passHostHeader from labels for a given service and a given docker container
func (p *Provider) getServicePassHostHeader(container dockerData, serviceName string) string {
	if servicePassHostHeader, ok := getContainerServiceLabel(container, serviceName, "frontend.passHostHeader"); ok {
		return servicePassHostHeader
	}
	return p.getPassHostHeader(container)
}

// Extract priority from labels for a given service and a given docker container
func (p *Provider) getServicePriority(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, "frontend.priority"); ok {
		return value
	}
	return p.getPriority(container)

}

// Extract backend from labels for a given service and a given docker container
func (p *Provider) getServiceBackend(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, "frontend.backend"); ok {
		return value
	}
	return p.getBackend(container) + "-" + provider.Normalize(serviceName)
}

// Extract rule from labels for a given service and a given docker container
func (p *Provider) getServiceFrontendRule(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, "frontend.rule"); ok {
		return value
	}
	return p.getFrontendRule(container)

}

// Extract port from labels for a given service and a given docker container
func (p *Provider) getServicePort(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, "port"); ok {
		return value
	}
	return p.getPort(container)
}

// Extract weight from labels for a given service and a given docker container
func (p *Provider) getServiceWeight(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, "weight"); ok {
		return value
	}
	return p.getWeight(container)
}

// Extract protocol from labels for a given service and a given docker container
func (p *Provider) getServiceProtocol(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, "protocol"); ok {
		return value
	}
	return p.getProtocol(container)
}

func (p *Provider) hasLoadBalancerLabel(container dockerData) bool {
	_, errMethod := getLabel(container, "traefik.backend.loadbalancer.method")
	_, errSticky := getLabel(container, "traefik.backend.loadbalancer.sticky")
	if errMethod != nil && errSticky != nil {
		return false
	}
	return true
}

func (p *Provider) hasMaxConnLabels(container dockerData) bool {
	if _, err := getLabel(container, "traefik.backend.maxconn.amount"); err != nil {
		return false
	}
	if _, err := getLabel(container, "traefik.backend.maxconn.extractorfunc"); err != nil {
		return false
	}
	return true
}

func (p *Provider) getCircuitBreakerExpression(container dockerData) string {
	if label, err := getLabel(container, "traefik.backend.circuitbreaker.expression"); err == nil {
		return label
	}
	return "NetworkErrorRatio() > 1"
}

func (p *Provider) getLoadBalancerMethod(container dockerData) string {
	if label, err := getLabel(container, "traefik.backend.loadbalancer.method"); err == nil {
		return label
	}
	return "wrr"
}

func (p *Provider) getMaxConnAmount(container dockerData) int64 {
	if label, err := getLabel(container, "traefik.backend.maxconn.amount"); err == nil {
		i, errConv := strconv.ParseInt(label, 10, 64)
		if errConv != nil {
			log.Errorf("Unable to parse traefik.backend.maxconn.amount %s", label)
			return math.MaxInt64
		}
		return i
	}
	return math.MaxInt64
}

func (p *Provider) getMaxConnExtractorFunc(container dockerData) string {
	if label, err := getLabel(container, "traefik.backend.maxconn.extractorfunc"); err == nil {
		return label
	}
	return "request.host"
}

func (p *Provider) containerFilter(container dockerData) bool {
	_, err := strconv.Atoi(container.Labels["traefik.port"])
	if len(container.NetworkSettings.Ports) == 0 && err != nil {
		log.Debugf("Filtering container without port and no traefik.port label %s", container.Name)
		return false
	}

	if !isContainerEnabled(container, p.ExposedByDefault) {
		log.Debugf("Filtering disabled container %s", container.Name)
		return false
	}

	constraintTags := strings.Split(container.Labels["traefik.tags"], ",")
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Container %v pruned by '%v' constraint", container.Name, failingConstraint.String())
		}
		return false
	}

	if container.Health != "" && container.Health != "healthy" {
		log.Debugf("Filtering unhealthy or starting container %s", container.Name)
		return false
	}

	return true
}

func (p *Provider) getFrontendName(container dockerData) string {
	// Replace '.' with '-' in quoted keys because of this issue https://github.com/BurntSushi/toml/issues/78
	return provider.Normalize(p.getFrontendRule(container))
}

// GetFrontendRule returns the frontend rule for the specified container, using
// it's label. It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(container dockerData) string {
	if label, err := getLabel(container, "traefik.frontend.rule"); err == nil {
		return label
	}
	if labels, err := getLabels(container, []string{"com.docker.compose.project", "com.docker.compose.service"}); err == nil {
		return "Host:" + p.getSubDomain(labels["com.docker.compose.service"]+"."+labels["com.docker.compose.project"]) + "." + p.Domain
	}

	return "Host:" + p.getSubDomain(container.ServiceName) + "." + p.Domain
}

func (p *Provider) getBackend(container dockerData) string {
	if label, err := getLabel(container, "traefik.backend"); err == nil {
		return provider.Normalize(label)
	}
	if labels, err := getLabels(container, []string{"com.docker.compose.project", "com.docker.compose.service"}); err == nil {
		return provider.Normalize(labels["com.docker.compose.service"] + "_" + labels["com.docker.compose.project"])
	}
	return provider.Normalize(container.ServiceName)
}

func (p *Provider) getIPAddress(container dockerData) string {
	if label, err := getLabel(container, "traefik.docker.network"); err == nil && label != "" {
		networkSettings := container.NetworkSettings
		if networkSettings.Networks != nil {
			network := networkSettings.Networks[label]
			if network != nil {
				return network.Addr
			}

			log.Warnf("Could not find network named '%s' for container '%s'! Maybe you're missing the project's prefix in the label? Defaulting to first available network.", label, container.Name)
		}
	}

	// If net==host, quick n' dirty, we return 127.0.0.1
	// This will work locally, but will fail with swarm.
	if "host" == container.NetworkSettings.NetworkMode {
		return "127.0.0.1"
	}

	if p.UseBindPortIP {
		port := p.getPort(container)
		for netport, portBindings := range container.NetworkSettings.Ports {
			if string(netport) == port+"/TCP" || string(netport) == port+"/UDP" {
				for _, p := range portBindings {
					return p.HostIP
				}
			}
		}
	}

	for _, network := range container.NetworkSettings.Networks {
		return network.Addr
	}
	return ""
}

func (p *Provider) getPort(container dockerData) string {
	if label, err := getLabel(container, "traefik.port"); err == nil {
		return label
	}
	for key := range container.NetworkSettings.Ports {
		return key.Port()
	}
	return ""
}

func (p *Provider) getWeight(container dockerData) string {
	if label, err := getLabel(container, "traefik.weight"); err == nil {
		return label
	}
	return "0"
}

func (p *Provider) getSticky(container dockerData) string {
	if label, err := getLabel(container, "traefik.backend.loadbalancer.sticky"); err == nil {
		return label
	}
	return "false"
}

func (p *Provider) getIsBackendLBSwarm(container dockerData) string {
	if label, err := getLabel(container, "traefik.backend.loadbalancer.swarm"); err == nil {
		return label
	}
	return "false"
}

func (p *Provider) getDomain(container dockerData) string {
	if label, err := getLabel(container, "traefik.domain"); err == nil {
		return label
	}
	return p.Domain
}

func (p *Provider) getProtocol(container dockerData) string {
	if label, err := getLabel(container, "traefik.protocol"); err == nil {
		return label
	}
	return "http"
}

func (p *Provider) getPassHostHeader(container dockerData) string {
	if passHostHeader, err := getLabel(container, "traefik.frontend.passHostHeader"); err == nil {
		return passHostHeader
	}
	return "true"
}

func (p *Provider) getPriority(container dockerData) string {
	if priority, err := getLabel(container, "traefik.frontend.priority"); err == nil {
		return priority
	}
	return "0"
}

func (p *Provider) getEntryPoints(container dockerData) []string {
	if entryPoints, err := getLabel(container, "traefik.frontend.entryPoints"); err == nil {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

func isContainerEnabled(container dockerData, exposedByDefault bool) bool {
	return exposedByDefault && container.Labels["traefik.enable"] != "false" || container.Labels["traefik.enable"] == "true"
}

func getLabel(container dockerData, label string) (string, error) {
	for key, value := range container.Labels {
		if key == label {
			return value, nil
		}
	}
	return "", errors.New("Label not found:" + label)
}

func getLabels(container dockerData, labels []string) (map[string]string, error) {
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

func listContainers(ctx context.Context, dockerClient client.ContainerAPIClient) ([]dockerData, error) {
	containerList, err := dockerClient.ContainerList(ctx, dockertypes.ContainerListOptions{})
	if err != nil {
		return []dockerData{}, err
	}
	containersInspected := []dockerData{}

	// get inspect containers
	for _, container := range containerList {
		containerInspected, err := dockerClient.ContainerInspect(ctx, container.ID)
		if err != nil {
			log.Warnf("Failed to inspect container %s, error: %s", container.ID, err)
		} else {
			dockerData := parseContainer(containerInspected)
			containersInspected = append(containersInspected, dockerData)
		}
	}
	return containersInspected, nil
}

func parseContainer(container dockertypes.ContainerJSON) dockerData {
	dockerData := dockerData{
		NetworkSettings: networkSettings{},
	}

	if container.ContainerJSONBase != nil {
		dockerData.Name = container.ContainerJSONBase.Name
		dockerData.ServiceName = dockerData.Name //Default ServiceName to be the container's Name.

		if container.ContainerJSONBase.HostConfig != nil {
			dockerData.NetworkSettings.NetworkMode = container.ContainerJSONBase.HostConfig.NetworkMode
		}

		if container.State != nil && container.State.Health != nil {
			dockerData.Health = container.State.Health.Status
		}
	}

	if container.Config != nil && container.Config.Labels != nil {
		dockerData.Labels = container.Config.Labels
	}

	if container.NetworkSettings != nil {
		if container.NetworkSettings.Ports != nil {
			dockerData.NetworkSettings.Ports = container.NetworkSettings.Ports
		}
		if container.NetworkSettings.Networks != nil {
			dockerData.NetworkSettings.Networks = make(map[string]*networkData)
			for name, containerNetwork := range container.NetworkSettings.Networks {
				dockerData.NetworkSettings.Networks[name] = &networkData{
					ID:   containerNetwork.NetworkID,
					Name: name,
					Addr: containerNetwork.IPAddress,
				}
			}
		}

	}

	return dockerData
}

// Escape beginning slash "/", convert all others to dash "-", and convert underscores "_" to dash "-"
func (p *Provider) getSubDomain(name string) string {
	return strings.Replace(strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1), "_", "-", -1)
}

func (p *Provider) listServices(ctx context.Context, dockerClient client.APIClient) ([]dockerData, error) {
	serviceList, err := dockerClient.ServiceList(ctx, dockertypes.ServiceListOptions{})
	if err != nil {
		return []dockerData{}, err
	}
	networkListArgs := filters.NewArgs()
	networkListArgs.Add("driver", "overlay")

	networkList, err := dockerClient.NetworkList(ctx, dockertypes.NetworkListOptions{Filters: networkListArgs})

	networkMap := make(map[string]*dockertypes.NetworkResource)
	if err != nil {
		log.Debug("Failed to network inspect on client for docker, error: %s", err)
		return []dockerData{}, err
	}
	for _, network := range networkList {
		networkToAdd := network
		networkMap[network.ID] = &networkToAdd
	}

	var dockerDataList []dockerData
	var dockerDataListTasks []dockerData

	for _, service := range serviceList {
		dockerData := parseService(service, networkMap)
		useSwarmLB, _ := strconv.ParseBool(p.getIsBackendLBSwarm(dockerData))
		isGlobalSvc := service.Spec.Mode.Global != nil

		if useSwarmLB {
			dockerDataList = append(dockerDataList, dockerData)
		} else {
			dockerDataListTasks, err = listTasks(ctx, dockerClient, service.ID, dockerData, networkMap, isGlobalSvc)

			for _, dockerDataTask := range dockerDataListTasks {
				dockerDataList = append(dockerDataList, dockerDataTask)
			}
		}
	}
	return dockerDataList, err

}

func parseService(service swarmtypes.Service, networkMap map[string]*dockertypes.NetworkResource) dockerData {
	dockerData := dockerData{
		ServiceName:     service.Spec.Annotations.Name,
		Name:            service.Spec.Annotations.Name,
		Labels:          service.Spec.Annotations.Labels,
		NetworkSettings: networkSettings{},
	}

	if service.Spec.EndpointSpec != nil {
		switch service.Spec.EndpointSpec.Mode {
		case swarm.ResolutionModeDNSRR:
			log.Debug("Ignored endpoint-mode not supported, service name: %s", dockerData.Name)
		case swarm.ResolutionModeVIP:
			dockerData.NetworkSettings.Networks = make(map[string]*networkData)
			for _, virtualIP := range service.Endpoint.VirtualIPs {
				networkService := networkMap[virtualIP.NetworkID]
				if networkService != nil {
					ip, _, _ := net.ParseCIDR(virtualIP.Addr)
					network := &networkData{
						Name: networkService.Name,
						ID:   virtualIP.NetworkID,
						Addr: ip.String(),
					}
					dockerData.NetworkSettings.Networks[network.Name] = network
				} else {
					log.Debug("Network not found, id: %s", virtualIP.NetworkID)
				}
			}
		}
	}
	return dockerData
}

func listTasks(ctx context.Context, dockerClient client.APIClient, serviceID string,
	serviceDockerData dockerData, networkMap map[string]*dockertypes.NetworkResource, isGlobalSvc bool) ([]dockerData, error) {
	serviceIDFilter := filters.NewArgs()
	serviceIDFilter.Add("service", serviceID)
	serviceIDFilter.Add("desired-state", "running")
	taskList, err := dockerClient.TaskList(ctx, dockertypes.TaskListOptions{Filter: serviceIDFilter})

	if err != nil {
		return []dockerData{}, err
	}
	var dockerDataList []dockerData

	for _, task := range taskList {
		if task.Status.State != swarm.TaskStateRunning {
			continue
		}
		dockerData := parseTasks(task, serviceDockerData, networkMap, isGlobalSvc)
		dockerDataList = append(dockerDataList, dockerData)
	}
	return dockerDataList, err
}

func parseTasks(task swarmtypes.Task, serviceDockerData dockerData, networkMap map[string]*dockertypes.NetworkResource, isGlobalSvc bool) dockerData {
	dockerData := dockerData{
		ServiceName:     serviceDockerData.Name,
		Name:            serviceDockerData.Name + "." + strconv.Itoa(task.Slot),
		Labels:          serviceDockerData.Labels,
		NetworkSettings: networkSettings{},
	}

	if isGlobalSvc == true {
		dockerData.Name = serviceDockerData.Name + "." + task.ID
	}

	if task.NetworksAttachments != nil {
		dockerData.NetworkSettings.Networks = make(map[string]*networkData)
		for _, virtualIP := range task.NetworksAttachments {
			if networkService, present := networkMap[virtualIP.Network.ID]; present {
				// Not sure about this next loop - when would a task have multiple IP's for the same network?
				for _, addr := range virtualIP.Addresses {
					ip, _, _ := net.ParseCIDR(addr)
					network := &networkData{
						ID:   virtualIP.Network.ID,
						Name: networkService.Name,
						Addr: ip.String(),
					}
					dockerData.NetworkSettings.Networks[network.Name] = network
				}
			}
		}
	}
	return dockerData
}
