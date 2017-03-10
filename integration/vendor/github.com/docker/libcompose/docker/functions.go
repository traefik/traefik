package docker

import (
	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

// GetContainersByFilter looks up the hosts containers with the specified filters and
// returns a list of container matching it, or an error.
func GetContainersByFilter(ctx context.Context, clientInstance client.APIClient, containerFilters ...map[string][]string) ([]types.Container, error) {
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
		All:    true,
		Filter: filterArgs,
	})
}

// GetContainer looks up the hosts containers with the specified ID
// or name and returns it, or an error.
func GetContainer(ctx context.Context, clientInstance client.APIClient, id string) (*types.ContainerJSON, error) {
	container, err := clientInstance.ContainerInspect(ctx, id)
	if err != nil {
		if client.IsErrContainerNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &container, nil
}
