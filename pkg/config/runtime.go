package config

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/containous/traefik/pkg/log"
)

// TODO(mpl): omitempty on all potential nils, for nicer endpoint outputs?

// RuntimeConfiguration holds the information about the currently running traefik instance.
type RuntimeConfiguration struct {
	Routers     map[string]*RouterInfo
	Middlewares map[string]*MiddlewareInfo
	Services    map[string]*ServiceInfo
	TCPRouters  map[string]*TCPRouterInfo
	TCPServices map[string]*TCPServiceInfo
}

// NewRuntimeConfig returns a RuntimeConfiguration initialized with the given conf. It never returns nil.
func NewRuntimeConfig(conf Configuration) *RuntimeConfiguration {
	if conf.HTTP == nil && conf.TCP == nil {
		return &RuntimeConfiguration{}
	}
	var (
		mI    map[string]*MiddlewareInfo
		rI    map[string]*RouterInfo
		sI    map[string]*ServiceInfo
		rITCP map[string]*TCPRouterInfo
		sITCP map[string]*TCPServiceInfo
	)
	if conf.HTTP != nil {
		routers := conf.HTTP.Routers
		services := conf.HTTP.Services
		middlewares := conf.HTTP.Middlewares
		rI = make(map[string]*RouterInfo, len(routers))
		for k, v := range routers {
			rI[k] = &RouterInfo{Router: v}
		}
		sI = make(map[string]*ServiceInfo, len(services))
		for k, v := range services {
			sI[k] = &ServiceInfo{Service: v}
		}
		mI = make(map[string]*MiddlewareInfo, len(middlewares))
		for k, v := range middlewares {
			mI[k] = &MiddlewareInfo{Middleware: v}
		}
	}
	if conf.TCP != nil {
		routers := conf.TCP.Routers
		services := conf.TCP.Services
		rITCP = make(map[string]*TCPRouterInfo, len(routers))
		for k, v := range routers {
			rITCP[k] = &TCPRouterInfo{TCPRouter: v}
		}
		sITCP = make(map[string]*TCPServiceInfo, len(services))
		for k, v := range services {
			sITCP[k] = &TCPServiceInfo{TCPService: v}
		}
	}
	if len(sI) == 0 && len(rI) == 0 && len(mI) == 0 &&
		len(rITCP) == 0 && len(sITCP) == 0 {
		return &RuntimeConfiguration{}
	}
	return &RuntimeConfiguration{
		Services:    sI,
		Routers:     rI,
		Middlewares: mI,
		TCPRouters:  rITCP,
		TCPServices: sITCP,
	}
}

// RouterInfo holds information about a currently running HTTP router
type RouterInfo struct {
	*Router        // dynamic configuration
	Err     string // initialization error
}

// TCPRouterInfo holds information about a currently running TCP router
type TCPRouterInfo struct {
	*TCPRouter        // dynamic configuration
	Err        string // initialization error
}

// MiddlewareInfo holds information about a currently running middleware
type MiddlewareInfo struct {
	*Middleware          // dynamic configuration
	Err         error    // initialization error
	UsedBy      []string // list of routers and services using that middleware
}

// ServiceInfo holds information about a currently running service
type ServiceInfo struct {
	*Service          // dynamic configuration
	Err      error    // initialization error
	UsedBy   []string // list of routers using that service

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
	var allStatus map[string]string
	if len(s.status) > 0 {
		allStatus = make(map[string]string, len(s.status))
	}
	for k, v := range s.status {
		allStatus[k] = v
	}
	return allStatus
}

// TCPServiceInfo holds information about a currently running TCP service
type TCPServiceInfo struct {
	*TCPService          // dynamic configuration
	Err         error    // initialization error
	UsedBy      []string // list of routers using that service
}

func getProviderName(elementName string) string {
	parts := strings.Split(elementName, ".")
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}

func getQualifiedName(provider, elementName string) string {
	parts := strings.Split(elementName, ".")
	if len(parts) == 1 {
		return provider + "." + elementName
	}
	return elementName
}

// PopulateUsedBy populates all the UsedBy lists of the underlying fields of r,
// based on the relations between the included services, routers, and middlewares.
func (r *RuntimeConfiguration) PopulateUsedBy() {
	if r == nil {
		return
	}
	routerInfos := r.Routers
	serviceInfos := r.Services
	middlewareInfos := r.Middlewares
	for routerName, routerInf := range routerInfos {
		providerName := getProviderName(routerName)
		if providerName == "" {
			log.FromContext(context.Background()).Errorf("router name is not fully qualified: %q", routerName)
			continue
		}

		for _, midName := range routerInf.Router.Middlewares {
			fullMidName := getQualifiedName(providerName, midName)
			if _, ok := middlewareInfos[fullMidName]; !ok {
				continue
			}
			middlewareInfos[fullMidName].UsedBy = append(middlewareInfos[fullMidName].UsedBy, routerName)
		}

		serviceName := getQualifiedName(providerName, routerInf.Router.Service)
		if _, ok := serviceInfos[serviceName]; !ok {
			continue
		}
		serviceInfos[serviceName].UsedBy = append(serviceInfos[serviceName].UsedBy, routerName)
	}
	for k := range serviceInfos {
		sort.Strings(serviceInfos[k].UsedBy)
	}
	for k := range middlewareInfos {
		sort.Strings(middlewareInfos[k].UsedBy)
	}

	tcpServiceInfos := r.TCPServices
	tcpRouterInfos := r.TCPRouters
	for routerName, routerInf := range tcpRouterInfos {
		providerName := getProviderName(routerName)
		if providerName == "" {
			log.FromContext(context.Background()).Errorf("tcp router name is not fully qualified: %q", routerName)
			continue
		}
		serviceName := getQualifiedName(providerName, routerInf.TCPRouter.Service)
		if _, ok := tcpServiceInfos[serviceName]; !ok {
			continue
		}
		tcpServiceInfos[serviceName].UsedBy = append(tcpServiceInfos[serviceName].UsedBy, routerName)
	}
	for k := range tcpServiceInfos {
		sort.Strings(tcpServiceInfos[k].UsedBy)
	}
}
