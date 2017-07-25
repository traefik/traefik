package network

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"golang.org/x/net/context"
)

// FakeClient is a fake NetworkAPIClient
type FakeClient struct {
	NetworkInspectFunc func(ctx context.Context, networkID string, verbose bool) (types.NetworkResource, error)
}

// NetworkConnect fakes connecting to a network
func (c *FakeClient) NetworkConnect(ctx context.Context, networkID, container string, config *network.EndpointSettings) error {
	return nil
}

// NetworkCreate fakes creating a network
func (c *FakeClient) NetworkCreate(ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error) {
	return types.NetworkCreateResponse{}, nil
}

// NetworkDisconnect fakes disconencting from a network
func (c *FakeClient) NetworkDisconnect(ctx context.Context, networkID, container string, force bool) error {
	return nil
}

// NetworkInspect fakes inspecting a network
func (c *FakeClient) NetworkInspect(ctx context.Context, networkID string, verbose bool) (types.NetworkResource, error) {
	if c.NetworkInspectFunc != nil {
		return c.NetworkInspectFunc(ctx, networkID, verbose)
	}
	return types.NetworkResource{}, nil
}

// NetworkInspectWithRaw fakes inspecting a network with a raw response
func (c *FakeClient) NetworkInspectWithRaw(ctx context.Context, networkID string, verbose bool) (types.NetworkResource, []byte, error) {
	return types.NetworkResource{}, nil, nil
}

// NetworkList fakes listing networks
func (c *FakeClient) NetworkList(ctx context.Context, options types.NetworkListOptions) ([]types.NetworkResource, error) {
	return nil, nil
}

// NetworkRemove fakes removing networks
func (c *FakeClient) NetworkRemove(ctx context.Context, networkID string) error {
	return nil
}

// NetworksPrune fakes pruning networks
func (c *FakeClient) NetworksPrune(ctx context.Context, pruneFilter filters.Args) (types.NetworksPruneReport, error) {
	return types.NetworksPruneReport{}, nil
}
