package docker

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	dockertypes "github.com/docker/docker/api/types"
	dockercontainertypes "github.com/docker/docker/api/types/container"
	eventtypes "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-connections/sockets"
)

const (
	// SwarmAPIVersion is a constant holding the version of the Provider API traefik will use
	SwarmAPIVersion = "1.24"

	// SwarmDefaultWatchTime is the duration of the interval when polling docker
	SwarmDefaultWatchTime = 15 * time.Second
)

type listener interface {
	Listen(context.Context, chan<- types.ConfigMessage, func(context.Context, eventtypes.Message, chan<- types.ConfigMessage)) error
}

func newTickerListener() *tickerListener {
	return &tickerListener{
		ticker: time.NewTicker(SwarmDefaultWatchTime),
	}
}

type tickerListener struct {
	ticker *time.Ticker
}

// Listen creates a ticker that polls periodically from the Docker Swarm daemon.
// Refreshes the service backend list on every poll.
// Default poll timer value is defined by the SwarmDefaultWatchTime constant.
func (t *tickerListener) Listen(ctx context.Context, configurationChan chan<- types.ConfigMessage, callbackFunc func(context.Context, eventtypes.Message, chan<- types.ConfigMessage)) error {
	log.Debug("Docker events listener: Starting up the Ticker listener...")

	for {
		select {
		case <-t.ticker.C:
			go callbackFunc(ctx, eventtypes.Message{}, configurationChan)
		case <-ctx.Done():
			return nil
		}
	}
}

func newStreamerListener(dockerClient client.APIClient) *streamerListener {
	return &streamerListener{
		dockerClient: dockerClient,
	}
}

type streamerListener struct {
	dockerClient client.APIClient
}

