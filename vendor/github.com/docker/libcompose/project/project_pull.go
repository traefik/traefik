package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Pull pulls the specified services (like docker pull).
func (p *Project) Pull(ctx context.Context, services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServicePullStart, events.ServicePull, func(service Service) error {
			return service.Pull(ctx)
		})
	}), nil)
}
