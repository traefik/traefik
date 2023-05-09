package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	metricsMiddle "github.com/traefik/traefik/v3/pkg/middlewares/metrics"
	tracingMiddle "github.com/traefik/traefik/v3/pkg/middlewares/tracing"
	"go.opentelemetry.io/otel/trace"
)

// ObservabilityMgr is a manager for observability (AccessLogs, Metrics and Tracing) enablement.
type ObservabilityMgr struct {
	config                 static.Observability
	accessLoggerMiddleware *accesslog.Handler
	metricsRegistry        metrics.Registry
	tracer                 trace.Tracer
}

// NewObservabilityMgr creates a new ObservabilityMgr.
func NewObservabilityMgr(config static.Observability, metricsRegistry metrics.Registry, accessLoggerMiddleware *accesslog.Handler, tracer trace.Tracer) *ObservabilityMgr {
	return &ObservabilityMgr{
		config:                 config,
		metricsRegistry:        metricsRegistry,
		accessLoggerMiddleware: accessLoggerMiddleware,
		tracer:                 tracer,
	}
}

// BuildEPChain an observability middleware chain by entry point.
func (c *ObservabilityMgr) BuildEPChain(ctx context.Context, entryPointName string) alice.Chain {
	chain := alice.New()

	if c == nil {
		return chain
	}

	if c.accessLoggerMiddleware != nil || c.metricsRegistry != nil && (c.metricsRegistry.IsEpEnabled() || c.metricsRegistry.IsRouterEnabled() || c.metricsRegistry.IsSvcEnabled()) {
		chain = chain.Append(capture.Wrap)
	}

	if c.accessLoggerMiddleware != nil {
		chain = chain.Append(accesslog.WrapHandler(c.accessLoggerMiddleware))
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return accesslog.NewFieldHandler(next, logs.EntryPointName, entryPointName, accesslog.InitServiceFields), nil
		})
	}

	if c.tracer != nil {
		chain = chain.Append(tracingMiddle.WrapEntryPointHandler(ctx, c.tracer, entryPointName))
	}

	if c.metricsRegistry != nil && c.metricsRegistry.IsEpEnabled() {
		metricsHandler := metricsMiddle.WrapEntryPointHandler(ctx, c.metricsRegistry, entryPointName)
		chain = chain.Append(tracingMiddle.WrapMiddleware(ctx, metricsHandler))
	}

	return chain
}

// ShouldObserve returns whether the observability should be enabled for the given resource.
func (c *ObservabilityMgr) ShouldObserve(resourceName string) bool {
	if c == nil {
		return false
	}

	return c.config.AddInternals || !strings.HasSuffix(resourceName, "@internal")
}

// MetricsRegistry is an accessor to the metrics registry.
func (c *ObservabilityMgr) MetricsRegistry() metrics.Registry {
	if c == nil {
		return nil
	}

	return c.metricsRegistry
}

// Close closes the accessLogger and tracer.
func (c *ObservabilityMgr) Close() {
	if c == nil {
		return
	}

	if c.accessLoggerMiddleware != nil {
		if err := c.accessLoggerMiddleware.Close(); err != nil {
			log.Error().Err(err).Msg("Could not close the access log file")
		}
	}
}
