package server

import (
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/server/internal"
)

func mergeConfiguration(configurations config.Configurations) config.Configuration {
	conf := config.Configuration{
		Routers:     make(map[string]*config.Router),
		Middlewares: make(map[string]*config.Middleware),
		Services:    make(map[string]*config.Service),
	}

	for provider, configuration := range configurations {
		for routerName, router := range configuration.Routers {
			conf.Routers[internal.MakeQualifiedName(provider, routerName)] = router
		}
		for middlewareName, middleware := range configuration.Middlewares {
			conf.Middlewares[internal.MakeQualifiedName(provider, middlewareName)] = middleware
		}
		for serviceName, service := range configuration.Services {
			conf.Services[internal.MakeQualifiedName(provider, serviceName)] = service
		}
		conf.TLS = append(conf.TLS, configuration.TLS...)
	}

	return conf
}
