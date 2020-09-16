package service

import (
	"net/http"

	"github.com/traefik/traefik/v2/pkg/api"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/safe"
)

// ManagerFactory a factory of service manager.
type ManagerFactory struct {
	metricsRegistry metrics.Registry

	defaultRoundTripper http.RoundTripper

	api              func(configuration *runtime.Configuration) http.Handler
	restHandler      http.Handler
	dashboardHandler http.Handler
	metricsHandler   http.Handler
	pingHandler      http.Handler

	routinesPool *safe.Pool
}

// NewManagerFactory creates a new ManagerFactory.
func NewManagerFactory(staticConfiguration static.Configuration, routinesPool *safe.Pool, metricsRegistry metrics.Registry) *ManagerFactory {
	factory := &ManagerFactory{
		metricsRegistry:     metricsRegistry,
		defaultRoundTripper: setupDefaultRoundTripper(staticConfiguration.ServersTransport),
		routinesPool:        routinesPool,
	}

	if staticConfiguration.API != nil {
		factory.api = api.NewBuilder(staticConfiguration)

		if staticConfiguration.API.Dashboard {
			factory.dashboardHandler = http.FileServer(staticConfiguration.API.DashboardAssets)
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
	svcManager := NewManager(configuration.Services, f.defaultRoundTripper, f.metricsRegistry, f.routinesPool)
	return NewInternalHandlers(f.api, configuration, f.restHandler, f.metricsHandler, f.pingHandler, f.dashboardHandler, svcManager)
}
