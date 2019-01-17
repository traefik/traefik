package provider

import (
	"context"
	"reflect"
	"sort"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/log"
)

// Merge Merges multiple configurations.
func Merge(ctx context.Context, configurations map[string]*config.Configuration) *config.Configuration {
	logger := log.FromContext(ctx)

	configuration := &config.Configuration{
		Routers:     make(map[string]*config.Router),
		Middlewares: make(map[string]*config.Middleware),
		Services:    make(map[string]*config.Service),
	}

	servicesToDelete := map[string]struct{}{}
	services := map[string][]string{}

	routersToDelete := map[string]struct{}{}
	routers := map[string][]string{}

	middlewaresToDelete := map[string]struct{}{}
	middlewares := map[string][]string{}

	var sortedKeys []string
	for key := range configurations {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	for _, root := range sortedKeys {
		conf := configurations[root]
		for serviceName, service := range conf.Services {
			services[serviceName] = append(services[serviceName], root)
			if !AddService(configuration, serviceName, service) {
				servicesToDelete[serviceName] = struct{}{}
			}
		}

		for routerName, router := range conf.Routers {
			routers[routerName] = append(routers[routerName], root)
			if !AddRouter(configuration, routerName, router) {
				routersToDelete[routerName] = struct{}{}
			}
		}

		for middlewareName, middleware := range conf.Middlewares {
			middlewares[middlewareName] = append(middlewares[middlewareName], root)
			if !AddMiddleware(configuration, middlewareName, middleware) {
				middlewaresToDelete[middlewareName] = struct{}{}
			}
		}
	}

	for serviceName := range servicesToDelete {
		logger.WithField(log.ServiceName, serviceName).
			Errorf("Service defined multiple times with different configuration in %v", services[serviceName])
		delete(configuration.Services, serviceName)
	}

	for routerName := range routersToDelete {
		logger.WithField(log.RouterName, routerName).
			Errorf("Router defined multiple times with different configuration in %v", routers[routerName])
		delete(configuration.Routers, routerName)
	}

	for middlewareName := range middlewaresToDelete {
		logger.WithField(log.MiddlewareName, middlewareName).
			Errorf("Middleware defined multiple times with different configuration in %v", middlewares[middlewareName])
		delete(configuration.Middlewares, middlewareName)
	}

	return configuration
}

// AddService Adds a service to a configurations.
func AddService(configuration *config.Configuration, serviceName string, service *config.Service) bool {
	if _, ok := configuration.Services[serviceName]; !ok {
		configuration.Services[serviceName] = service
		return true
	}

	if !configuration.Services[serviceName].LoadBalancer.Mergeable(service.LoadBalancer) {
		return false
	}

	configuration.Services[serviceName].LoadBalancer.Servers = append(configuration.Services[serviceName].LoadBalancer.Servers, service.LoadBalancer.Servers...)
	return true
}

// AddRouter Adds a router to a configurations.
func AddRouter(configuration *config.Configuration, routerName string, router *config.Router) bool {
	if _, ok := configuration.Routers[routerName]; !ok {
		configuration.Routers[routerName] = router
		return true
	}

	return reflect.DeepEqual(configuration.Routers[routerName], router)
}

// AddMiddleware Adds a middleware to a configurations.
func AddMiddleware(configuration *config.Configuration, middlewareName string, middleware *config.Middleware) bool {
	if _, ok := configuration.Middlewares[middlewareName]; !ok {
		configuration.Middlewares[middlewareName] = middleware
		return true
	}

	return reflect.DeepEqual(configuration.Middlewares[middlewareName], middleware)
}
