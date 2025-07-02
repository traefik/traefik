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
	"github.com/traefik/traefik/v3/pkg/types"
)

type contextKey int

const (
	accessLogsKey     contextKey = iota
	metricsKey        contextKey = iota
	semConvMetricsKey contextKey = iota
	minimalTracing    contextKey = iota
	detailedTracing   contextKey = iota
)

// AccessLogsEnabled returns whether metrics are enabled.
func AccessLogsEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(accessLogsKey).(bool); ok {
		return enabled
	}

	return false
}

// MetricsEnabled returns whether metrics are enabled.
func MetricsEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(metricsKey).(bool); ok {
		return enabled
	}

	return false
}

// SemConvMetricsEnabled returns whether metrics are enabled.
func SemConvMetricsEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(semConvMetricsKey).(bool); ok {
		return enabled
	}

	return false
}

// MinimalTraceEnabled returns whether minimal tracing is enabled.
func MinimalTraceEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(minimalTracing).(bool); ok {
		return enabled
	}

	return false
}

// DetailedTraceEnabled returns whether detailed tracing is enabled.
func DetailedTraceEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(detailedTracing).(bool); ok {
		return enabled
	}

	return false
}

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

// BuildContext returns a context with the observability configuration.
func (o *ObservabilityMgr) BuildContext(ctx context.Context, serviceName string, config *dynamic.RouterObservabilityConfig) context.Context {
	ctx = context.WithValue(ctx, accessLogsKey, o.shouldAccessLog(serviceName, config))
	ctx = context.WithValue(ctx, metricsKey, o.shouldMeter(serviceName, config))
	ctx = context.WithValue(ctx, semConvMetricsKey, o.shouldMeterSemConv(serviceName, config))
	ctx = context.WithValue(ctx, minimalTracing, o.shouldTrace(serviceName, config, types.MinimalVerbosity))
	return context.WithValue(ctx, detailedTracing, o.shouldTrace(serviceName, config, types.DetailedVerbosity))
}

// BuildEPChain an observability middleware chain by entry point.
func (o *ObservabilityMgr) BuildEPChain(ctx context.Context, entryPointName string) alice.Chain {
	chain := alice.New()

	if o == nil {
		return chain
	}

	if AccessLogsEnabled(ctx) || MetricsEnabled(ctx) {
		chain = chain.Append(capture.Wrap)
	}

	// As the Entry point observability middleware ensures that the tracing is added to the request and logger context,
	// it needs to be added before the access log middleware to ensure that the trace ID is logged.
	if o.tracer != nil && MinimalTraceEnabled(ctx) {
		chain = chain.Append(observability.EntryPointHandler(ctx, o.tracer, entryPointName))
	}

	if o.accessLoggerMiddleware != nil && AccessLogsEnabled(ctx) {
		chain = chain.Append(accesslog.WrapHandler(o.accessLoggerMiddleware))
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return accesslog.NewFieldHandler(next, logs.EntryPointName, entryPointName, accesslog.InitServiceFields), nil
		})
	}

	if MetricsEnabled(ctx) {
		metricsHandler := mmetrics.WrapEntryPointHandler(ctx, o.metricsRegistry, entryPointName)

		if DetailedTraceEnabled(ctx) {
			chain = chain.Append(observability.WrapMiddleware(ctx, metricsHandler))
		} else {
			chain = chain.Append(metricsHandler)
		}
	}

	// Semantic convention server metrics handler.
	if SemConvMetricsEnabled(ctx) {
		chain = chain.Append(observability.SemConvServerMetricsHandler(ctx, o.semConvMetricRegistry))
	}

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

// shouldAccessLog returns whether the access logs should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) shouldAccessLog(serviceName string, observabilityConfig *dynamic.RouterObservabilityConfig) bool {
	if o == nil {
		return false
	}

	if o.config.AccessLog == nil {
		return false
	}

	if strings.HasSuffix(serviceName, "@internal") && !o.config.AccessLog.AddInternals {
		return false
	}

	return observabilityConfig == nil || observabilityConfig.AccessLogs == nil || *observabilityConfig.AccessLogs
}

// shouldMeter returns whether the metrics should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) shouldMeter(serviceName string, observabilityConfig *dynamic.RouterObservabilityConfig) bool {
	if o == nil || o.metricsRegistry == nil {
		return false
	}

	if !o.metricsRegistry.IsEpEnabled() && !o.metricsRegistry.IsRouterEnabled() && !o.metricsRegistry.IsSvcEnabled() {
		return false
	}

	if o.config.Metrics == nil {
		return false
	}

	if strings.HasSuffix(serviceName, "@internal") && !o.config.Metrics.AddInternals {
		return false
	}

	return observabilityConfig == nil || observabilityConfig.Metrics == nil || *observabilityConfig.Metrics
}

// shouldMeterSemConv returns whether the OTel semantic convention metrics should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) shouldMeterSemConv(serviceName string, observabilityConfig *dynamic.RouterObservabilityConfig) bool {
	if o == nil {
		return false
	}

	if o.config.Metrics == nil {
		return false
	}

	if strings.HasSuffix(serviceName, "@internal") && !o.config.Metrics.AddInternals {
		return false
	}

	return observabilityConfig == nil || observabilityConfig.Metrics == nil || *observabilityConfig.Metrics
}

// shouldTrace returns whether the tracing should be enabled for the given serviceName and the observability config.
func (o *ObservabilityMgr) shouldTrace(serviceName string, observabilityConfig *dynamic.RouterObservabilityConfig, verbosity types.TracingVerbosity) bool {
	if o == nil {
		return false
	}

	if o.config.Tracing == nil {
		return false
	}

	if strings.HasSuffix(serviceName, "@internal") && !o.config.Tracing.AddInternals {
		return false
	}

	if !observabilityConfig.TraceVerbosity.Allows(verbosity) {
		return false
	}

	return observabilityConfig.Tracing == nil || *observabilityConfig.Tracing
}
