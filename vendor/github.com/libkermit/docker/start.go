package docker

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
)

// Start lets you create and start a container with the specified image, and
// default configuration.
func (p *Project) Start(image string) (types.ContainerJSON, error) {
	return p.StartWithConfig(image, ContainerConfig{})
}

// StartWithConfig lets you create and start a container with the specified
// image, and some custom simple configuration.
func (p *Project) StartWithConfig(image string, containerConfig ContainerConfig) (types.ContainerJSON, error) {
	container, err := p.CreateWithConfig(image, containerConfig)
	if err != nil {
		return types.ContainerJSON{}, err
	}

	if err := p.Client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{}); err != nil {
		return container, err
	}

	return p.Inspect(container.ID)
}
