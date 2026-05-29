package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	eventtypes "github.com/moby/moby/api/types/events"
	"github.com/moby/moby/client"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/job"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
)

const dockerName = "docker"

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	Shared       `yaml:",inline" export:"true"`
	ClientConfig `yaml:",inline" export:"true"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Watch = true
	p.ExposedByDefault = true
	p.Endpoint = "unix:///var/run/docker.sock"
	p.DefaultRule = DefaultTemplateRule
}

// Init the provider.
func (p *Provider) Init() error {
	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %w", err)
	}

	p.defaultRuleTpl = defaultRuleTpl
	return nil
}

// Provide allows the docker provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		logger := log.Ctx(routineCtx).With().Str(logs.ProviderName, dockerName).Logger()
		ctxLog := logger.WithContext(routineCtx)

		operation := func() error {
			var err error
			ctx, cancel := context.WithCancel(ctxLog)
			defer cancel()

			ctx = log.Ctx(ctx).With().Str(logs.ProviderName, dockerName).Logger().WithContext(ctx)

			dockerClient, err := p.createClient(ctxLog)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create Docker API client")
				return err
			}
			defer func() { _ = dockerClient.Close() }()

			builder := NewDynConfBuilder(p.Shared, dockerClient, false)

			serverVersion, err := dockerClient.ServerVersion(ctx, client.ServerVersionOptions{})
			if err != nil {
				logger.Error().Err(err).Msg("Failed to retrieve information of the docker client and server host")
				return err
			}

			logger.Debug().Msgf("Provider connection established with docker %s (API %s)", serverVersion.Version, serverVersion.APIVersion)

			dockerDataList, err := p.listContainers(ctx, dockerClient)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to list containers for docker")
				return err
			}

			configuration := builder.build(ctxLog, dockerDataList)
			configurationChan <- dynamic.Message{
				ProviderName:  dockerName,
				Configuration: configuration,
			}

			if p.Watch {
				startStopHandle := func(m eventtypes.Message) {
					logger.Debug().Msgf("Provider event received %+v", m)
					containers, err := p.listContainers(ctx, dockerClient)
					if err != nil {
						logger.Error().Err(err).Msg("Failed to list containers for docker")
						// Call cancel to get out of the monitor
						return
					}

					configuration := builder.build(ctx, containers)
					if configuration != nil {
						message := dynamic.Message{
							ProviderName:  dockerName,
							Configuration: configuration,
						}
						select {
						case configurationChan <- message:
						case <-ctx.Done():
						}
					}
				}

				res := dockerClient.Events(ctx, client.EventsListOptions{
					Filters: make(client.Filters).Add("type", string(eventtypes.ContainerEventType)),
				})
				for {
					select {
					case event := <-res.Messages:
						if event.Action == "start" ||
							event.Action == "die" ||
							strings.HasPrefix(string(event.Action), "health_status") {
							startStopHandle(event)
						}
					case err := <-res.Err:
						if errors.Is(err, io.EOF) {
							logger.Debug().Msg("Provider event stream closed")
						}
						return err
					case <-ctx.Done():
						return nil
					}
				}
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

func (p *Provider) createClient(ctx context.Context) (*client.Client, error) {
	return createClient(ctx, p.ClientConfig)
}

func (p *Provider) listContainers(ctx context.Context, dockerClient client.ContainerAPIClient) ([]dockerData, error) {
	containerList, err := dockerClient.ContainerList(ctx, client.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	var inspectedContainers []dockerData
	// get inspect containers
	for _, c := range containerList.Items {
		dData := inspectContainers(ctx, dockerClient, c.ID)
		if len(dData.Name) == 0 {
			continue
		}

		extraConf, err := p.extractDockerLabels(dData)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("Skip container %s", getServiceName(dData))
			continue
		}

		dData.ExtraConf = extraConf

		inspectedContainers = append(inspectedContainers, dData)
	}

	return inspectedContainers, nil
}
