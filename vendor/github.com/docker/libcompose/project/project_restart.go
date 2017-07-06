package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Restart restarts the specified services (like docker restart).
func (p *Project) Restart(ctx context.Context, timeout int, services ...string) error {
	return p.perform(events.ProjectRestartStart, events.ProjectRestartDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceRestartStart, events.ServiceRestart, func(service Service) error {
			return service.Restart(ctx, timeout)
		})
	}), nil)
}
