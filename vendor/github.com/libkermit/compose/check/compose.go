// Package check aims to provide simple "helper" methods to ease the use of
// compose (through libcompose) in (integration) tests using the go-check package.
package check

import (
	"github.com/go-check/check"

	"github.com/docker/docker/api/types"
	"github.com/libkermit/compose"
)

// Project holds compose related project attributes
type Project struct {
	project *compose.Project
}

// CreateProject creates a compose project with the given name based on the
// specified compose files
func CreateProject(c *check.C, name string, composeFiles ...string) *Project {
	project, err := compose.CreateProject(name, composeFiles...)
	c.Assert(err, check.IsNil,
		check.Commentf("error while creating compose project %s from %v", name, composeFiles))
	return &Project{
		project: project,
	}
}

// Start creates and starts the compose project.
func (p *Project) Start(c *check.C) {
	c.Assert(p.project.Start(), check.IsNil,
		check.Commentf("error while starting compose project"))
}

// Stop shuts down and clean the project
func (p *Project) Stop(c *check.C) {
	c.Assert(p.project.Stop(), check.IsNil,
		check.Commentf("error while stopping compose project"))
}

// Scale scale a service up
func (p *Project) Scale(c *check.C, service string, count int) {
	c.Assert(p.project.Scale(service, count), check.IsNil,
		check.Commentf("error while scaling the service '%s'", service))
}

// Containers lists containers for a given services.
func (p *Project) Containers(c *check.C, service string) []types.ContainerJSON {
	containers, err := p.project.Containers(service)
	c.Assert(err, check.IsNil,
		check.Commentf("error while getting the containers for service '%s'", service))
	return containers
}

// Container return the one and only container for a given services.
// It fails if there is more than one container for the service.
func (p *Project) Container(c *check.C, service string) types.ContainerJSON {
	container, err := p.project.Container(service)
	c.Assert(err, check.IsNil,
		check.Commentf("error while getting the container for service '%s'", service))
	return container
}
