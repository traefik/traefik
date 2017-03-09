package docker

import (
	"os"
	"path/filepath"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/docker/network"
	"github.com/docker/libcompose/labels"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
)

// ComposeVersion is name of docker-compose.yml file syntax supported version
const ComposeVersion = "1.5.0"

// NewProject creates a Project with the specified context.
func NewProject(context *Context, parseOptions *config.ParseOptions) (project.APIProject, error) {
	if context.ResourceLookup == nil {
		context.ResourceLookup = &lookup.FileConfigLookup{}
	}

	if context.EnvironmentLookup == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		context.EnvironmentLookup = &lookup.ComposableEnvLookup{
			Lookups: []config.EnvironmentLookup{
				&lookup.EnvfileLookup{
					Path: filepath.Join(cwd, ".env"),
				},
				&lookup.OsEnvLookup{},
			},
		}
	}

	if context.AuthLookup == nil {
		context.AuthLookup = NewConfigAuthLookup(context)
	}

	if context.ServiceFactory == nil {
		context.ServiceFactory = &ServiceFactory{
			context: context,
		}
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

	// FIXME(vdemeester) Remove the context duplication ?
	runtime := &Project{
		clientFactory: context.ClientFactory,
	}
	p := project.NewProject(&context.Context, runtime, parseOptions)

	err := p.Parse()
	if err != nil {
		return nil, err
	}

	if err = context.open(); err != nil {
		logrus.Errorf("Failed to open project %s: %v", p.Name, err)
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
		Filter: filter,
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
