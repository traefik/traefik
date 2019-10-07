package runtime

import (
	"sort"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
)

// Status of the router/service
const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"
	StatusWarning  = "warning"
)

// Configuration holds the information about the currently running traefik instance.
type Configuration struct {
	Routers     map[string]*RouterInfo     `json:"routers,omitempty"`
	Middlewares map[string]*MiddlewareInfo `json:"middlewares,omitempty"`
	Services    map[string]*ServiceInfo    `json:"services,omitempty"`
	TCPRouters  map[string]*TCPRouterInfo  `json:"tcpRouters,omitempty"`
	TCPServices map[string]*TCPServiceInfo `json:"tcpServices,omitempty"`
}

// NewConfig returns a Configuration initialized with the given conf. It never returns nil.
func NewConfig(conf dynamic.Configuration) *Configuration {
	if conf.HTTP == nil && conf.TCP == nil {
		return &Configuration{}
	}

	runtimeConfig := &Configuration{}

	if conf.HTTP != nil {
		routers := conf.HTTP.Routers
		if len(routers) > 0 {
			runtimeConfig.Routers = make(map[string]*RouterInfo, len(routers))
			for k, v := range routers {
				runtimeConfig.Routers[k] = &RouterInfo{Router: v, Status: StatusEnabled}
			}
		}

		services := conf.HTTP.Services
		if len(services) > 0 {
			runtimeConfig.Services = make(map[string]*ServiceInfo, len(services))
			for k, v := range services {
				runtimeConfig.Services[k] = &ServiceInfo{Service: v, Status: StatusEnabled}
			}
		}

		middlewares := conf.HTTP.Middlewares
		if len(middlewares) > 0 {
			runtimeConfig.Middlewares = make(map[string]*MiddlewareInfo, len(middlewares))
			for k, v := range middlewares {
				runtimeConfig.Middlewares[k] = &MiddlewareInfo{Middleware: v, Status: StatusEnabled}
			}
		}
	}

	if conf.TCP != nil {
		if len(conf.TCP.Routers) > 0 {
			runtimeConfig.TCPRouters = make(map[string]*TCPRouterInfo, len(conf.TCP.Routers))
			for k, v := range conf.TCP.Routers {
				runtimeConfig.TCPRouters[k] = &TCPRouterInfo{TCPRouter: v, Status: StatusEnabled}
			}
		}

		if len(conf.TCP.Services) > 0 {
			runtimeConfig.TCPServices = make(map[string]*TCPServiceInfo, len(conf.TCP.Services))
			for k, v := range conf.TCP.Services {
				runtimeConfig.TCPServices[k] = &TCPServiceInfo{TCPService: v, Status: StatusEnabled}
			}
		}
	}

	return runtimeConfig
}

// PopulateUsedBy populates all the UsedBy lists of the underlying fields of r,
// based on the relations between the included services, routers, and middlewares.
func (c *Configuration) PopulateUsedBy() {
	if c == nil {
		return
	}

	logger := log.WithoutContext()

	for routerName, routerInfo := range c.Routers {
		// lazily initialize Status in case caller forgot to do it
		if routerInfo.Status == "" {
			routerInfo.Status = StatusEnabled
		}

		providerName := getProviderName(routerName)
		if providerName == "" {
			logger.WithField(log.RouterName, routerName).Error("router name is not fully qualified")
			continue
		}

		for _, midName := range routerInfo.Router.Middlewares {
			fullMidName := getQualifiedName(providerName, midName)
			if _, ok := c.Middlewares[fullMidName]; !ok {
				continue
			}
			c.Middlewares[fullMidName].UsedBy = append(c.Middlewares[fullMidName].UsedBy, routerName)
		}

		serviceName := getQualifiedName(providerName, routerInfo.Router.Service)
		if _, ok := c.Services[serviceName]; !ok {
			continue
		}
		c.Services[serviceName].UsedBy = append(c.Services[serviceName].UsedBy, routerName)
	}

	for k, serviceInfo := range c.Services {
		// lazily initialize Status in case caller forgot to do it
		if serviceInfo.Status == "" {
			serviceInfo.Status = StatusEnabled
		}

		sort.Strings(c.Services[k].UsedBy)
	}

	for midName, mid := range c.Middlewares {
		// lazily initialize Status in case caller forgot to do it
		if mid.Status == "" {
			mid.Status = StatusEnabled
		}

		sort.Strings(c.Middlewares[midName].UsedBy)
	}

	for routerName, routerInfo := range c.TCPRouters {
		// lazily initialize Status in case caller forgot to do it
		if routerInfo.Status == "" {
			routerInfo.Status = StatusEnabled
		}

		providerName := getProviderName(routerName)
		if providerName == "" {
			logger.WithField(log.RouterName, routerName).Error("tcp router name is not fully qualified")
			continue
		}

		serviceName := getQualifiedName(providerName, routerInfo.TCPRouter.Service)
		if _, ok := c.TCPServices[serviceName]; !ok {
			continue
		}
		c.TCPServices[serviceName].UsedBy = append(c.TCPServices[serviceName].UsedBy, routerName)
	}

	for k, serviceInfo := range c.TCPServices {
		// lazily initialize Status in case caller forgot to do it
		if serviceInfo.Status == "" {
			serviceInfo.Status = StatusEnabled
		}

		sort.Strings(c.TCPServices[k].UsedBy)
	}
}

func contains(entryPoints []string, entryPointName string) bool {
	for _, name := range entryPoints {
		if name == entryPointName {
			return true
		}
	}
	return false
}

func getProviderName(elementName string) string {
	parts := strings.Split(elementName, "@")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func getQualifiedName(provider, elementName string) string {
	parts := strings.Split(elementName, "@")
	if len(parts) == 1 {
		return elementName + "@" + provider
	}
	return elementName
}
