package eureka

import (
	"strconv"
	"text/template"

	"github.com/ArthurHlt/go-eureka-client/eureka"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// Build the configuration from Provider server
func (p *Provider) buildConfiguration(apps *eureka.Applications) (*types.Configuration, error) {
	var eurekaFuncMap = template.FuncMap{
		"getPort":       getPort,
		"getProtocol":   getProtocol,
		"getWeight":     getWeight,
		"getInstanceID": getInstanceID,
	}

	templateObjects := struct {
		Applications []eureka.Application
	}{
		Applications: apps.Applications,
	}

	configuration, err := p.GetConfiguration("templates/eureka.tmpl", eurekaFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration, nil
}

func getInstanceID(instance eureka.InstanceInfo) string {
	defaultID := provider.Normalize(instance.IpAddr) + "-" + getPort(instance)
	return label.GetStringValue(instance.Metadata.Map, label.TraefikBackendID, defaultID)
}

func getPort(instance eureka.InstanceInfo) string {
	if instance.SecurePort.Enabled {
		return strconv.Itoa(instance.SecurePort.Port)
	}
	return strconv.Itoa(instance.Port.Port)
}

func getProtocol(instance eureka.InstanceInfo) string {
	if instance.SecurePort.Enabled {
		return "https"
	}
	return label.DefaultProtocol
}

func getWeight(instance eureka.InstanceInfo) int {
	return label.GetIntValue(instance.Metadata.Map, label.TraefikWeight, label.DefaultWeight)
}
