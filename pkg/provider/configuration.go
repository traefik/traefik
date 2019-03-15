package provider

import (
	"bytes"
	"context"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
)

// Merge Merges multiple configurations.
func Merge(ctx context.Context, configurations map[string]*config.Configuration) *config.Configuration {
	logger := log.FromContext(ctx)

	configuration := &config.Configuration{
		HTTP: &config.HTTPConfiguration{
			Routers:     make(map[string]*config.Router),
			Middlewares: make(map[string]*config.Middleware),
			Services:    make(map[string]*config.Service),
		},
		TCP: &config.TCPConfiguration{
			Routers:  make(map[string]*config.TCPRouter),
			Services: make(map[string]*config.TCPService),
		},
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
		for serviceName, service := range conf.HTTP.Services {
			services[serviceName] = append(services[serviceName], root)
			if !AddService(configuration.HTTP, serviceName, service) {
				servicesToDelete[serviceName] = struct{}{}
			}
		}

		for routerName, router := range conf.HTTP.Routers {
			routers[routerName] = append(routers[routerName], root)
			if !AddRouter(configuration.HTTP, routerName, router) {
				routersToDelete[routerName] = struct{}{}
			}
		}

		for middlewareName, middleware := range conf.HTTP.Middlewares {
			middlewares[middlewareName] = append(middlewares[middlewareName], root)
			if !AddMiddleware(configuration.HTTP, middlewareName, middleware) {
				middlewaresToDelete[middlewareName] = struct{}{}
			}
		}
	}

	for serviceName := range servicesToDelete {
		logger.WithField(log.ServiceName, serviceName).
			Errorf("Service defined multiple times with different configurations in %v", services[serviceName])
		delete(configuration.HTTP.Services, serviceName)
	}

	for routerName := range routersToDelete {
		logger.WithField(log.RouterName, routerName).
			Errorf("Router defined multiple times with different configurations in %v", routers[routerName])
		delete(configuration.HTTP.Routers, routerName)
	}

	for middlewareName := range middlewaresToDelete {
		logger.WithField(log.MiddlewareName, middlewareName).
			Errorf("Middleware defined multiple times with different configurations in %v", middlewares[middlewareName])
		delete(configuration.HTTP.Middlewares, middlewareName)
	}

	return configuration
}

// AddService Adds a service to a configurations.
func AddService(configuration *config.HTTPConfiguration, serviceName string, service *config.Service) bool {
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
func AddRouter(configuration *config.HTTPConfiguration, routerName string, router *config.Router) bool {
	if _, ok := configuration.Routers[routerName]; !ok {
		configuration.Routers[routerName] = router
		return true
	}

	return reflect.DeepEqual(configuration.Routers[routerName], router)
}

// AddMiddleware Adds a middleware to a configurations.
func AddMiddleware(configuration *config.HTTPConfiguration, middlewareName string, middleware *config.Middleware) bool {
	if _, ok := configuration.Middlewares[middlewareName]; !ok {
		configuration.Middlewares[middlewareName] = middleware
		return true
	}

	return reflect.DeepEqual(configuration.Middlewares[middlewareName], middleware)
}

// MakeDefaultRuleTemplate Creates the default rule template.
func MakeDefaultRuleTemplate(defaultRule string, funcMap template.FuncMap) (*template.Template, error) {
	defaultFuncMap := sprig.TxtFuncMap()
	defaultFuncMap["normalize"] = Normalize

	for k, fn := range funcMap {
		defaultFuncMap[k] = fn
	}

	return template.New("defaultRule").Funcs(defaultFuncMap).Parse(defaultRule)
}

// BuildRouterConfiguration Builds a router configuration.
func BuildRouterConfiguration(ctx context.Context, configuration *config.HTTPConfiguration, defaultRouterName string, defaultRuleTpl *template.Template, model interface{}) {
	logger := log.FromContext(ctx)

	if len(configuration.Routers) == 0 {
		if len(configuration.Services) > 1 {
			log.FromContext(ctx).Info("Could not create a router for the container: too many services")
		} else {
			configuration.Routers = make(map[string]*config.Router)
			configuration.Routers[defaultRouterName] = &config.Router{}
		}
	}

	for routerName, router := range configuration.Routers {
		loggerRouter := logger.WithField(log.RouterName, routerName)
		if len(router.Rule) == 0 {
			writer := &bytes.Buffer{}
			if err := defaultRuleTpl.Execute(writer, model); err != nil {
				loggerRouter.Errorf("Error while parsing default rule: %v", err)
				delete(configuration.Routers, routerName)
				continue
			}

			router.Rule = writer.String()
			if len(router.Rule) == 0 {
				loggerRouter.Error("Undefined rule")
				delete(configuration.Routers, routerName)
				continue
			}
		}

		if len(router.Service) == 0 {
			if len(configuration.Services) > 1 {
				delete(configuration.Routers, routerName)
				loggerRouter.
					Error("Could not define the service name for the router: too many services")
				continue
			}

			for serviceName := range configuration.Services {
				router.Service = serviceName
			}
		}
	}
}

// Normalize Replace all special chars with `-`.
func Normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}