// Listen creates a live event streamer with the Docker Swarm daemon.
// Refreshes the service backend list on every service swarm event.
func (s *streamerListener) Listen(ctx context.Context, configurationChan chan<- types.ConfigMessage, callbackFunc func(context.Context, eventtypes.Message, chan<- types.ConfigMessage)) error {
	log.Debug("Docker events listener: Starting up the Streamer listener...")

	eventsMsgChan, eventsErrChan := s.dockerClient.Events(
		ctx,
		dockertypes.EventsOptions{
			Filters: filters.NewArgs(
				filters.Arg("scope", "swarm"),
				filters.Arg("type", "service"),
			),
		},
	)

	for {
		select {
		case evt := <-eventsMsgChan:
			log.Debugf("Docker events Streamer listener: Incoming event: %#v", evt)

			go callbackFunc(ctx, evt, configurationChan)
		case evtErr := <-eventsErrChan:
			log.Errorf("Docker events listener: Events error, %s", evtErr.Error())

			return evtErr
		case <-ctx.Done():
			return nil
		}
	}
}

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`
	Endpoint              string           `description:"Docker server endpoint. Can be a tcp or a unix socket endpoint"`
	Domain                string           `description:"Default domain used"`
	TLS                   *types.ClientTLS `description:"Enable Docker TLS support" export:"true"`
	ExposedByDefault      bool             `description:"Expose containers by default" export:"true"`
	UseBindPortIP         bool             `description:"Use the ip address from the bound port, rather than from the inner network" export:"true"`
	SwarmMode             bool             `description:"Use Docker on Swarm Mode" export:"true"`
	EventHandlers         []Event          `description:"Event handlers with callback support" export:"true"`
	Network               string           `description:"Default Docker network used" export:"true"`
	dockerClient          client.APIClient
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	return p.BaseProvider.Init(constraints)
}

// dockerData holds the need data to the Provider p
type dockerData struct {
	ServiceName     string
	Name            string
	Labels          map[string]string // List of labels set to container or service
	NetworkSettings networkSettings
	Health          string
	Node            *dockertypes.ContainerNode
	SegmentLabels   map[string]string
	SegmentName     string
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
	if p.dockerClient != nil {
		return p.dockerClient, nil
	}

	var httpClient *http.Client

	if p.TLS != nil {
		config, err := p.TLS.CreateTLSConfig()
		if err != nil {
			return nil, err
		}
		tr := &http.Transport{
			TLSClientConfig: config,
		}

		hostURL, err := client.ParseHostURL(p.Endpoint)
		if err != nil {
			return nil, err
		}
		sockets.ConfigureTransport(tr, hostURL.Scheme, hostURL.Host)

		httpClient = &http.Client{
			Transport: tr,
		}
	}

	httpHeaders := map[string]string{
		"User-Agent": "Traefik " + version.Version,
	}

	var apiVersion string
	if p.SwarmMode {
		apiVersion = SwarmAPIVersion
	} else {
		apiVersion = DockerAPIVersion
	}

	var err error
	p.dockerClient, err = client.NewClient(p.Endpoint, apiVersion, httpClient, httpHeaders)

	return p.dockerClient, err
}

// Provide allows the docker provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
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
			serverVersion, err := dockerClient.ServerVersion(ctx)
			if err != nil {
				log.Errorf("Failed to retrieve information of the docker client and server host: %s", err)
				return err
			}
			log.Debugf("Provider connection established with docker %s (API %s)", serverVersion.Version, serverVersion.APIVersion)
			var dockerDataList []dockerData
			if p.SwarmMode {
				dockerDataList, err = listServices(ctx, dockerClient)
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

			configuration := p.buildConfiguration(dockerDataList)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "docker",
				Configuration: configuration,
			}
			if p.Watch {
				if p.SwarmMode {
					errChan := make(chan error)
					pool.Go(func(stop chan bool) {
						defer close(errChan)

						watchCtx, cancel := context.WithCancel(ctx)
						defer cancel()

						swarmEventsCap, err := swarmEventsCapabilities(watchCtx, dockerClient)
						if err != nil {
							log.Errorf("Unable to retrieve Docker Swarm event listener capabilities, error %s", err.Error())

							errChan <- err
							return
						}

						var l listener
						if swarmEventsCap {
							l = newStreamerListener(dockerClient)
						} else {
							l = newTickerListener()
						}

						if err := l.Listen(watchCtx, configurationChan, p.eventCallback); err != nil {
							log.Errorf("Error while listening/polling Docker Swarm events, error %s", err.Error())

							errChan <- err
							return
						}
					})

					if err, ok := <-errChan; ok {
						return err
					}
					// channel closed
				} else {
					watchCtx, cancel := context.WithCancel(ctx)
					defer cancel()

					f := filters.NewArgs()
					f.Add("type", "container")
					options := dockertypes.EventsOptions{
						Filters: f,
					}

					startStopHandle := func(m eventtypes.Message) {
						log.Debugf("Provider event received %+v", m)
						containers, err := listContainers(watchCtx, dockerClient)
						if err != nil {
							log.Errorf("Failed to list containers for docker, error %s", err)
							// Call cancel to get out of the monitor
							cancel()
							return
						}
						configuration := p.buildConfiguration(containers)
						if configuration != nil {
							configurationChan <- types.ConfigMessage{
								ProviderName:  "docker",
								Configuration: configuration,
							}
						}
					}

					eventsc, errc := dockerClient.Events(watchCtx, options)
					for {
						select {
						case event := <-eventsc:
							if event.Action == "start" ||
								event.Action == "die" ||
								strings.HasPrefix(event.Action, "health_status") {
								startStopHandle(event)
							}
						case err := <-errc:
							if err == io.EOF {
								log.Debug("Provider event stream closed")
							}

							return err
						}
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

func (p *Provider) listAndUpdateServices(ctx context.Context, dockerClient client.APIClient, configurationChan chan<- types.ConfigMessage) error {
	log.Debug("listAndUpdateServices called!")
	services, err := listServices(ctx, dockerClient)
	if err != nil {
		return err
	}
	log.Debugf("Services found! %#v", services)

	configuration := p.buildConfiguration(services)
	log.Debugf("Configuration built: %#v", configuration)
	if configuration != nil {
		configurationChan <- types.ConfigMessage{
			ProviderName:  "docker",
			Configuration: configuration,
		}
	}

	return nil
}

func (p *Provider) eventCallback(ctx context.Context, msg eventtypes.Message, configurationChan chan<- types.ConfigMessage) {
	dockerClient, err := p.createClient()
	if err != nil {
		return
	}

	log.Debugf("Docker event callback function executed with payload: %#v", msg)

	if msg.Actor.ID == "" {
		if err := p.listAndUpdateServices(ctx, dockerClient, configurationChan); err != nil {
			log.Error(err.Error())
		}

		return
	}

	service, err := getService(ctx, dockerClient, msg.Actor.ID)
	if err != nil {
		// TODO: How should we treat this kind of an error?
		log.Error(err.Error())

		return
	}

	if service.Spec.Mode.Global == nil && service.Spec.Mode.Replicated == nil {
		log.Error("Service has no specified mode! This should never happen...")

		return
	}

	if service.Spec.Mode.Global != nil {
		// TODO: For now we just sleep for 1 second if it's a global service,
		// since we don't know how many tasks should be expected (unless we list the nodes, etc. - but there might also be a race condition).
		log.Info("Service is in global mode. Sleep for 1 second and cross our fingers we won't run into any race conditions...")
		time.Sleep(1 * time.Second)
	}

	taskList, err := listTasks(ctx, dockerClient, service.ID, "running")
	if err != nil {
		// TODO: How should we treat this kind of an error?
		log.Error(err.Error())

		return
	}

	retry := false
	// Retry if there are no tasks found.
	if len(taskList) == 0 {
		log.Infof("No tasks for service %s found! Retrying...", service.ID)

		retry = true
	}

	if service.Spec.Mode.Replicated != nil {
		if service.Spec.Mode.Replicated.Replicas == nil {
			log.Error("Service is in replicated mode, but no replicas are defined! This should never happen...")

			return
		}

		// Retry if the service is in replicated mode and the list of tasks is shorter than the number of replicas.
		// We don't need to retry if the number of services is longer than the number of replicas. That means the service is being scaled in.
		numberOfReplicas := int(*service.Spec.Mode.Replicated.Replicas)
		if len(taskList) < numberOfReplicas {
			log.Infof("Task list length for service %s is %d, expected it to be %d. Retrying...", service.ID, len(taskList), numberOfReplicas)

			retry = true
		}
	}

TaskLoop:
	for _, task := range taskList {
		log.Debugf("State of task %s: %s", task.ID, task.Status.State)

		if task.Status.State != swarmtypes.TaskStateRunning {
			switch task.Status.State {
			case
				swarmtypes.TaskStateNew,
				swarmtypes.TaskStatePending,
				swarmtypes.TaskStateAssigned,
				swarmtypes.TaskStateAccepted,
				swarmtypes.TaskStatePreparing,
				swarmtypes.TaskStateStarting:
				retry = true

				break TaskLoop
			}
		}
	}

	if !retry {
		log.Debug("Event callback: Updating routing configuration if needed...")

		p.listAndUpdateServices(ctx, dockerClient, configurationChan)
	} else {
		// We should only reach this place when new tasks are being created.
		// Therefore, sleeping here shouldn't affect the graceful scale down.
		log.Debug("Callback task state check: Retrying in 1 second...")

		// Sleep for 1 second between retries.
		time.Sleep(1 * time.Second)

		log.Debug("Event callback: Retrying...")
		go p.eventCallback(ctx, msg, configurationChan)
	}
}

func listContainers(ctx context.Context, dockerClient client.ContainerAPIClient) ([]dockerData, error) {
	containerList, err := dockerClient.ContainerList(ctx, dockertypes.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	var containersInspected []dockerData
	// get inspect containers
	for _, container := range containerList {
		dData := inspectContainers(ctx, dockerClient, container.ID)
		if len(dData.Name) > 0 {
			containersInspected = append(containersInspected, dData)
		}
	}
	return containersInspected, nil
}

func inspectContainers(ctx context.Context, dockerClient client.ContainerAPIClient, containerID string) dockerData {
	dData := dockerData{}
	containerInspected, err := dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Warnf("Failed to inspect container %s, error: %s", containerID, err)
	} else {
		// This condition is here to avoid to have empty IP https://github.com/containous/traefik/issues/2459
		// We register only container which are running
		if containerInspected.ContainerJSONBase != nil && containerInspected.ContainerJSONBase.State != nil && containerInspected.ContainerJSONBase.State.Running {
			dData = parseContainer(containerInspected)
		}
	}
	return dData
}

func parseContainer(container dockertypes.ContainerJSON) dockerData {
	dData := dockerData{
		NetworkSettings: networkSettings{},
	}

	if container.ContainerJSONBase != nil {
		dData.Name = container.ContainerJSONBase.Name
		dData.ServiceName = dData.Name // Default ServiceName to be the container's Name.
		dData.Node = container.ContainerJSONBase.Node

		if container.ContainerJSONBase.HostConfig != nil {
			dData.NetworkSettings.NetworkMode = container.ContainerJSONBase.HostConfig.NetworkMode
		}

		if container.State != nil && container.State.Health != nil {
			dData.Health = container.State.Health.Status
		}
	}

	if container.Config != nil && container.Config.Labels != nil {
		dData.Labels = container.Config.Labels
	}

	if container.NetworkSettings != nil {
		if container.NetworkSettings.Ports != nil {
			dData.NetworkSettings.Ports = container.NetworkSettings.Ports
		}
		if container.NetworkSettings.Networks != nil {
			dData.NetworkSettings.Networks = make(map[string]*networkData)
			for name, containerNetwork := range container.NetworkSettings.Networks {
				dData.NetworkSettings.Networks[name] = &networkData{
					ID:   containerNetwork.NetworkID,
					Name: name,
					Addr: containerNetwork.IPAddress,
				}
			}
		}
	}
	return dData
}

func listServices(ctx context.Context, dockerClient client.APIClient) ([]dockerData, error) {
	serviceList, err := dockerClient.ServiceList(ctx, dockertypes.ServiceListOptions{})
	log.Debugf("Service list: %#v", serviceList)
	if err != nil {
		return nil, err
	}

	swarmEventsCap, err := swarmEventsCapabilities(ctx, dockerClient)
	if err != nil {
		return nil, err
	}

	networkListArgs := filters.NewArgs()
	if swarmEventsCap {
		networkListArgs.Add("scope", "swarm")
	} else {
		networkListArgs.Add("driver", "overlay")
	}

	networkList, err := dockerClient.NetworkList(ctx, dockertypes.NetworkListOptions{Filters: networkListArgs})
	if err != nil {
		log.Debugf("Failed to network inspect on client for docker, error: %s", err)
		return nil, err
	}

	networkMap := make(map[string]*dockertypes.NetworkResource)
	for _, network := range networkList {
		networkToAdd := network
		networkMap[network.ID] = &networkToAdd
	}

	var dockerDataList []dockerData
	var dockerDataListTasks []dockerData

	for _, service := range serviceList {
		dData := parseService(service, networkMap)

		if isBackendLBSwarm(dData) {
			if len(dData.NetworkSettings.Networks) > 0 {
				dockerDataList = append(dockerDataList, dData)
			} else {
				log.Warnf("No network found for service %s", service.Spec.Name)
			}
		} else {
			isGlobalSvc := service.Spec.Mode.Global != nil
			dockerDataListTasks, err = listAndParseTasks(ctx, dockerClient, service.ID, dData, networkMap, isGlobalSvc)
			if err != nil {
				log.Warnf("No tasks found for service %s, error %s", service.Spec.Name, err.Error())
			} else {
				log.Debugf("Tasks for service %s: %#v", service.Spec.Name, dockerDataListTasks)
				dockerDataList = append(dockerDataList, dockerDataListTasks...)
			}
		}
	}

	return dockerDataList, err
}

func parseService(service swarmtypes.Service, networkMap map[string]*dockertypes.NetworkResource) dockerData {
	dData := dockerData{
		ServiceName:     service.Spec.Annotations.Name,
		Name:            service.Spec.Annotations.Name,
		Labels:          service.Spec.Annotations.Labels,
		NetworkSettings: networkSettings{},
	}

	if service.Spec.EndpointSpec != nil {
		if service.Spec.EndpointSpec.Mode == swarmtypes.ResolutionModeDNSRR {
			if isBackendLBSwarm(dData) {
				log.Warnf("Ignored %s endpoint-mode not supported, service name: %s. Fallback to Træfik load balancing", swarmtypes.ResolutionModeDNSRR, service.Spec.Annotations.Name)
			}
		} else if service.Spec.EndpointSpec.Mode == swarmtypes.ResolutionModeVIP {
			dData.NetworkSettings.Networks = make(map[string]*networkData)
			for _, virtualIP := range service.Endpoint.VirtualIPs {
				networkService := networkMap[virtualIP.NetworkID]
				if networkService != nil {
					if len(virtualIP.Addr) > 0 {
						ip, _, _ := net.ParseCIDR(virtualIP.Addr)
						network := &networkData{
							Name: networkService.Name,
							ID:   virtualIP.NetworkID,
							Addr: ip.String(),
						}
						dData.NetworkSettings.Networks[network.Name] = network
					} else {
						log.Debugf("No virtual IPs found in network %s", virtualIP.NetworkID)
					}
				} else {
					log.Debugf("Network not found, id: %s", virtualIP.NetworkID)
				}
			}
		}
	}
	return dData
}

func listTasks(ctx context.Context, dockerClient client.APIClient, serviceID, desiredState string) ([]swarmtypes.Task, error) {
	return dockerClient.TaskList(
		ctx,
		dockertypes.TaskListOptions{
			Filters: filters.NewArgs(
				filters.Arg("service", serviceID),
				filters.Arg("desired-state", desiredState),
			),
		},
	)
}

func listAndParseTasks(ctx context.Context, dockerClient client.APIClient, serviceID string,
	serviceDockerData dockerData, networkMap map[string]*dockertypes.NetworkResource, isGlobalSvc bool) ([]dockerData, error) {
	taskList, err := listTasks(ctx, dockerClient, serviceID, "running")
	if err != nil {
		return nil, err
	}

	var dockerDataList []dockerData
	for _, task := range taskList {
		if task.Status.State != swarmtypes.TaskStateRunning {
			log.Warnf(
				"Task %s is not in the desired state (current state: %s, desired state: %s, service: %s)",
				task.ID,
				task.Status.State,
				swarmtypes.TaskStateRunning,
				serviceID,
			)

			continue
		}
		dData := parseTasks(task, serviceDockerData, networkMap, isGlobalSvc)
		if len(dData.NetworkSettings.Networks) > 0 {
			dockerDataList = append(dockerDataList, dData)
		} else {
			log.Warnf("No networks found for task %s (service: %s)", task.ID, serviceID)
		}
	}
	return dockerDataList, err
}

func parseTasks(task swarmtypes.Task, serviceDockerData dockerData,
	networkMap map[string]*dockertypes.NetworkResource, isGlobalSvc bool) dockerData {
	dData := dockerData{
		ServiceName:     serviceDockerData.Name,
		Name:            serviceDockerData.Name + "." + strconv.Itoa(task.Slot),
		Labels:          serviceDockerData.Labels,
		NetworkSettings: networkSettings{},
	}

	if isGlobalSvc {
		dData.Name = serviceDockerData.Name + "." + task.ID
	}

	if task.NetworksAttachments != nil {
		dData.NetworkSettings.Networks = make(map[string]*networkData)
		for _, virtualIP := range task.NetworksAttachments {
			if networkService, present := networkMap[virtualIP.Network.ID]; present {
				if len(virtualIP.Addresses) > 0 {
					// Not sure about this next loop - when would a task have multiple IP's for the same network?
					for _, addr := range virtualIP.Addresses {
						ip, _, _ := net.ParseCIDR(addr)
						network := &networkData{
							ID:   virtualIP.Network.ID,
							Name: networkService.Name,
							Addr: ip.String(),
						}
						dData.NetworkSettings.Networks[network.Name] = network
					}
				} else {
					log.Debugf("No IP addresses found for network %s", virtualIP.Network.ID)
				}
			}
		}
	}
	return dData
}

func getService(ctx context.Context, dockerClient client.APIClient, serviceID string) (swarmtypes.Service, error) {
	services, err := dockerClient.ServiceList(
		ctx,
		dockertypes.ServiceListOptions{
			Filters: filters.NewArgs(
				filters.Arg("id", serviceID),
			),
		},
	)
	if err != nil {
		return swarmtypes.Service{}, err
	}

	if len(services) != 1 {
		return swarmtypes.Service{}, fmt.Errorf("Failed to find service with id %s", serviceID)
	}

	return services[0], nil
}

func swarmEventsCapabilities(ctx context.Context, dockerClient client.APIClient) (bool, error) {
	res := false

	serverVersion, err := dockerClient.ServerVersion(ctx)
	if err != nil {
		return res, err
	}

	// https://docs.docker.com/engine/api/v1.29/#tag/Network (Docker 17.06)
	if versions.GreaterThanOrEqualTo(serverVersion.APIVersion, "1.29") {
		res = true
	}

	return res, nil
}
