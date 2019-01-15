package provider

import (
	"context"
	"reflect"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/log"
)

// MergeServices Merges services in global configuration.
func MergeServices(ctx context.Context, confFromLabel *config.Configuration, configuration *config.Configuration) {
	logger := log.FromContext(ctx)

	for serviceName, service := range confFromLabel.Services {
		_, ok := configuration.Services[serviceName]
		if !ok {
			configuration.Services[serviceName] = service
		} else {
			if configuration.Services[serviceName].LoadBalancer.Mergeable(service.LoadBalancer) {
				configuration.Services[serviceName].LoadBalancer.Servers = append(configuration.Services[serviceName].LoadBalancer.Servers, service.LoadBalancer.Servers...)
			} else {
				logger.WithField(log.ServiceName, serviceName).
					Error("Service defined two times with different configuration")
				delete(configuration.Services, serviceName)
			}
		}
	}
}

// MergeRouters Merges routers in global configuration.
func MergeRouters(ctx context.Context, confFromLabel *config.Configuration, configuration *config.Configuration) {
	logger := log.FromContext(ctx)

	for routerName, router := range confFromLabel.Routers {
		if _, ok := configuration.Routers[routerName]; !ok {
			configuration.Routers[routerName] = router
		} else if !reflect.DeepEqual(configuration.Routers[routerName], router) {
			logger.WithField(log.RouterName, routerName).
				Error("Router defined two times with different configuration")
			delete(configuration.Routers, routerName)
		}
	}
}

// MergeMiddlewares Merges middlewares in global configuration.
func MergeMiddlewares(ctx context.Context, confFromLabel *config.Configuration, configuration *config.Configuration) {
	logger := log.FromContext(ctx)

	for middlewareName, middleware := range confFromLabel.Middlewares {
		if _, ok := configuration.Middlewares[middlewareName]; !ok {
			configuration.Middlewares[middlewareName] = middleware
		} else if !reflect.DeepEqual(configuration.Middlewares[middlewareName], middleware) {
			logger.WithField(log.MiddlewareName, middlewareName).
				Error("Middleware defined two times with different configuration")
			delete(configuration.Middlewares, middlewareName)
		}
	}
}
