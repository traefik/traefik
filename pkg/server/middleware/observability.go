package middleware

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	mmetrics "github.com/traefik/traefik/v3/pkg/middlewares/metrics"
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
func (o *ObservabilityMgr) BuildEPChain(ctx context.Context, entryPointName string, resourceName string, observabilityConfig *dynamic.RouterObservabilityConfig) alice.Chain {
	chain := alice.New()

	if o == nil {
		return chain
	}

	if o.accessLoggerMiddleware != nil || o.metricsRegistry != nil && (o.metricsRegistry.IsEpEnabled() || o.metricsRegistry.IsRouterEnabled() || o.metricsRegistry.IsSvcEnabled()) {
		if o.ShouldAddAccessLogs(resourceName, observabilityConfig) || o.ShouldAddMetrics(resourceName, observabilityConfig) {
			chain = chain.Append(capture.Wrap)
		}
	}

	// As the Entry point observability middleware ensures that the tracing is added to the request and logger context,
	// it needs to be added before the access log middleware to ensure that the trace ID is logged.
	if o.tracer != nil && o.ShouldAddTracing(resourceName, observabilityConfig) {
		chain = chain.Append(observability.EntryPointHandler(ctx, o.tracer, entryPointName))
	}

	if o.accessLoggerMiddleware != nil && o.ShouldAddAccessLogs(resourceName, observabilityConfig) {
		chain = chain.Append(accesslog.WrapHandler(o.accessLoggerMiddleware))
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return accesslog.NewFieldHandler(next, logs.EntryPointName, entryPointName, accesslog.InitServiceFields), nil
		})
	}

	// Semantic convention server metrics handler.
	if o.semConvMetricRegistry != nil && o.ShouldAddMetrics(resourceName, observabilityConfig) {
		chain = chain.Append(observability.SemConvServerMetricsHandler(ctx, o.semConvMetricRegistry))
	}

	if o.metricsRegistry != nil && o.metricsRegistry.IsEpEnabled() && o.ShouldAddMetrics(resourceName, observabilityConfig) {
		metricsHandler := mmetrics.WrapEntryPointHandler(ctx, o.metricsRegistry, entryPointName)

		if o.tracer != nil && o.ShouldAddTracing(resourceName, observabilityConfig) {
			chain = chain.Append(observability.WrapMiddleware(ctx, metricsHandler))
		} else {
			chain = chain.Append(metricsHandler)
		}
	}

	// Inject context keys to control whether to produce metrics further downstream (services, round-tripper),
	// because the router configuration cannot be evaluated during build time for services.
	if observabilityConfig != nil && observabilityConfig.Metrics != nil && !*observabilityConfig.Metrics {
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(rw, req.WithContext(context.WithValue(req.Context(), observability.DisableMetricsKey, true)))
			}), nil
		})
	}

	return chain
}

// ShouldAddAccessLogs returns whether the access logs should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) ShouldAddAccessLogs(serviceName string, observabilityConfig *dynamic.RouterObservabilityConfig) bool {
	if o == nil {
		return false
	}

	if o.config.AccessLog == nil {
		return false
	}

	if strings.HasSuffix(serviceName, "@internal") && !o.config.AccessLog.AddInternals {
		return false
	}

	return observabilityConfig == nil || observabilityConfig.AccessLogs != nil && *observabilityConfig.AccessLogs
}

// ShouldAddMetrics returns whether the metrics should be enabled for the given resource and the observability config.
func (o *ObservabilityMgr) ShouldAddMetrics(serviceName string, observabilityConfig *dynamic.RouterObservabilityConfig) bool {
	if o == nil {
		return false
	}

	if o.config.Metrics == nil {
		return false
	}

	if strings.HasSuffix(serviceName, "@internal") && !o.config.Metrics.AddInternals {
		return false
	}

	return observabilityConfig == nil || observabilityConfig.Metrics != nil && *observabilityConfig.Metrics
}

// ShouldAddTracing returns whether the tracing should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) ShouldAddTracing(serviceName string, observabilityConfig *dynamic.RouterObservabilityConfig) bool {
	if o == nil {
		return false
	}

	if o.config.Tracing == nil {
		return false
	}

	if strings.HasSuffix(serviceName, "@internal") && !o.config.Tracing.AddInternals {
		return false
	}

	return observabilityConfig == nil || observabilityConfig.Tracing != nil && *observabilityConfig.Tracing
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
