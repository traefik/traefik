package project

import "golang.org/x/net/context"

// Ps list containers for the specified services.
func (p *Project) Ps(ctx context.Context, services ...string) (InfoSet, error) {
	allInfo := InfoSet{}

	if len(services) == 0 {
		services = p.ServiceConfigs.Keys()
	}

	for _, name := range services {

		service, err := p.CreateService(name)
		if err != nil {
			return nil, err
		}

		info, err := service.Info(ctx)
		if err != nil {
			return nil, err
		}

		allInfo = append(allInfo, info...)
	}
	return allInfo, nil
}
