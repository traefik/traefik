package container

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// ListByFilter looks up the hosts containers with the specified filters and
// returns a list of container matching it, or an error.
func ListByFilter(ctx context.Context, clientInstance client.ContainerAPIClient, containerFilters ...map[string][]string) ([]types.Container, error) {
	filterArgs := filters.NewArgs()

	// FIXME(vdemeester) I don't like 3 for loops >_<
	for _, filter := range containerFilters {
		for key, filterValue := range filter {
			for _, value := range filterValue {
				filterArgs.Add(key, value)
			}
		}
	}

	return clientInstance.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
}

// Get looks up the hosts containers with the specified ID
// or name and returns it, or an error.
func Get(ctx context.Context, clientInstance client.ContainerAPIClient, id string) (*types.ContainerJSON, error) {
	container, err := clientInstance.ContainerInspect(ctx, id)
	if err != nil {
		if client.IsErrContainerNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &container, nil
}
