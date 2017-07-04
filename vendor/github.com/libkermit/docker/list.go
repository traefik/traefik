package docker

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

// List lists the containers managed by kermit
func (p *Project) List() ([]types.Container, error) {
	filters := filters.NewArgs()
	for key, value := range p.Labels {
		filters.Add("label", fmt.Sprintf("%s=%s", key, value))
	}
	containers, err := p.Client.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}
	return containers, nil
}
