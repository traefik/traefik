package docker

import (
	"context"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
)

type fakeTasksClient struct {
	dockerclient.APIClient
	tasks     []swarm.Task
	container dockertypes.ContainerJSON
	err       error
}

func (c *fakeTasksClient) TaskList(ctx context.Context, options dockertypes.TaskListOptions) ([]swarm.Task, error) {
	return c.tasks, c.err
}

func (c *fakeTasksClient) ContainerInspect(ctx context.Context, container string) (dockertypes.ContainerJSON, error) {
	return c.container, c.err
}

type fakeServicesClient struct {
	dockerclient.APIClient
	dockerVersion string
	networks      []dockertypes.NetworkResource
	services      []swarm.Service
	tasks         []swarm.Task
	err           error
}

func (c *fakeServicesClient) ServiceList(ctx context.Context, options dockertypes.ServiceListOptions) ([]swarm.Service, error) {
	return c.services, c.err
}

func (c *fakeServicesClient) ServerVersion(ctx context.Context) (dockertypes.Version, error) {
	return dockertypes.Version{APIVersion: c.dockerVersion}, c.err
}

func (c *fakeServicesClient) NetworkList(ctx context.Context, options dockertypes.NetworkListOptions) ([]dockertypes.NetworkResource, error) {
	return c.networks, c.err
}

func (c *fakeServicesClient) TaskList(ctx context.Context, options dockertypes.TaskListOptions) ([]swarm.Task, error) {
	return c.tasks, c.err
}
