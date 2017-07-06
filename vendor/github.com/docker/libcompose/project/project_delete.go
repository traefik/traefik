package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
)

// Delete removes the specified services (like docker rm).
func (p *Project) Delete(ctx context.Context, options options.Delete, services ...string) error {
	return p.perform(events.ProjectDeleteStart, events.ProjectDeleteDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServiceDeleteStart, events.ServiceDelete, func(service Service) error {
			return service.Delete(ctx, options)
		})
	}), nil)
}
