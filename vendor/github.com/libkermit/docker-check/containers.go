package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/go-check/check"
	"github.com/libkermit/docker"
)

// Create lets you create a container with the specified image, and default
// configuration.
func (p *Project) Create(c *check.C, image string) types.ContainerJSON {
	container, err := p.project.Create(image)
	c.Assert(err, check.IsNil,
		check.Commentf("error while creating the container with image %s", image))
	return container
}

// CreateWithConfig lets you create a container with the specified image, and
// some custom simple configuration.
func (p *Project) CreateWithConfig(c *check.C, image string, containerConfig docker.ContainerConfig) types.ContainerJSON {
	container, err := p.project.CreateWithConfig(image, containerConfig)
	c.Assert(err, check.IsNil,
		check.Commentf("error while creating the container with image %s and config %v", image, containerConfig))
	return container
}

// Start lets you create and start a container with the specified image, and
// default configuration.
func (p *Project) Start(c *check.C, image string) types.ContainerJSON {
	container, err := p.project.Start(image)
	c.Assert(err, check.IsNil,
		check.Commentf("error while starting the container with image %s", image))
	return container
}

// StartWithConfig lets you create and start a container with the specified
// image, and some custom simple configuration.
func (p *Project) StartWithConfig(c *check.C, image string, containerConfig docker.ContainerConfig) types.ContainerJSON {
	container, err := p.project.StartWithConfig(image, containerConfig)
	c.Assert(err, check.IsNil,
		check.Commentf("error while creating the starting with image %s", image))
	return container
}

// Stop stops the container with a default timeout.
func (p *Project) Stop(c *check.C, containerID string) {
	c.Assert(p.project.Stop(containerID), check.IsNil,
		check.Commentf("error while stopping container %s", containerID))
}

// StopWithTimeout stops the container with the specified timeout.
func (p *Project) StopWithTimeout(c *check.C, containerID string, timeout int) {
	c.Assert(p.project.StopWithTimeout(containerID, timeout), check.IsNil,
		check.Commentf("error while stopping container %s with timeout %d", containerID, timeout))
}

// Inspect returns the container informations
func (p *Project) Inspect(c *check.C, containerID string) types.ContainerJSON {
	container, err := p.project.Inspect(containerID)
	c.Assert(err, check.IsNil,
		check.Commentf("error while inspecting with container %s", containerID))
	return container
}

// List lists the containers managed by kermit
func (p *Project) List(c *check.C) []types.Container {
	containers, err := p.project.List()
	c.Assert(err, check.IsNil,
		check.Commentf("error while listing containers"))
	return containers
}

// IsRunning checks if the container is running or not
func (p *Project) IsRunning(c *check.C, containerID string) bool {
	isRunning, err := p.project.IsRunning(containerID)
	c.Assert(err, check.IsNil,
		check.Commentf("error while inspecting for running with container %s", containerID))
	return isRunning
}

// IsStopped checks if the container is running or not
func (p *Project) IsStopped(c *check.C, containerID string) bool {
	isStopped, err := p.project.IsPaused(containerID)
	c.Assert(err, check.IsNil,
		check.Commentf("error while inspecting for stopped with container %s", containerID))
	return isStopped
}

// IsPaused checks if the container is running or not
func (p *Project) IsPaused(c *check.C, containerID string) bool {
	isPaused, err := p.project.IsPaused(containerID)
	c.Assert(err, check.IsNil,
		check.Commentf("error while inspecting for paused with container %s", containerID))
	return isPaused
}

// Remove removes the container
func (p *Project) Remove(c *check.C, containerID string) {
	c.Assert(p.project.Remove(containerID), check.IsNil,
		check.Commentf("error while removing the container %s", containerID))
}

// Clean stops and removes (by default, controllable with the keep) kermit containers
func (p *Project) Clean(c *check.C, keep bool) {
	c.Assert(p.project.Clean(keep), check.IsNil,
		check.Commentf("error while cleaning the containers"))
}
