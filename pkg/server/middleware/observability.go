package middleware

import (
	"context"
	"io"
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
	config                 static.Configuration
	accessLoggerMiddleware *accesslog.Handler
	metricsRegistry        metrics.Registry
	tracer                 trace.Tracer
	tracerCloser           io.Closer
}

// NewObservabilityMgr creates a new ObservabilityMgr.
func NewObservabilityMgr(config static.Configuration, metricsRegistry metrics.Registry, accessLoggerMiddleware *accesslog.Handler, tracer trace.Tracer, tracerCloser io.Closer) *ObservabilityMgr {
	return &ObservabilityMgr{
		config:                 config,
		metricsRegistry:        metricsRegistry,
		accessLoggerMiddleware: accessLoggerMiddleware,
		tracer:                 tracer,
		tracerCloser:           tracerCloser,
	}
}

// BuildEPChain an observability middleware chain by entry point.
func (c *ObservabilityMgr) BuildEPChain(ctx context.Context, entryPointName string, resourceName string) alice.Chain {
	chain := alice.New()

	if c == nil {
		return chain
	}

	if c.accessLoggerMiddleware != nil || c.metricsRegistry != nil && (c.metricsRegistry.IsEpEnabled() || c.metricsRegistry.IsRouterEnabled() || c.metricsRegistry.IsSvcEnabled()) {
		if c.ShouldAddAccessLogs(resourceName) || c.ShouldAddMetrics(resourceName) {
			chain = chain.Append(capture.Wrap)
		}
	}

	if c.accessLoggerMiddleware != nil && c.ShouldAddAccessLogs(resourceName) {
		chain = chain.Append(accesslog.WrapHandler(c.accessLoggerMiddleware))
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return accesslog.NewFieldHandler(next, logs.EntryPointName, entryPointName, accesslog.InitServiceFields), nil
		})
	}

	if c.tracer != nil && c.ShouldAddTracing(resourceName) {
		chain = chain.Append(tracingMiddle.WrapEntryPointHandler(ctx, c.tracer, entryPointName))
	}

	if c.metricsRegistry != nil && c.metricsRegistry.IsEpEnabled() && c.ShouldAddMetrics(resourceName) {
		metricsHandler := metricsMiddle.WrapEntryPointHandler(ctx, c.metricsRegistry, entryPointName)

		if c.tracer != nil && c.ShouldAddTracing(resourceName) {
			chain = chain.Append(tracingMiddle.WrapMiddleware(ctx, metricsHandler))
		} else {
			chain = chain.Append(metricsHandler)
		}
	}

	return chain
}

// ShouldAddAccessLogs returns whether the access logs should be enabled for the given resource.
func (c *ObservabilityMgr) ShouldAddAccessLogs(resourceName string) bool {
	if c == nil {
		return false
	}

	return c.config.AccessLog != nil && (c.config.AccessLog.AddInternals || !strings.HasSuffix(resourceName, "@internal"))
}

// ShouldAddMetrics returns whether the metrics should be enabled for the given resource.
func (c *ObservabilityMgr) ShouldAddMetrics(resourceName string) bool {
	if c == nil {
		return false
	}

	return c.config.Metrics != nil && (c.config.Metrics.AddInternals || !strings.HasSuffix(resourceName, "@internal"))
}

// ShouldAddTracing returns whether the tracing should be enabled for the given resource.
func (c *ObservabilityMgr) ShouldAddTracing(resourceName string) bool {
	if c == nil {
		return false
	}

	return c.config.Tracing != nil && (c.config.Tracing.AddInternals || !strings.HasSuffix(resourceName, "@internal"))
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

	if c.tracerCloser != nil {
		if err := c.tracerCloser.Close(); err != nil {
			log.Error().Err(err).Msg("Could not close the tracer")
		}
	}
}

func (c *ObservabilityMgr) RotateAccessLogs() error {
	if c.accessLoggerMiddleware == nil {
		return nil
	}

	return c.accessLoggerMiddleware.Rotate()
}
