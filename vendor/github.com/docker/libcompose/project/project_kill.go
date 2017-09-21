package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Kill kills the specified services (like docker kill).
func (p *Project) Kill(ctx context.Context, signal string, services ...string) error {
	return p.perform(events.ProjectKillStart, events.ProjectKillDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServiceKillStart, events.ServiceKill, func(service Service) error {
			return service.Kill(ctx, signal)
		})
	}), nil)
}
