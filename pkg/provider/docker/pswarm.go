package docker

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/job"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
)

// SwarmAPIVersion is a constant holding the version of the Provider API traefik will use.
const SwarmAPIVersion = "1.24"

const swarmName = "swarm"

var _ provider.Provider = (*SwarmProvider)(nil)

// SwarmProvider holds configurations of the provider.
type SwarmProvider struct {
	Shared       `yaml:",inline" export:"true"`
	ClientConfig `yaml:",inline" export:"true"`

	RefreshSeconds ptypes.Duration `description:"Polling interval for swarm mode." json:"refreshSeconds,omitempty" toml:"refreshSeconds,omitempty" yaml:"refreshSeconds,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (p *SwarmProvider) SetDefaults() {
	p.Watch = true
	p.ExposedByDefault = true
	p.Endpoint = "unix:///var/run/docker.sock"
	p.RefreshSeconds = ptypes.Duration(15 * time.Second)
	p.DefaultRule = DefaultTemplateRule
}

// Init the provider.
func (p *SwarmProvider) Init() error {
	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %w", err)
	}

	p.defaultRuleTpl = defaultRuleTpl
	return nil
}

func (p *SwarmProvider) createClient(ctx context.Context) (*client.Client, error) {
	p.ClientConfig.apiVersion = SwarmAPIVersion
	return createClient(ctx, p.ClientConfig)
}

// Provide allows the docker provider to provide configurations to traefik using the given configuration channel.
func (p *SwarmProvider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		logger := log.Ctx(routineCtx).With().Str(logs.ProviderName, swarmName).Logger()
		ctxLog := logger.WithContext(routineCtx)

		operation := func() error {
			var err error
			ctx, cancel := context.WithCancel(ctxLog)
			defer cancel()

			ctx = log.Ctx(ctx).With().Str(logs.ProviderName, swarmName).Logger().WithContext(ctx)

			dockerClient, err := p.createClient(ctx)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create Docker API client")
				return err
			}
			defer func() { _ = dockerClient.Close() }()

			builder := NewDynConfBuilder(p.Shared, dockerClient, true)

			serverVersion, err := dockerClient.ServerVersion(ctx)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to retrieve information of the docker client and server host")
				return err
			}

			logger.Debug().Msgf("Provider connection established with docker %s (API %s)", serverVersion.Version, serverVersion.APIVersion)

			dockerDataList, err := p.listServices(ctx, dockerClient)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to list services for docker swarm mode")
				return err
			}

			configuration := builder.build(ctxLog, dockerDataList)
			configurationChan <- dynamic.Message{
				ProviderName:  swarmName,
				Configuration: configuration,
			}
			if p.Watch {
				errChan := make(chan error)

				// TODO: This need to be change. Linked to Swarm events docker/docker#23827
				ticker := time.NewTicker(time.Duration(p.RefreshSeconds))

				pool.GoCtx(func(ctx context.Context) {
					logger := log.Ctx(ctx).With().Str(logs.ProviderName, swarmName).Logger()
					ctx = logger.WithContext(ctx)

					defer close(errChan)
					for {
						select {
						case <-ticker.C:
							services, err := p.listServices(ctx, dockerClient)
							if err != nil {
								logger.Error().Err(err).Msg("Failed to list services for docker swarm mode")
								errChan <- err
								return
							}

							configuration := builder.build(ctx, services)
							if configuration != nil {
								configurationChan <- dynamic.Message{
									ProviderName:  swarmName,
									Configuration: configuration,
								}
							}

						case <-ctx.Done():
							ticker.Stop()
							return
						}
					}
				})
				if err, ok := <-errChan; ok {
					return err
				}
				// channel closed
			}
			return nil
		}

		notify := func(err error, time time.Duration) {
			logger.Error().Err(err).Msgf("Provider error, retrying in %s", time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxLog), notify)
		if err != nil {
			logger.Error().Err(err).Msg("Cannot retrieve data")
		}
	})

	return nil
}

func (p *SwarmProvider) listServices(ctx context.Context, dockerClient client.APIClient) ([]dockerData, error) {
	logger := log.Ctx(ctx)

	serviceList, err := dockerClient.ServiceList(ctx, dockertypes.ServiceListOptions{})
	if err != nil {
		return nil, err
	}

	serverVersion, err := dockerClient.ServerVersion(ctx)
	if err != nil {
		return nil, err
	}

	networkListArgs := filters.NewArgs()
	// https://docs.docker.com/engine/api/v1.29/#tag/Network (Docker 17.06)
	if versions.GreaterThanOrEqualTo(serverVersion.APIVersion, "1.29") {
		networkListArgs.Add("scope", "swarm")
	} else {
		networkListArgs.Add("driver", "overlay")
	}

	networkList, err := dockerClient.NetworkList(ctx, dockertypes.NetworkListOptions{Filters: networkListArgs})
	if err != nil {
		logger.Debug().Err(err).Msg("Failed to network inspect on client for docker")
		return nil, err
	}

	networkMap := make(map[string]*dockertypes.NetworkResource)
	for _, network := range networkList {
		networkMap[network.ID] = &network
	}

	var dockerDataList []dockerData
	var dockerDataListTasks []dockerData

	for _, service := range serviceList {
		dData, err := p.parseService(ctx, service, networkMap)
		if err != nil {
			logger.Error().Err(err).Msgf("Skip container %s", getServiceName(dData))
			continue
		}

		if dData.ExtraConf.LBSwarm {
			if len(dData.NetworkSettings.Networks) > 0 {
				dockerDataList = append(dockerDataList, dData)
			}
		} else {
			isGlobalSvc := service.Spec.Mode.Global != nil
			dockerDataListTasks, err = listTasks(ctx, dockerClient, service.ID, dData, networkMap, isGlobalSvc)
			if err != nil {
				logger.Warn().Err(err).Send()
			} else {
				dockerDataList = append(dockerDataList, dockerDataListTasks...)
			}
		}
	}

	return dockerDataList, err
}

func (p *SwarmProvider) parseService(ctx context.Context, service swarmtypes.Service, networkMap map[string]*dockertypes.NetworkResource) (dockerData, error) {
	logger := log.Ctx(ctx)

	dData := dockerData{
		ID:              service.ID,
		ServiceName:     service.Spec.Annotations.Name,
		Name:            service.Spec.Annotations.Name,
		Labels:          service.Spec.Annotations.Labels,
		NetworkSettings: networkSettings{},
	}

	extraConf, err := p.extractSwarmLabels(dData)
	if err != nil {
		return dockerData{}, err
	}
	dData.ExtraConf = extraConf

	if service.Spec.EndpointSpec == nil {
		return dData, nil
	}
	if service.Spec.EndpointSpec.Mode == swarmtypes.ResolutionModeDNSRR {
		if dData.ExtraConf.LBSwarm {
			logger.Warn().Msgf("Ignored %s endpoint-mode not supported, service name: %s. Fallback to Traefik load balancing", swarmtypes.ResolutionModeDNSRR, service.Spec.Annotations.Name)
		}
	} else if service.Spec.EndpointSpec.Mode == swarmtypes.ResolutionModeVIP {
		dData.NetworkSettings.Networks = make(map[string]*networkData)
		for _, virtualIP := range service.Endpoint.VirtualIPs {
			networkService := networkMap[virtualIP.NetworkID]
			if networkService == nil {
				logger.Debug().Msgf("Network not found, id: %s", virtualIP.NetworkID)
				continue
			}
			if len(virtualIP.Addr) == 0 {
				logger.Debug().Msgf("No virtual IPs found in network %s", virtualIP.NetworkID)
				continue
			}
			ip, _, _ := net.ParseCIDR(virtualIP.Addr)
			network := &networkData{
				Name: networkService.Name,
				ID:   virtualIP.NetworkID,
				Addr: ip.String(),
			}
			dData.NetworkSettings.Networks[network.Name] = network
		}
	}
	return dData, nil
}

func listTasks(ctx context.Context, dockerClient client.APIClient, serviceID string,
	serviceDockerData dockerData, networkMap map[string]*dockertypes.NetworkResource, isGlobalSvc bool,
) ([]dockerData, error) {
	serviceIDFilter := filters.NewArgs()
	serviceIDFilter.Add("service", serviceID)
	serviceIDFilter.Add("desired-state", "running")

	taskList, err := dockerClient.TaskList(ctx, dockertypes.TaskListOptions{Filters: serviceIDFilter})
	if err != nil {
		return nil, err
	}

	var dockerDataList []dockerData
	for _, task := range taskList {
		if task.Status.State != swarmtypes.TaskStateRunning {
			continue
		}
		dData := parseTasks(ctx, task, serviceDockerData, networkMap, isGlobalSvc)
		if len(dData.NetworkSettings.Networks) > 0 {
			dockerDataList = append(dockerDataList, dData)
		}
	}
	return dockerDataList, err
}

func parseTasks(ctx context.Context, task swarmtypes.Task, serviceDockerData dockerData,
	networkMap map[string]*dockertypes.NetworkResource, isGlobalSvc bool,
) dockerData {
	dData := dockerData{
		ID:              task.ID,
		ServiceName:     serviceDockerData.Name,
		Name:            serviceDockerData.Name + "." + strconv.Itoa(task.Slot),
		Labels:          serviceDockerData.Labels,
		ExtraConf:       serviceDockerData.ExtraConf,
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
					log.Ctx(ctx).Debug().Msgf("No IP addresses found for network %s", virtualIP.Network.ID)
				}
			}
		}
	}
	return dData
}
