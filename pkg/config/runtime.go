package config

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/containous/traefik/pkg/log"
)

// RuntimeConfiguration holds the information about the currently running traefik instance.
type RuntimeConfiguration struct {
	Routers     map[string]*RouterInfo     `json:"routers,omitempty"`
	Middlewares map[string]*MiddlewareInfo `json:"middlewares,omitempty"`
	Services    map[string]*ServiceInfo    `json:"services,omitempty"`
	TCPRouters  map[string]*TCPRouterInfo  `json:"tcpRouters,omitempty"`
	TCPServices map[string]*TCPServiceInfo `json:"tcpServices,omitempty"`
}

// NewRuntimeConfig returns a RuntimeConfiguration initialized with the given conf. It never returns nil.
func NewRuntimeConfig(conf Configuration) *RuntimeConfiguration {
	if conf.HTTP == nil && conf.TCP == nil {
		return &RuntimeConfiguration{}
	}

	runtimeConfig := &RuntimeConfiguration{}

	if conf.HTTP != nil {
		routers := conf.HTTP.Routers
		if len(routers) > 0 {
			runtimeConfig.Routers = make(map[string]*RouterInfo, len(routers))
			for k, v := range routers {
				runtimeConfig.Routers[k] = &RouterInfo{Router: v}
			}
		}

		services := conf.HTTP.Services
		if len(services) > 0 {
			runtimeConfig.Services = make(map[string]*ServiceInfo, len(services))
			for k, v := range services {
				runtimeConfig.Services[k] = &ServiceInfo{Service: v}
			}
		}

		middlewares := conf.HTTP.Middlewares
		if len(middlewares) > 0 {
			runtimeConfig.Middlewares = make(map[string]*MiddlewareInfo, len(middlewares))
			for k, v := range middlewares {
				runtimeConfig.Middlewares[k] = &MiddlewareInfo{Middleware: v}
			}
		}
	}

	if conf.TCP != nil {
		if len(conf.TCP.Routers) > 0 {
			runtimeConfig.TCPRouters = make(map[string]*TCPRouterInfo, len(conf.TCP.Routers))
			for k, v := range conf.TCP.Routers {
				runtimeConfig.TCPRouters[k] = &TCPRouterInfo{TCPRouter: v}
			}
		}

		if len(conf.TCP.Services) > 0 {
			runtimeConfig.TCPServices = make(map[string]*TCPServiceInfo, len(conf.TCP.Services))
			for k, v := range conf.TCP.Services {
				runtimeConfig.TCPServices[k] = &TCPServiceInfo{TCPService: v}
			}
		}
	}

	return runtimeConfig
}

