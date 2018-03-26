// Package check aims to provide simple "helper" methods to ease the use of
// compose (through libcompose) in (integration) tests using the go-check package.
package check

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/go-check/check"
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

// Start creates and starts services of the compose project,if no services are given, all the services are created/started.
func (p *Project) Start(c *check.C, services ...string) {
	errString := "error while creating and starting compose project"
	if services != nil && len(services) > 0 {
		errString = fmt.Sprintf("error while creating and starting services %s compose project", strings.Join(services, ","))
	}
	c.Assert(p.project.Start(services...), check.IsNil,
		check.Commentf(errString))
}

// StartOnly starts services of the compose project,if no services are given, all the services are started.
func (p *Project) StartOnly(c *check.C, services ...string) {
	errString := "error while starting compose project"
	if services != nil && len(services) > 0 {
		errString = fmt.Sprintf("error while starting services %s compose project", strings.Join(services, ","))
	}
	c.Assert(p.project.StartOnly(services...), check.IsNil,
		check.Commentf(errString))
}

// StopOnly stops services of the compose project,if no services are given, all the services are stopped.
func (p *Project) StopOnly(c *check.C, services ...string) {
	errString := "error while stopping deleting compose project"
	if services != nil && len(services) > 0 {
		errString = fmt.Sprintf("error while stopping deleting services %s compose project", strings.Join(services, ","))
	}
	c.Assert(p.project.StopOnly(services...), check.IsNil,
		check.Commentf(errString))
}

// Stop shuts down and clean services of the compose project,if no services are given, all the services are stopped/deleted.
func (p *Project) Stop(c *check.C, services ...string) {
	errString := "error while stopping and deleting compose project"
	if services != nil && len(services) > 0 {
		errString = fmt.Sprintf("error while stopping and deleting services %s compose project", strings.Join(services, ","))
	}
	c.Assert(p.project.Stop(services...), check.IsNil,
		check.Commentf(errString))
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

// NoContainer check is there is no container for the service given
// It fails if there one or more containers or if the error returned
// does not indicate an empty container list
func (p *Project) NoContainer(c *check.C, service string) {
	validErr := "No container found for '" + service + "' service"
	_, err := p.project.Container(service)
	c.Assert(err, check.NotNil,
		check.Commentf("error while getting the container for service '%s'", service))
	c.Assert(err.Error(), check.Equals, validErr,
		check.Commentf("error while getting the container for service '%s'", service))
}
