package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Pause pauses the specified services containers (like docker pause).
func (p *Project) Pause(ctx context.Context, services ...string) error {
	return p.perform(events.ProjectPauseStart, events.ProjectPauseDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServicePauseStart, events.ServicePause, func(service Service) error {
			return service.Pause(ctx)
		})
	}), nil)
}
