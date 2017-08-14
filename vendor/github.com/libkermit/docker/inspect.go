package docker

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
)

// Inspect returns the container informations
func (p *Project) Inspect(containerID string) (types.ContainerJSON, error) {
	container, err := p.Client.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return types.ContainerJSON{}, err
	}

	return container, nil
}
