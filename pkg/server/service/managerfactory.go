package service

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/traefik/traefik/v2/pkg/api"
	"github.com/traefik/traefik/v2/pkg/api/dashboard"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/proxy"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/tls/client"
)

// ManagerFactory a factory of service manager.
type ManagerFactory struct {
	metricsRegistry metrics.Registry

	proxyBuilder           *proxy.Builder
	tlsClientConfigManager *client.TLSConfigManager

	api              func(configuration *runtime.Configuration) http.Handler
	restHandler      http.Handler
	dashboardHandler http.Handler
	metricsHandler   http.Handler
	pingHandler      http.Handler
	acmeHTTPHandler  http.Handler

	routinesPool *safe.Pool
}

// NewManagerFactory creates a new ManagerFactory.
func NewManagerFactory(staticConfiguration static.Configuration, routinesPool *safe.Pool, metricsRegistry metrics.Registry, proxyBuilder *proxy.Builder, tlsClientConfigManager *client.TLSConfigManager, acmeHTTPHandler http.Handler) *ManagerFactory {
	factory := &ManagerFactory{
		metricsRegistry:        metricsRegistry,
		routinesPool:           routinesPool,
		proxyBuilder:           proxyBuilder,
		tlsClientConfigManager: tlsClientConfigManager,
		acmeHTTPHandler:        acmeHTTPHandler,
	}

	if staticConfiguration.API != nil {
		apiRouterBuilder := api.NewBuilder(staticConfiguration)

		if staticConfiguration.API.Dashboard {
			factory.dashboardHandler = dashboard.Handler{}
			factory.api = func(configuration *runtime.Configuration) http.Handler {
				router := apiRouterBuilder(configuration).(*mux.Router)
				dashboard.Append(router, nil)
				return router
			}
		} else {
			factory.api = apiRouterBuilder
		}
	}

	if staticConfiguration.Providers != nil && staticConfiguration.Providers.Rest != nil {
		factory.restHandler = staticConfiguration.Providers.Rest.CreateRouter()
	}

	if staticConfiguration.Metrics != nil && staticConfiguration.Metrics.Prometheus != nil {
		factory.metricsHandler = metrics.PrometheusHandler()
	}

	// This check is necessary because even when staticConfiguration.Ping == nil ,
	// the affectation would make factory.pingHandle become a typed nil, which does not pass the nil test,
	// and would break things elsewhere.
	if staticConfiguration.Ping != nil {
		factory.pingHandler = staticConfiguration.Ping
	}

	return factory
}

// Build creates a service manager.
func (f *ManagerFactory) Build(configuration *runtime.Configuration) *InternalHandlers {
	svcManager := NewManager(configuration.Services, f.metricsRegistry, f.routinesPool, f.proxyBuilder, f.tlsClientConfigManager)

	var apiHandler http.Handler
	if f.api != nil {
		apiHandler = f.api(configuration)
	}

	return NewInternalHandlers(svcManager, apiHandler, f.restHandler, f.metricsHandler, f.pingHandler, f.dashboardHandler, f.acmeHTTPHandler)
}
