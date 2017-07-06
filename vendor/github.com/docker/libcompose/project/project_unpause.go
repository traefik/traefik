package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Unpause pauses the specified services containers (like docker pause).
func (p *Project) Unpause(ctx context.Context, services ...string) error {
	return p.perform(events.ProjectUnpauseStart, events.ProjectUnpauseDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServiceUnpauseStart, events.ServiceUnpause, func(service Service) error {
			return service.Unpause(ctx)
		})
	}), nil)
}
