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
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/tracing"
)

// ObservabilityMgr is a manager for observability (AccessLogs, Metrics and Tracing) enablement.
type ObservabilityMgr struct {
	config                 static.Configuration
	accessLoggerMiddleware *accesslog.Handler
	metricsRegistry        metrics.Registry
	semConvMetricRegistry  *metrics.SemConvMetricsRegistry
	tracer                 *tracing.Tracer
	tracerCloser           io.Closer
}

// NewObservabilityMgr creates a new ObservabilityMgr.
func NewObservabilityMgr(config static.Configuration, metricsRegistry metrics.Registry, semConvMetricRegistry *metrics.SemConvMetricsRegistry, accessLoggerMiddleware *accesslog.Handler, tracer *tracing.Tracer, tracerCloser io.Closer) *ObservabilityMgr {
	return &ObservabilityMgr{
		config:                 config,
		metricsRegistry:        metricsRegistry,
		semConvMetricRegistry:  semConvMetricRegistry,
		accessLoggerMiddleware: accessLoggerMiddleware,
		tracer:                 tracer,
		tracerCloser:           tracerCloser,
	}
}

// BuildEPChain an observability middleware chain by entry point.
func (o *ObservabilityMgr) BuildEPChain(ctx context.Context, entryPointName string, resourceName string) alice.Chain {
	chain := alice.New()

	if o == nil {
		return chain
	}

	if o.accessLoggerMiddleware != nil || o.metricsRegistry != nil && (o.metricsRegistry.IsEpEnabled() || o.metricsRegistry.IsRouterEnabled() || o.metricsRegistry.IsSvcEnabled()) {
		if o.ShouldAddAccessLogs(resourceName) || o.ShouldAddMetrics(resourceName) {
			chain = chain.Append(capture.Wrap)
		}
	}

	if o.accessLoggerMiddleware != nil && o.ShouldAddAccessLogs(resourceName) {
		chain = chain.Append(accesslog.WrapHandler(o.accessLoggerMiddleware))
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return accesslog.NewFieldHandler(next, logs.EntryPointName, entryPointName, accesslog.InitServiceFields), nil
		})
	}

	if (o.tracer != nil && o.ShouldAddTracing(resourceName)) || (o.metricsRegistry != nil && o.metricsRegistry.IsEpEnabled() && o.ShouldAddMetrics(resourceName)) {
		chain = chain.Append(observability.WrapEntryPointHandler(ctx, o.tracer, o.semConvMetricRegistry, entryPointName))
	}

	if o.metricsRegistry != nil && o.metricsRegistry.IsEpEnabled() && o.ShouldAddMetrics(resourceName) {
		metricsHandler := metricsMiddle.WrapEntryPointHandler(ctx, o.metricsRegistry, entryPointName)

		if o.tracer != nil && o.ShouldAddTracing(resourceName) {
			chain = chain.Append(observability.WrapMiddleware(ctx, metricsHandler))
		} else {
			chain = chain.Append(metricsHandler)
		}
	}

	return chain
}

// ShouldAddAccessLogs returns whether the access logs should be enabled for the given resource.
func (o *ObservabilityMgr) ShouldAddAccessLogs(resourceName string) bool {
	if o == nil {
		return false
	}

	return o.config.AccessLog != nil && (o.config.AccessLog.AddInternals || !strings.HasSuffix(resourceName, "@internal"))
}

// ShouldAddMetrics returns whether the metrics should be enabled for the given resource.
func (o *ObservabilityMgr) ShouldAddMetrics(resourceName string) bool {
	if o == nil {
		return false
	}

	return o.config.Metrics != nil && (o.config.Metrics.AddInternals || !strings.HasSuffix(resourceName, "@internal"))
}

// ShouldAddTracing returns whether the tracing should be enabled for the given resource.
func (o *ObservabilityMgr) ShouldAddTracing(resourceName string) bool {
	if o == nil {
		return false
	}

	return o.config.Tracing != nil && (o.config.Tracing.AddInternals || !strings.HasSuffix(resourceName, "@internal"))
}

// MetricsRegistry is an accessor to the metrics registry.
func (o *ObservabilityMgr) MetricsRegistry() metrics.Registry {
	if o == nil {
		return nil
	}

	return o.metricsRegistry
}

// SemConvMetricsRegistry is an accessor to the semantic conventions metrics registry.
func (o *ObservabilityMgr) SemConvMetricsRegistry() *metrics.SemConvMetricsRegistry {
	if o == nil {
		return nil
	}

	return o.semConvMetricRegistry
}

// Close closes the accessLogger and tracer.
func (o *ObservabilityMgr) Close() {
	if o == nil {
		return
	}

	if o.accessLoggerMiddleware != nil {
		if err := o.accessLoggerMiddleware.Close(); err != nil {
			log.Error().Err(err).Msg("Could not close the access log file")
		}
	}

	if o.tracerCloser != nil {
		if err := o.tracerCloser.Close(); err != nil {
			log.Error().Err(err).Msg("Could not close the tracer")
		}
	}
}

func (o *ObservabilityMgr) RotateAccessLogs() error {
	if o.accessLoggerMiddleware == nil {
		return nil
	}

	return o.accessLoggerMiddleware.Rotate()
}
