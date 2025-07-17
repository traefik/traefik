package middleware

import (
	"context"
	"io"
	"net/http"

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
	"github.com/traefik/traefik/v3/pkg/types"
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
func (o *ObservabilityMgr) BuildEPChain(ctx context.Context, entryPointName string, internal bool, config dynamic.RouterObservabilityConfig) alice.Chain {
	chain := alice.New()

	if o == nil {
		return chain
	}

	// Injection of the observability variables in the request context.
	// This injection must be the first step in order for other observability middlewares to rely on it.
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return o.observabilityContextHandler(next, internal, config), nil
	})

	// Capture middleware for accessLogs or metrics.
	if o.shouldAccessLog(internal, config) || o.shouldMeter(internal, config) || o.shouldMeterSemConv(internal, config) {
		chain = chain.Append(capture.Wrap)
	}

	// As the Entry point observability middleware ensures that the tracing is added to the request and logger context,
	// it needs to be added before the access log middleware to ensure that the trace ID is logged.
	chain = chain.Append(observability.EntryPointHandler(ctx, o.tracer, entryPointName))

	// Access log handlers.
	chain = chain.Append(o.accessLoggerMiddleware.AliceConstructor())
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return accesslog.NewFieldHandler(next, logs.EntryPointName, entryPointName, accesslog.InitServiceFields), nil
	})

	// Entrypoint metrics handler.
	metricsHandler := mmetrics.EntryPointMetricsHandler(ctx, o.metricsRegistry, entryPointName)
	chain = chain.Append(observability.WrapMiddleware(ctx, metricsHandler))

	// Semantic convention server metrics handler.
	chain = chain.Append(observability.SemConvServerMetricsHandler(ctx, o.semConvMetricRegistry))

	return chain
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

func (o *ObservabilityMgr) observabilityContextHandler(next http.Handler, internal bool, config dynamic.RouterObservabilityConfig) http.Handler {
	return observability.WithObservabilityHandler(next, observability.Observability{
		AccessLogsEnabled:      o.shouldAccessLog(internal, config),
		MetricsEnabled:         o.shouldMeter(internal, config),
		SemConvMetricsEnabled:  o.shouldMeterSemConv(internal, config),
		TracingEnabled:         o.shouldTrace(internal, config, types.MinimalVerbosity),
		DetailedTracingEnabled: o.shouldTrace(internal, config, types.DetailedVerbosity),
	})
}

// shouldAccessLog returns whether the access logs should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) shouldAccessLog(internal bool, observabilityConfig dynamic.RouterObservabilityConfig) bool {
	if o == nil {
		return false
	}

	if o.config.AccessLog == nil {
		return false
	}

	if internal && !o.config.AccessLog.AddInternals {
		return false
	}

	return observabilityConfig.AccessLogs == nil || *observabilityConfig.AccessLogs
}

// shouldMeter returns whether the metrics should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) shouldMeter(internal bool, observabilityConfig dynamic.RouterObservabilityConfig) bool {
	if o == nil || o.metricsRegistry == nil {
		return false
	}

	if !o.metricsRegistry.IsEpEnabled() && !o.metricsRegistry.IsRouterEnabled() && !o.metricsRegistry.IsSvcEnabled() {
		return false
	}

	if o.config.Metrics == nil {
		return false
	}

	if internal && !o.config.Metrics.AddInternals {
		return false
	}

	return observabilityConfig.Metrics == nil || *observabilityConfig.Metrics
}

// shouldMeterSemConv returns whether the OTel semantic convention metrics should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) shouldMeterSemConv(internal bool, observabilityConfig dynamic.RouterObservabilityConfig) bool {
	if o == nil || o.semConvMetricRegistry == nil {
		return false
	}

	if o.config.Metrics == nil {
		return false
	}

	if internal && !o.config.Metrics.AddInternals {
		return false
	}

	return observabilityConfig.Metrics == nil || *observabilityConfig.Metrics
}

// shouldTrace returns whether the tracing should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) shouldTrace(internal bool, observabilityConfig dynamic.RouterObservabilityConfig, verbosity types.TracingVerbosity) bool {
	if o == nil {
		return false
	}

	if o.config.Tracing == nil {
		return false
	}

	if internal && !o.config.Tracing.AddInternals {
		return false
	}

	if !observabilityConfig.TraceVerbosity.Allows(verbosity) {
		return false
	}

	return observabilityConfig.Tracing == nil || *observabilityConfig.Tracing
}
