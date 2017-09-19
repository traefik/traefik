package docker

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
)

// Remove removes the container
func (p *Project) Remove(containerID string) error {
	if err := p.Client.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return err
	}
	return nil
}
