package project

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
)

// Containers lists the containers for the specified services. Can be filter using
// the Filter struct.
func (p *Project) Containers(ctx context.Context, filter Filter, services ...string) ([]string, error) {
	containers := []string{}
	var lock sync.Mutex

	err := p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.NoEvent, events.NoEvent, func(service Service) error {
			serviceContainers, innerErr := service.Containers(ctx)
			if innerErr != nil {
				return innerErr
			}

			for _, container := range serviceContainers {
				running := container.IsRunning(ctx)
				switch filter.State {
				case Running:
					if !running {
						continue
					}
				case Stopped:
					if running {
						continue
					}
				case AnyState:
					// Don't do a thing
				default:
					// Invalid state filter
					return fmt.Errorf("Invalid container filter: %s", filter.State)
				}
				containerID := container.ID()
				lock.Lock()
				containers = append(containers, containerID)
				lock.Unlock()
			}
			return nil
		})
	}), nil)
	if err != nil {
		return nil, err
	}
	return containers, nil
}
