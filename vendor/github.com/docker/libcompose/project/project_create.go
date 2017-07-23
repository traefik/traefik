package project

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
)

// Create creates the specified services (like docker create).
func (p *Project) Create(ctx context.Context, options options.Create, services ...string) error {
	if options.NoRecreate && options.ForceRecreate {
		return fmt.Errorf("no-recreate and force-recreate cannot be combined")
	}
	if err := p.initialize(ctx); err != nil {
		return err
	}
	return p.perform(events.ProjectCreateStart, events.ProjectCreateDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceCreateStart, events.ServiceCreate, func(service Service) error {
			return service.Create(ctx, options)
		})
	}), nil)
}
