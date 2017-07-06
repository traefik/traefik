package project

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
)

// Down stops the specified services and clean related containers (like docker stop + docker rm).
func (p *Project) Down(ctx context.Context, opts options.Down, services ...string) error {
	if !opts.RemoveImages.Valid() {
		return fmt.Errorf("--rmi flag must be local, all or empty")
	}
	if err := p.Stop(ctx, 10, services...); err != nil {
		return err
	}
	if opts.RemoveOrphans {
		if err := p.runtime.RemoveOrphans(ctx, p.Name, p.ServiceConfigs); err != nil {
			return err
		}
	}
	if err := p.Delete(ctx, options.Delete{
		RemoveVolume: opts.RemoveVolume,
	}, services...); err != nil {
		return err
	}

	networks, err := p.context.NetworksFactory.Create(p.Name, p.NetworkConfigs, p.ServiceConfigs, p.isNetworkEnabled())
	if err != nil {
		return err
	}
	if err := networks.Remove(ctx); err != nil {
		return err
	}

	if opts.RemoveVolume {
		volumes, err := p.context.VolumesFactory.Create(p.Name, p.VolumeConfigs, p.ServiceConfigs, p.isVolumeEnabled())
		if err != nil {
			return err
		}
		if err := volumes.Remove(ctx); err != nil {
			return err
		}
	}

	return p.forEach([]string{}, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.NoEvent, events.NoEvent, func(service Service) error {
			return service.RemoveImage(ctx, opts.RemoveImages)
		})
	}), func(service Service) error {
		return service.Create(ctx, options.Create{})
	})
}
