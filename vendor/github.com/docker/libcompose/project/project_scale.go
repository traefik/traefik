package project

import (
	"fmt"

	"golang.org/x/net/context"

	log "github.com/sirupsen/logrus"
)

// Scale scales the specified services.
func (p *Project) Scale(ctx context.Context, timeout int, servicesScale map[string]int) error {
	// This code is a bit verbose but I wanted to parse everything up front
	order := make([]string, 0, 0)
	services := make(map[string]Service)

	for name := range servicesScale {
		if !p.ServiceConfigs.Has(name) {
			return fmt.Errorf("%s is not defined in the template", name)
		}

		service, err := p.CreateService(name)
		if err != nil {
			return fmt.Errorf("Failed to lookup service: %s: %v", service, err)
		}

		order = append(order, name)
		services[name] = service
	}

	for _, name := range order {
		scale := servicesScale[name]
		log.Infof("Setting scale %s=%d...", name, scale)
		err := services[name].Scale(ctx, scale, timeout)
		if err != nil {
			return fmt.Errorf("Failed to set the scale %s=%d: %v", name, scale, err)
		}
	}
	return nil
}
