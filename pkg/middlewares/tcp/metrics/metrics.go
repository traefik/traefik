package metrics

import (
	"context"

	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

const (
	typeName   = "MetricsTCP"
	nameRouter = "metrics-router"
)

type metricsMiddleware struct {
	next                 tcp.Handler
	openConnectionsGauge gokitmetrics.Gauge
}

// NewRouterMiddleware creates a new metrics middleware for a TCP Router.
func NewRouterMiddleware(ctx context.Context, next tcp.Handler, registry metrics.Registry, routerName string, serviceName string) tcp.Handler {
	middlewares.GetLogger(ctx, nameRouter, typeName).Debug().Msg("Creating middleware")

	return &metricsMiddleware{
		next:                 next,
		openConnectionsGauge: registry.RouterTCPOpenConnectionsGauge().With("router", routerName, "service", serviceName),
	}
}

// RouterMetricsHandler returns the metrics TCP router handler.
func RouterMetricsHandler(ctx context.Context, registry metrics.Registry, routerName string, serviceName string) tcp.Constructor {
	return func(next tcp.Handler) (tcp.Handler, error) {
		if registry == nil || !registry.IsTCPRouterEnabled() {
			return next, nil
		}

		return NewRouterMiddleware(ctx, next, registry, routerName, serviceName), nil
	}
}

// ServeTCP serves the given TCP connection.
func (m *metricsMiddleware) ServeTCP(conn tcp.WriteCloser) {
	m.openConnectionsGauge.Add(1)
	defer m.openConnectionsGauge.Add(-1)

	m.next.ServeTCP(conn)
}
