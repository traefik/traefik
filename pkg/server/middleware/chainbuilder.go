package middleware

import (
	"context"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v2/pkg/middlewares/capture"
	metricsmiddleware "github.com/traefik/traefik/v2/pkg/middlewares/metrics"
	mTracing "github.com/traefik/traefik/v2/pkg/middlewares/tracing"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

// ChainBuilder Creates a middleware chain by entry point. It is used for middlewares that are created almost systematically and that need to be created before all others.
type ChainBuilder struct {
	metricsRegistry        metrics.Registry
	accessLoggerMiddleware *accesslog.Handler
	tracer                 *tracing.Tracing
}

// NewChainBuilder Creates a new ChainBuilder.
func NewChainBuilder(metricsRegistry metrics.Registry, accessLoggerMiddleware *accesslog.Handler, tracer *tracing.Tracing) *ChainBuilder {
	return &ChainBuilder{
		metricsRegistry:        metricsRegistry,
		accessLoggerMiddleware: accessLoggerMiddleware,
		tracer:                 tracer,
	}
}

// Build a middleware chain by entry point.
func (c *ChainBuilder) Build(ctx context.Context, entryPointName string) alice.Chain {
	chain := alice.New()

	if c.accessLoggerMiddleware != nil || c.metricsRegistry != nil && (c.metricsRegistry.IsEpEnabled() || c.metricsRegistry.IsRouterEnabled() || c.metricsRegistry.IsSvcEnabled()) {
		chain = chain.Append(capture.Wrap)
	}

	if c.accessLoggerMiddleware != nil {
		chain = chain.Append(accesslog.WrapHandler(c.accessLoggerMiddleware))
	}

	if c.tracer != nil {
		chain = chain.Append(mTracing.WrapEntryPointHandler(ctx, c.tracer, entryPointName))
	}

	if c.metricsRegistry != nil && c.metricsRegistry.IsEpEnabled() {
		chain = chain.Append(metricsmiddleware.WrapEntryPointHandler(ctx, c.metricsRegistry, entryPointName))
	}

	return chain
}

// Close accessLogger and tracer.
func (c *ChainBuilder) Close() {
	if c.accessLoggerMiddleware != nil {
		if err := c.accessLoggerMiddleware.Close(); err != nil {
			log.WithoutContext().Errorf("Could not close the access log file: %s", err)
		}
	}

	if c.tracer != nil {
		c.tracer.Close()
	}
}
