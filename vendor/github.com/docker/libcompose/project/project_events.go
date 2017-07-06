package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Events listen for real time events from containers (of the project).
func (p *Project) Events(ctx context.Context, services ...string) (chan events.ContainerEvent, error) {
	events := make(chan events.ContainerEvent)
	if len(services) == 0 {
		services = p.ServiceConfigs.Keys()
	}
	// FIXME(vdemeester) handle errors (chan) here
	for _, service := range services {
		s, err := p.CreateService(service)
		if err != nil {
			return nil, err
		}
		go s.Events(ctx, events)
	}
	return events, nil
}
