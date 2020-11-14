package middleware

import (
	"context"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/middlewares/accesslog"
	metricsmiddleware "github.com/traefik/traefik/v2/pkg/middlewares/metrics"
	"github.com/traefik/traefik/v2/pkg/middlewares/requestdecorator"
	mTracing "github.com/traefik/traefik/v2/pkg/middlewares/tracing"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"github.com/traefik/traefik/v2/pkg/tracing/jaeger"
)

// ChainBuilder Creates a middleware chain by entry point. It is used for middlewares that are created almost systematically and that need to be created before all others.
type ChainBuilder struct {
	metricsRegistry        metrics.Registry
	accessLoggerMiddleware *accesslog.Handler
	tracer                 *tracing.Tracing
	requestDecorator       *requestdecorator.RequestDecorator
}

// NewChainBuilder Creates a new ChainBuilder.
func NewChainBuilder(staticConfiguration static.Configuration, metricsRegistry metrics.Registry, accessLoggerMiddleware *accesslog.Handler) *ChainBuilder {
	return &ChainBuilder{
		metricsRegistry:        metricsRegistry,
		accessLoggerMiddleware: accessLoggerMiddleware,
		tracer:                 setupTracing(staticConfiguration.Tracing),
		requestDecorator:       requestdecorator.New(staticConfiguration.HostResolver),
	}
}

// Build a middleware chain by entry point.
func (c *ChainBuilder) Build(ctx context.Context, entryPointName string) alice.Chain {
	chain := alice.New()

	if c.accessLoggerMiddleware != nil {
		chain = chain.Append(accesslog.WrapHandler(c.accessLoggerMiddleware))
	}

	if c.tracer != nil {
		chain = chain.Append(mTracing.WrapEntryPointHandler(ctx, c.tracer, entryPointName))
	}

	if c.metricsRegistry != nil && c.metricsRegistry.IsEpEnabled() {
		chain = chain.Append(metricsmiddleware.WrapEntryPointHandler(ctx, c.metricsRegistry, entryPointName))
	}

	return chain.Append(requestdecorator.WrapHandler(c.requestDecorator))
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

func setupTracing(conf *static.Tracing) *tracing.Tracing {
	if conf == nil {
		return nil
	}

	var backend tracing.Backend

	if conf.Jaeger != nil {
		backend = conf.Jaeger
	}

	if conf.Zipkin != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create Zipkin backend.")
		} else {
			backend = conf.Zipkin
		}
	}

	if conf.Datadog != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create Datadog backend.")
		} else {
			backend = conf.Datadog
		}
	}

	if conf.Instana != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create Instana backend.")
		} else {
			backend = conf.Instana
		}
	}

	if conf.Haystack != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create Haystack backend.")
		} else {
			backend = conf.Haystack
		}
	}

	if conf.Elastic != nil {
		if backend != nil {
			log.WithoutContext().Error("Multiple tracing backend are not supported: cannot create Elastic backend.")
		} else {
			backend = conf.Elastic
		}
	}

	if backend == nil {
		log.WithoutContext().Debug("Could not initialize tracing, using Jaeger by default")
		defaultBackend := &jaeger.Config{}
		defaultBackend.SetDefaults()
		backend = defaultBackend
	}

	tracer, err := tracing.NewTracing(conf.ServiceName, conf.SpanNameLimit, backend)
	if err != nil {
		log.WithoutContext().Warnf("Unable to create tracer: %v", err)
		return nil
	}
	return tracer
}
