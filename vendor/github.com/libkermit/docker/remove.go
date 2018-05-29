package docker

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
)

// Remove removes the container
func (p *Project) Remove(containerID string) error {
	return p.Client.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
		Force: true,
	})
}
