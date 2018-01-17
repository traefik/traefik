package docker

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker/auth"
	"github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/docker/network"
	"github.com/docker/libcompose/docker/service"
	"github.com/docker/libcompose/docker/volume"
	"github.com/docker/libcompose/labels"
	"github.com/docker/libcompose/project"
	"github.com/sirupsen/logrus"
)

// NewProject creates a Project with the specified context.
func NewProject(context *ctx.Context, parseOptions *config.ParseOptions) (project.APIProject, error) {

	if err := context.LookupConfig(); err != nil {
		logrus.Errorf("Failed to load docker config: %v", err)
	}

	if context.AuthLookup == nil {
		context.AuthLookup = auth.NewConfigLookup(context.ConfigFile)
	}

	if context.ServiceFactory == nil {
		context.ServiceFactory = service.NewFactory(context)
	}

	if context.ClientFactory == nil {
		factory, err := client.NewDefaultFactory(client.Options{})
		if err != nil {
			return nil, err
		}
		context.ClientFactory = factory
	}

	if context.NetworksFactory == nil {
		networksFactory := &network.DockerFactory{
			ClientFactory: context.ClientFactory,
		}
		context.NetworksFactory = networksFactory
	}

	if context.VolumesFactory == nil {
		volumesFactory := &volume.DockerFactory{
			ClientFactory: context.ClientFactory,
		}
		context.VolumesFactory = volumesFactory
	}

	// FIXME(vdemeester) Remove the context duplication ?
	runtime := &Project{
		clientFactory: context.ClientFactory,
	}
	p := project.NewProject(&context.Context, runtime, parseOptions)

	err := p.Parse()
	if err != nil {
		return nil, err
	}

	return p, err
}

// Project implements project.RuntimeProject and define docker runtime specific methods.
type Project struct {
	clientFactory client.Factory
}

// RemoveOrphans implements project.RuntimeProject.RemoveOrphans.
// It will remove orphan containers that are part of the project but not to any services.
func (p *Project) RemoveOrphans(ctx context.Context, projectName string, serviceConfigs *config.ServiceConfigs) error {
	client := p.clientFactory.Create(nil)
	filter := filters.NewArgs()
	filter.Add("label", labels.PROJECT.EqString(projectName))
	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		Filters: filter,
	})
	if err != nil {
		return err
	}
	currentServices := map[string]struct{}{}
	for _, serviceName := range serviceConfigs.Keys() {
		currentServices[serviceName] = struct{}{}
	}
	for _, container := range containers {
		serviceLabel := container.Labels[labels.SERVICE.Str()]
		if _, ok := currentServices[serviceLabel]; !ok {
			if err := client.ContainerKill(ctx, container.ID, "SIGKILL"); err != nil {
				return err
			}
			if err := client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
				Force: true,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
