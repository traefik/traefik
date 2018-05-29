// Package compose aims to provide simple "helper" methods to ease the use of
// compose (through libcompose) in (integration) tests.
package compose

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
	d "github.com/libkermit/docker"
)

// Project holds compose related project attributes
type Project struct {
	composeFiles   []string
	composeProject project.APIProject
	name           string
	listenChan     chan events.Event
	started        chan struct{}
	stopped        chan struct{}
	deleted        chan struct{}
	client         client.APIClient
	hasOpenedChan  bool
}

// CreateProject creates a compose project with the given name based on the
// specified compose files
func CreateProject(name string, composeFiles ...string) (*Project, error) {
	// FIXME(vdemeester) temporarly normalize the project name, should not be needed.
	r := regexp.MustCompile("[^a-z0-9]+")
	name = r.ReplaceAllString(strings.ToLower(name), "")

	apiClient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	ping := types.Ping{APIVersion: d.CurrentAPIVersion}
	apiClient.NegotiateAPIVersionPing(ping)

	composeProject, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: composeFiles,
			ProjectName:  name,
		},
	}, &config.ParseOptions{
		Interpolate: true,
		Validate:    true,
	})
	if err != nil {
		return nil, err
	}
	p := &Project{
		composeFiles:   composeFiles,
		composeProject: composeProject,
		name:           name,
		listenChan:     make(chan events.Event),
		started:        make(chan struct{}),
		stopped:        make(chan struct{}),
		deleted:        make(chan struct{}),
		client:         apiClient,
		hasOpenedChan:  true,
	}

	// Listen to compose events
	go p.startListening()
	p.composeProject.AddListener(p.listenChan)

	return p, nil
}

// Start creates and starts the compose project.
func (p *Project) Start(services ...string) error {
	// If project chan are closed, recreate new compose project
	if !p.hasOpenedChan {
		newProject, _ := CreateProject(p.name, p.composeFiles...)
		*p = *newProject
	}
	ctx := context.Background()
	err := p.composeProject.Create(ctx, options.Create{}, services...)
	if err != nil {
		return err
	}

	return p.StartOnly(services...)
}

// StartOnly only starts created services which are stopped.
func (p *Project) StartOnly(services ...string) error {
	ctx := context.Background()
	err := p.composeProject.Start(ctx, services...)
	if err != nil {
		return err
	}
	// Wait for compose to start
	<-p.started
	return nil
}

// StopOnly only stop services without delete them.
func (p *Project) StopOnly(services ...string) error {
	ctx := context.Background()
	err := p.composeProject.Stop(ctx, 10, services...)
	if err != nil {
		return err
	}
	<-p.stopped
	return nil
}

// Stop shuts down and clean the project
func (p *Project) Stop(services ...string) error {
	// FIXME(vdemeester) handle timeout
	err := p.StopOnly(services...)
	if err != nil {
		return err
	}

	err = p.composeProject.Delete(context.Background(), options.Delete{}, services...)
	if err != nil {
		return err
	}
	<-p.deleted

	existingContainers, err := p.existContainers(project.AnyState)
	if err != nil {
		return err
	}
	// Close channels only if there are no running services
	if !existingContainers {
		p.hasOpenedChan = false
		close(p.started)
		close(p.stopped)
		close(p.deleted)
		close(p.listenChan)
	}
	return nil
}

// Check if containers exist in the desirated state for the given services
func (p *Project) existContainers(stateFiltered project.State, services ...string) (bool, error) {
	existingContainers := false
	var err error
	containersFound, err := p.composeProject.Containers(context.Background(), project.Filter{stateFiltered})
	if err == nil && containersFound != nil && len(containersFound) > 0 {
		existingContainers = true
	}
	return existingContainers, err
}

// Scale scale a service up
func (p *Project) Scale(service string, count int) error {
	return p.composeProject.Scale(context.Background(), 10, map[string]int{
		service: count,
	})
}

func (p *Project) startListening() {
	for event := range p.listenChan {
		// FIXME Add a timeout on event ?
		if event.EventType == events.ProjectStartDone {
			p.started <- struct{}{}
		}
		if event.EventType == events.ProjectStopDone {
			p.stopped <- struct{}{}
		}
		if event.EventType == events.ProjectDeleteDone {
			p.deleted <- struct{}{}
		}
	}
}

// Containers lists containers for a given services.
func (p *Project) Containers(service string) ([]types.ContainerJSON, error) {
	ctx := context.Background()
	containers := []types.ContainerJSON{}
	// Let's use engine-api for now as there is nothing really useful in
	// libcompose for now.
	filter := filters.NewArgs()
	filter.Add("label", "com.docker.compose.project="+p.name)
	filter.Add("label", "com.docker.compose.service="+service)
	containerList, err := p.client.ContainerList(ctx, types.ContainerListOptions{
		Filters: filter,
	})
	if err != nil {
		return containers, err
	}
	for _, c := range containerList {
		container, err := p.client.ContainerInspect(ctx, c.ID)
		if err != nil {
			return containers, err
		}
		containers = append(containers, container)
	}
	return containers, nil
}

// Container returns the one and only container for a given services. It returns an error
// if the service has more than one container (in case of scale)
func (p *Project) Container(service string) (types.ContainerJSON, error) {
	containers, err := p.Containers(service)
	if err != nil {
		return types.ContainerJSON{}, err
	}
	if len(containers) > 1 {
		return types.ContainerJSON{}, fmt.Errorf("More than one container are running for '%s' service", service)
	}
	if len(containers) == 0 {
		return types.ContainerJSON{}, fmt.Errorf("No container found for '%s' service", service)
	}
	return containers[0], nil
}
