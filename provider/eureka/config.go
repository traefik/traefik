package eureka

import (
	"io/ioutil"
	"strconv"
	"text/template"

	"github.com/ArthurHlt/go-eureka-client/eureka"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// Build the configuration from Provider server
func (p *Provider) buildConfiguration() (*types.Configuration, error) {
	var EurekaFuncMap = template.FuncMap{
		"getPort":       getPort,
		"getProtocol":   getProtocol,
		"getWeight":     getWeight,
		"getInstanceID": getInstanceID,
	}

	eureka.GetLogger().SetOutput(ioutil.Discard)

	client := eureka.NewClient([]string{
		p.Endpoint,
	})

	applications, err := client.GetApplications()
	if err != nil {
		return nil, err
	}

	templateObjects := struct {
		Applications []eureka.Application
	}{
		applications.Applications,
	}

	configuration, err := p.GetConfiguration("templates/eureka.tmpl", EurekaFuncMap, templateObjects)
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

func getWeight(instance eureka.InstanceInfo) string {
	return label.GetStringValue(instance.Metadata.Map, label.TraefikWeight, label.DefaultWeight)
}
