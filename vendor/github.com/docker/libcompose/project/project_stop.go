package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Stop stops the specified services (like docker stop).
func (p *Project) Stop(ctx context.Context, timeout int, services ...string) error {
	return p.perform(events.ProjectStopStart, events.ProjectStopDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServiceStopStart, events.ServiceStop, func(service Service) error {
			return service.Stop(ctx, timeout)
		})
	}), nil)
}
