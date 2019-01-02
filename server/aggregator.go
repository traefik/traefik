package server

import (
	"fmt"

	"github.com/containous/traefik/config"
)

func mergeConfiguration(configurations config.Configurations) config.Configuration {
	conf := config.Configuration{
		Routers:     make(map[string]*config.Router),
		Middlewares: make(map[string]*config.Middleware),
		Services:    make(map[string]*config.Service),
	}

	for provider, configuration := range configurations {
		for routerName, router := range configuration.Routers {
			conf.Routers[getQualifiedName(provider, routerName)] = router
		}
		for middlewareName, middleware := range configuration.Middlewares {
			conf.Middlewares[getQualifiedName(provider, middlewareName)] = middleware
		}
		for serviceName, service := range configuration.Services {
			conf.Services[getQualifiedName(provider, serviceName)] = service
		}
		conf.TLS = append(conf.TLS, configuration.TLS...)
	}

	return conf
}

func getQualifiedName(provider string, element string) string {
	return fmt.Sprintf("%s.%s", provider, element)
}