// PopulateUsedBy populates all the UsedBy lists of the underlying fields of r,
// based on the relations between the included services, routers, and middlewares.
func (r *RuntimeConfiguration) PopulateUsedBy() {
	if r == nil {
		return
	}

	logger := log.WithoutContext()

	for routerName, routerInfo := range r.Routers {
		providerName := getProviderName(routerName)
		if providerName == "" {
			logger.WithField(log.RouterName, routerName).Error("router name is not fully qualified")
			continue
		}

		for _, midName := range routerInfo.Router.Middlewares {
			fullMidName := getQualifiedName(providerName, midName)
			if _, ok := r.Middlewares[fullMidName]; !ok {
				continue
			}
			r.Middlewares[fullMidName].UsedBy = append(r.Middlewares[fullMidName].UsedBy, routerName)
		}

		serviceName := getQualifiedName(providerName, routerInfo.Router.Service)
		if _, ok := r.Services[serviceName]; !ok {
			continue
		}
		r.Services[serviceName].UsedBy = append(r.Services[serviceName].UsedBy, routerName)
	}

	for k := range r.Services {
		sort.Strings(r.Services[k].UsedBy)
	}

	for k := range r.Middlewares {
		sort.Strings(r.Middlewares[k].UsedBy)
	}

	for routerName, routerInfo := range r.TCPRouters {
		providerName := getProviderName(routerName)
		if providerName == "" {
			logger.WithField(log.RouterName, routerName).Error("tcp router name is not fully qualified")
			continue
		}

		serviceName := getQualifiedName(providerName, routerInfo.TCPRouter.Service)
		if _, ok := r.TCPServices[serviceName]; !ok {
			continue
		}
		r.TCPServices[serviceName].UsedBy = append(r.TCPServices[serviceName].UsedBy, routerName)
	}

	for k := range r.TCPServices {
		sort.Strings(r.TCPServices[k].UsedBy)
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

// GetRoutersByEntrypoints returns all the http routers by entrypoints name and routers name
func (r *RuntimeConfiguration) GetRoutersByEntrypoints(ctx context.Context, entryPoints []string, tls bool) map[string]map[string]*RouterInfo {
	entryPointsRouters := make(map[string]map[string]*RouterInfo)

	for rtName, rt := range r.Routers {
		if (tls && rt.TLS == nil) || (!tls && rt.TLS != nil) {
			continue
		}

		eps := rt.EntryPoints
		if len(eps) == 0 {
			eps = entryPoints
		}
		for _, entryPointName := range eps {
			if !contains(entryPoints, entryPointName) {
				log.FromContext(log.With(ctx, log.Str(log.EntryPointName, entryPointName))).
					Errorf("entryPoint %q doesn't exist", entryPointName)
				continue
			}

			if _, ok := entryPointsRouters[entryPointName]; !ok {
				entryPointsRouters[entryPointName] = make(map[string]*RouterInfo)
			}

			entryPointsRouters[entryPointName][rtName] = rt
		}
	}

	return entryPointsRouters
}

// GetTCPRoutersByEntrypoints returns all the tcp routers by entrypoints name and routers name
func (r *RuntimeConfiguration) GetTCPRoutersByEntrypoints(ctx context.Context, entryPoints []string) map[string]map[string]*TCPRouterInfo {
	entryPointsRouters := make(map[string]map[string]*TCPRouterInfo)

	for rtName, rt := range r.TCPRouters {
		eps := rt.EntryPoints
		if len(eps) == 0 {
			eps = entryPoints
		}

		for _, entryPointName := range eps {
			if !contains(entryPoints, entryPointName) {
				log.FromContext(log.With(ctx, log.Str(log.EntryPointName, entryPointName))).
					Errorf("entryPoint %q doesn't exist", entryPointName)
				continue
			}

			if _, ok := entryPointsRouters[entryPointName]; !ok {
				entryPointsRouters[entryPointName] = make(map[string]*TCPRouterInfo)
			}

			entryPointsRouters[entryPointName][rtName] = rt
		}
	}

	return entryPointsRouters
}

// RouterInfo holds information about a currently running HTTP router
type RouterInfo struct {
	*Router        // dynamic configuration
	Err     string `json:"error,omitempty"` // initialization error
}

// TCPRouterInfo holds information about a currently running TCP router
type TCPRouterInfo struct {
	*TCPRouter        // dynamic configuration
	Err        string `json:"error,omitempty"` // initialization error
}

// MiddlewareInfo holds information about a currently running middleware
type MiddlewareInfo struct {
	*Middleware          // dynamic configuration
	Err         error    `json:"error,omitempty"`  // initialization error
	UsedBy      []string `json:"usedBy,omitempty"` // list of routers and services using that middleware
}

// ServiceInfo holds information about a currently running service
type ServiceInfo struct {
	*Service          // dynamic configuration
	Err      error    `json:"error,omitempty"`  // initialization error
	UsedBy   []string `json:"usedBy,omitempty"` // list of routers using that service

	statusMu sync.RWMutex
	status   map[string]string // keyed by server URL
}

// UpdateStatus sets the status of the server in the ServiceInfo.
// It is the responsibility of the caller to check that s is not nil.
func (s *ServiceInfo) UpdateStatus(server string, status string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	if s.status == nil {
		s.status = make(map[string]string)
	}
	s.status[server] = status
}

// GetAllStatus returns all the statuses of all the servers in ServiceInfo.
// It is the responsibility of the caller to check that s is not nil
func (s *ServiceInfo) GetAllStatus() map[string]string {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()

	if len(s.status) == 0 {
		return nil
	}

	allStatus := make(map[string]string, len(s.status))
	for k, v := range s.status {
		allStatus[k] = v
	}
	return allStatus
}

// TCPServiceInfo holds information about a currently running TCP service
type TCPServiceInfo struct {
	*TCPService          // dynamic configuration
	Err         error    `json:"error,omitempty"`  // initialization error
	UsedBy      []string `json:"usedBy,omitempty"` // list of routers using that service
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
