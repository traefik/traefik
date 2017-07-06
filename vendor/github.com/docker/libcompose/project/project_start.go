package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Start starts the specified services (like docker start).
func (p *Project) Start(ctx context.Context, services ...string) error {
	return p.perform(events.ProjectStartStart, events.ProjectStartDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceStartStart, events.ServiceStart, func(service Service) error {
			return service.Start(ctx)
		})
	}), nil)
}
