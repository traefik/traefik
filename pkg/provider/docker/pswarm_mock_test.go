package docker

import (
	"context"

	containertypes "github.com/moby/moby/api/types/container"
	networktypes "github.com/moby/moby/api/types/network"
	swarmtypes "github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
)

type fakeTasksClient struct {
	client.APIClient

	tasks     []swarmtypes.Task
	container containertypes.InspectResponse
	err       error
}

func (c *fakeTasksClient) TaskList(ctx context.Context, options client.TaskListOptions) (client.TaskListResult, error) {
	return client.TaskListResult{Items: c.tasks}, c.err
}

func (c *fakeTasksClient) ContainerInspect(ctx context.Context, container string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	return client.ContainerInspectResult{Container: c.container}, c.err
}

type fakeServicesClient struct {
	client.APIClient

	dockerVersion string
	networks      []networktypes.Summary
	nodes         []swarmtypes.Node
	services      []swarmtypes.Service
	tasks         []swarmtypes.Task
	err           error
}

func (c *fakeServicesClient) NodeInspect(ctx context.Context, nodeID string, options client.NodeInspectOptions) (client.NodeInspectResult, error) {
	for _, node := range c.nodes {
		if node.ID == nodeID {
			return client.NodeInspectResult{
				Node: node,
			}, nil
		}
	}
	return client.NodeInspectResult{}, c.err
}

func (c *fakeServicesClient) ServiceList(ctx context.Context, options client.ServiceListOptions) (client.ServiceListResult, error) {
	return client.ServiceListResult{Items: c.services}, c.err
}

func (c *fakeServicesClient) ServerVersion(ctx context.Context, options client.ServerVersionOptions) (client.ServerVersionResult, error) {
	return client.ServerVersionResult{APIVersion: c.dockerVersion}, c.err
}

func (c *fakeServicesClient) NetworkList(ctx context.Context, options client.NetworkListOptions) (client.NetworkListResult, error) {
	return client.NetworkListResult{Items: c.networks}, c.err
}

func (c *fakeServicesClient) TaskList(ctx context.Context, options client.TaskListOptions) (client.TaskListResult, error) {
	return client.TaskListResult{Items: c.tasks}, c.err
}
