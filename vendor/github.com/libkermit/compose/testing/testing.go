// Package testing aims to provide simple "helper" methods to ease the use of
// compose (through libcompose) in (integration) tests using built-in testing.
package testing

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/libkermit/compose"
)

// Project holds compose related project attributes
type Project struct {
	project *compose.Project
}

// CreateProject creates a compose project with the given name based on the
// specified compose files
func CreateProject(t *testing.T, name string, composeFiles ...string) *Project {
	project, err := compose.CreateProject(name, composeFiles...)
	if err != nil {
		t.Fatalf("error while creating compose project %s from %v: %s", name, composeFiles, err.Error())
	}
	return &Project{
		project: project,
	}
}

// Start creates and starts the compose project.
func (p *Project) Start(t *testing.T) {
	if err := p.project.Start(); err != nil {
		t.Fatalf("error while starting compose project: %s", err.Error())
	}
}

// Stop shuts down and clean the project
func (p *Project) Stop(t *testing.T) {
	if err := p.project.Stop(); err != nil {
		t.Fatalf("error while stopping compose project: %s", err.Error())
	}
}

// Scale scale a service up
func (p *Project) Scale(t *testing.T, service string, count int) {
	if err := p.project.Scale(service, count); err != nil {
		t.Fatalf("error while scaling the service '%s'", service)
	}
}

// Containers lists containers for a given services.
func (p *Project) Containers(t *testing.T, service string) []types.ContainerJSON {
	containers, err := p.project.Containers(service)
	if err != nil {
		t.Fatalf("error while getting the containers for service '%s'", service)
	}
	return containers
}

// Container return the one and only container for a given services.
// It fails if there is more than one container for the service.
func (p *Project) Container(t *testing.T, service string) types.ContainerJSON {
	container, err := p.project.Container(service)
	if err != nil {
		t.Fatalf("error while getting the container for service '%s'", service)
	}
	return container
}
