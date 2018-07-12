package docker

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	listen(*safe.Pool, chan<- types.ConfigMessage, func(eventtypes.Message, chan<- types.ConfigMessage) error) error
}

func newTickerListener() *tickerListener {
	return &tickerListener{
		ticker: time.NewTicker(SwarmDefaultWatchTime),
	}
}

type tickerListener struct {
	ticker *time.Ticker
}

// listen creates a ticker that polls periodically from the Docker Swarm daemon.
// Refreshes the service backend list on every poll.
// Default poll timer value is defined by the SwarmDefaultWatchTime constant.
func (t *tickerListener) listen(pool *safe.Pool, configurationChan chan<- types.ConfigMessage, callbackFunc func(eventtypes.Message, chan<- types.ConfigMessage) error) error {
	log.Debug("Docker events listener: Starting up the Ticker listener...")
	for {
		select {
		case <-t.ticker.C:
			err := callbackFunc(eventtypes.Message{}, configurationChan)
			if err != nil {
				log.Errorf("Docker ticker listener: Callback error: %v", err)
			}
		case <-pool.Ctx().Done():
			pool.Cleanup()
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

// listen creates a live event streamer with the Docker Swarm daemon.
// Refreshes the service backend list on every service swarm event.
func (s *streamerListener) listen(pool *safe.Pool, configurationChan chan<- types.ConfigMessage, callbackFunc func(eventtypes.Message, chan<- types.ConfigMessage) error) error {
	log.Debug("Docker events listener: Starting up the Streamer listener...")

	eventsMsgChan, eventsErrChan := s.dockerClient.Events(
		pool.Ctx(),
		dockertypes.EventsOptions{
			Filters: filters.NewArgs(
				filters.Arg("scope", "swarm"),
				filters.Arg("type", eventtypes.ServiceEventType),
			),
		},
	)

	for {
		select {
		case evt := <-eventsMsgChan:
			err := callbackFunc(evt, configurationChan)
			if err != nil {
				log.Errorf("Docker events listener: Callback error: %v", err)
			}
		case evtErr := <-eventsErrChan:
			pool.Cleanup()
			if evtErr != nil {
				log.Errorf("Docker events listener: Events error: %v", evtErr)
			}
			return evtErr
		case <-pool.Ctx().Done():
			pool.Cleanup()
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
	Network               string           `description:"Default Docker network used" export:"true"`
	dockerClient          client.APIClient
	swarmPool             *safe.Pool
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

	pool.GoCtx(func(ctx context.Context) {
		operation := func() error {
			var err error
			p.dockerClient, err = p.createClient()
			if err != nil {
				log.Errorf("Failed to create a client for docker, error: %s", err)
				return err
			}

			serverVersion, err := p.dockerClient.ServerVersion(ctx)
			if err != nil {
				log.Errorf("Failed to retrieve information of the docker client and server host: %s", err)
				return err
			}

			log.Debugf("Provider connection established with docker %s (API %s)", serverVersion.Version, serverVersion.APIVersion)
			var dockerDataList []dockerData
			if p.SwarmMode {
				p.swarmPool = safe.NewPool(ctx)
				dockerDataList, err = listServices(ctx, p.dockerClient)
				if err != nil {
					log.Errorf("Failed to list services for docker swarm mode, error %s", err)
					return err
				}
			} else {
				dockerDataList, err = listContainers(ctx, p.dockerClient)
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
				errChan := make(chan error)
				defer close(errChan)
				pool.GoCtx(func(ctx context.Context) {
					if p.SwarmMode {
						swarmEventsCap, err := swarmEventsCapabilities(ctx, p.dockerClient)
						if err != nil {
							log.Errorf("Unable to retrieve Docker Swarm event listener capabilities: %v", err)
							errChan <- err
							return
						}

						var l listener
						if swarmEventsCap {
							l = newStreamerListener(p.dockerClient)
						} else {
							l = newTickerListener()
						}

						safe.Go(func() {
							if err := l.listen(p.swarmPool, configurationChan, p.eventCallback); err != nil {
								log.Errorf("Error while listening/polling Docker Swarm events: %v", err)
								errChan <- err
								return
							}
						})
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
							containers, err := listContainers(watchCtx, p.dockerClient)
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

						eventMessagesChan, eventErrChan := p.dockerClient.Events(watchCtx, options)
						for {
							select {
							case event := <-eventMessagesChan:
								if event.Action == "start" ||
									event.Action == "die" ||
									strings.HasPrefix(event.Action, "health_status") {
									startStopHandle(event)
								}
							case err := <-eventErrChan:
								if err == io.EOF {
									log.Debug("Provider event stream closed")
								}
								errChan <- err
							}
						}
					}
				})
				select {
				case err := <-errChan:
					return err
				case <-pool.Ctx().Done():
					return nil
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
	services, err := listServices(ctx, dockerClient)
	if err != nil {
		return err
	}

	configuration := p.buildConfiguration(services)
	if configuration != nil {
		configurationChan <- types.ConfigMessage{
			ProviderName:  "docker",
			Configuration: configuration,
		}
	}

	return nil
}

func (p *Provider) eventCallback(msg eventtypes.Message, configurationChan chan<- types.ConfigMessage) error {
	dockerClient, err := p.createClient()
	if err != nil {
		return err
	}

	p.swarmPool.GoCtx(func(ctx context.Context) {
		myTimer := &time.Timer{}
		lock := sync.RWMutex{}
		stopRoutine := false
		retryChan := make(chan bool)

		safe.Go(func() {
			p.listCheckAndUpdateServices(ctx, msg, configurationChan, dockerClient, retryChan)
		})

		for {
			select {
			case <-ctx.Done():
				lock.Lock()
				stopRoutine = true
				lock.Unlock()
				return
			case <-myTimer.C:
				safe.Go(func() {
					p.listCheckAndUpdateServices(ctx, msg, configurationChan, dockerClient, retryChan)
				})
			case retry := <-retryChan:
				if !retry {
					return
				}
				lock.RLock()
				if !stopRoutine {
					lock.RUnlock()
					// Sleep for 1 second between retries.
					log.Debug("Event callback: Retrying in 1 second...")
					myTimer = time.NewTimer(1 * time.Second)
				} else {
					lock.RUnlock()
					return
				}
			}
		}
	})
	return nil
}

func (p *Provider) listCheckAndUpdateServices(ctx context.Context, msg eventtypes.Message, configurationChan chan<- types.ConfigMessage, dockerClient client.APIClient, retryChan chan<- bool) {
	if len(msg.Actor.ID) == 0 {
		err := p.listAndUpdateServices(ctx, dockerClient, configurationChan)
		if err != nil {
			log.Errorf("Unable to list and update services for empty message: %v", err)
		}
		retryChan <- false
		return
	}

	retry, err := checkServiceContent(ctx, dockerClient, msg.Actor.ID)
	if err != nil {
		log.Errorf("Unable to check services: %v", err)
		retryChan <- false
		return
	}

	if retry {
		retryChan <- true
		return
	}

	err = p.listAndUpdateServices(ctx, dockerClient, configurationChan)
	if err != nil {
		log.Errorf("Unable to list and update services: %v", err)
	}
	retryChan <- false
	return

}

func checkServiceContent(ctx context.Context, dockerClient client.APIClient, actorID string) (bool, error) {
	service, err := getService(ctx, dockerClient, actorID)
	if err != nil {
		return false, err
	}

	if service == nil {
		return false, nil
	}

	if service.Spec.Mode.Global == nil && service.Spec.Mode.Replicated == nil {
		return false, fmt.Errorf("service %s mode is not valid", service.ID)
	}

	if service.Spec.Mode.Global != nil {
		// TODO: For now we just sleep for 1 second if it's a global service,
		// since we don't know how many tasks should be expected (unless we list the nodes, etc. - but there might also be a race condition).
		log.Info("Service is in global mode. Sleep for 1 second and cross our fingers we won't run into any race conditions...")
		time.Sleep(1 * time.Second)
	}

	taskList, err := listTasks(ctx, dockerClient, service.ID, string(swarmtypes.TaskStateRunning))
	if err != nil {
		return false, fmt.Errorf("unable to list tasks for service %q: %v", service.ID, err)
	}

	// Check if the task list size is correct
	retry, err := checkTaskListLength(taskList, service)
	if err != nil {
		return false, fmt.Errorf("unable to check tasks list length for service %q: %v", service.ID, err)
	}
	if retry {
		return true, nil
	}
	// Check all the service tasks state
	return checkTaskListState(taskList), nil
}

func checkTaskListLength(taskList []swarmtypes.Task, service *swarmtypes.Service) (bool, error) {
	// Retry if there are no tasks found.
	if len(taskList) == 0 {
		log.Debugf("No tasks for service %s found! Retrying...", service.ID)
		return true, nil
	}

	if service.Spec.Mode.Replicated != nil {
		if service.Spec.Mode.Replicated.Replicas == nil {
			return false, fmt.Errorf("service %s is in replicated mode, but no replicas are defined", service.ID)
		}

		// Retry if the service is in replicated mode and the list of tasks is shorter than the number of replicas.
		// We don't need to retry if the number of services is longer than the number of replicas. That means the service is being scaled in.
		numberOfReplicas := int(*service.Spec.Mode.Replicated.Replicas)
		if len(taskList) < numberOfReplicas {
			log.Debugf("Task list length for service %s is %d, expected it to be %d. Retrying...", service.ID, len(taskList), numberOfReplicas)
			return true, nil
		}
	}

	return false, nil
}

func checkTaskListState(taskList []swarmtypes.Task) bool {
	for _, task := range taskList {
		if task.Status.State != swarmtypes.TaskStateRunning {
			switch task.Status.State {
			case
				swarmtypes.TaskStateNew,
				swarmtypes.TaskStatePending,
				swarmtypes.TaskStateAssigned,
				swarmtypes.TaskStateAccepted,
				swarmtypes.TaskStatePreparing,
				swarmtypes.TaskStateStarting:
				return true
			}
		}
	}
	return false
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
		log.Debugf("Failed to network inspect on client for docker: %v", err)
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
				log.Warnf("No tasks found for service %s: %v", service.Spec.Name, err)
			} else {
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
	desiredState := swarmtypes.TaskStateRunning
	taskList, err := listTasks(ctx, dockerClient, serviceID, string(desiredState))
	if err != nil {
		return nil, err
	}

	var dockerDataList []dockerData
	for _, task := range taskList {
		if task.Status.State != desiredState {
			log.Warnf(
				"Task %s is not in the desired state (current state: %s, desired state: %s, service: %s)",
				task.ID,
				task.Status.State,
				desiredState,
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

func getService(ctx context.Context, dockerClient client.APIClient, serviceID string) (*swarmtypes.Service, error) {
	services, err := dockerClient.ServiceList(
		ctx,
		dockertypes.ServiceListOptions{
			Filters: filters.NewArgs(
				filters.Arg("id", serviceID),
			),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	if len(services) > 1 {
		return nil, fmt.Errorf("too many services found for id %q: %d instead of 1", serviceID, len(services))
	}

	return &services[0], nil
}

// swarmEventsCapabilities checks the docker API version to adapt the provider behavior
// in function of the capalities allowed by the version
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
