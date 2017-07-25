package testing

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/libkermit/docker"
)

// Create lets you create a container with the specified image, and default
// configuration.
func (p *Project) Create(t *testing.T, image string) types.ContainerJSON {
	container, err := p.project.Create(image)
	if err != nil {
		t.Fatalf("error while creating the container with image %s: %s", image, err.Error())
	}
	return container
}

// CreateWithConfig lets you create a container with the specified image, and
// some custom simple configuration.
func (p *Project) CreateWithConfig(t *testing.T, image string, containerConfig docker.ContainerConfig) types.ContainerJSON {
	container, err := p.project.CreateWithConfig(image, containerConfig)
	if err != nil {
		t.Fatalf("error while creating the container with image %s and config %v: %s", image, containerConfig, err.Error())
	}
	return container
}

// Start lets you create and start a container with the specified image, and
// default configuration.
func (p *Project) Start(t *testing.T, image string) types.ContainerJSON {
	container, err := p.project.Start(image)
	if err != nil {
		t.Fatalf("error while starting the container with image %s: %s", image, err.Error())
	}
	return container
}

// StartWithConfig lets you create and start a container with the specified
// image, and some custom simple configuration.
func (p *Project) StartWithConfig(t *testing.T, image string, containerConfig docker.ContainerConfig) types.ContainerJSON {
	container, err := p.project.StartWithConfig(image, containerConfig)
	if err != nil {
		t.Fatalf("error while creating the starting with image %s: %s", image, err.Error())
	}
	return container
}

// Stop stops the container with a default timeout.
func (p *Project) Stop(t *testing.T, containerID string) {
	if err := p.project.Stop(containerID); err != nil {
		t.Fatalf("error while stopping container %s", containerID)
	}
}

// StopWithTimeout stops the container with the specified timeout.
func (p *Project) StopWithTimeout(t *testing.T, containerID string, timeout int) {
	if err := p.project.StopWithTimeout(containerID, timeout); err != nil {
		t.Fatalf("error while stopping container %s with timeout %d", containerID, timeout)
	}
}

// Inspect returns the container informations
func (p *Project) Inspect(t *testing.T, containerID string) types.ContainerJSON {
	container, err := p.project.Inspect(containerID)
	if err != nil {
		t.Fatalf("error while inspecting with container %s: %s", containerID, err.Error())
	}
	return container
}

// List lists the containers managed by kermit
func (p *Project) List(t *testing.T) []types.Container {
	containers, err := p.project.List()
	if err != nil {
		t.Fatalf("error while listing containers: %s", err.Error())
	}
	return containers
}

// IsRunning checks if the container is running or not
func (p *Project) IsRunning(t *testing.T, containerID string) bool {
	isRunning, err := p.project.IsRunning(containerID)
	if err != nil {
		t.Fatalf("error while inspecting for running with container %s: %s", containerID, err.Error())
	}
	return isRunning
}

// IsStopped checks if the container is running or not
func (p *Project) IsStopped(t *testing.T, containerID string) bool {
	isStopped, err := p.project.IsPaused(containerID)
	if err != nil {
		t.Fatalf("error while inspecting for stopped with container %s: %s", containerID, err.Error())
	}
	return isStopped
}

// IsPaused checks if the container is running or not
func (p *Project) IsPaused(t *testing.T, containerID string) bool {
	isPaused, err := p.project.IsPaused(containerID)
	if err != nil {
		t.Fatalf("error while inspecting for paused with container %s: %s", containerID, err.Error())
	}
	return isPaused
}

// Remove removes the container
func (p *Project) Remove(t *testing.T, containerID string) {
	if err := p.project.Remove(containerID); err != nil {
		t.Fatalf("error while removing the container %s: %s", containerID, err.Error())
	}
}

// Clean stops and removes (by default, controllable with the keep) kermit containers
func (p *Project) Clean(t *testing.T, keep bool) {
	if err := p.project.Clean(keep); err != nil {
		t.Fatalf("error while cleaning the containers %s", err.Error())
	}
}
