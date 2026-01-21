package docker

import (
	"context"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
)

type fakeTasksClient struct {
	dockerclient.APIClient

	tasks     []swarmtypes.Task
	container containertypes.InspectResponse
	err       error
}

func (c *fakeTasksClient) TaskList(ctx context.Context, options swarmtypes.TaskListOptions) ([]swarmtypes.Task, error) {
	return c.tasks, c.err
}

func (c *fakeTasksClient) ContainerInspect(ctx context.Context, container string) (containertypes.InspectResponse, error) {
	return c.container, c.err
}

type fakeServicesClient struct {
	dockerclient.APIClient

	dockerVersion string
	networks      []networktypes.Summary
	nodes         []swarmtypes.Node
	services      []swarmtypes.Service
	tasks         []swarmtypes.Task
	err           error
}

func (c *fakeServicesClient) NodeInspectWithRaw(ctx context.Context, nodeID string) (swarmtypes.Node, []byte, error) {
	for _, node := range c.nodes {
		if node.ID == nodeID {
			return node, nil, nil
		}
	}
	return swarmtypes.Node{}, nil, c.err
}

func (c *fakeServicesClient) ServiceList(ctx context.Context, options swarmtypes.ServiceListOptions) ([]swarmtypes.Service, error) {
	return c.services, c.err
}

func (c *fakeServicesClient) ServerVersion(ctx context.Context) (dockertypes.Version, error) {
	return dockertypes.Version{APIVersion: c.dockerVersion}, c.err
}

func (c *fakeServicesClient) NetworkList(ctx context.Context, options networktypes.ListOptions) ([]networktypes.Summary, error) {
	return c.networks, c.err
}

func (c *fakeServicesClient) TaskList(ctx context.Context, options swarmtypes.TaskListOptions) ([]swarmtypes.Task, error) {
	return c.tasks, c.err
}
