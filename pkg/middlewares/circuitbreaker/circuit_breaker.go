package circuitbreaker

import (
	"context"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/vulcand/oxy/cbreaker"
)

const (
	typeName = "CircuitBreaker"
)

type circuitBreaker struct {
	circuitBreaker *cbreaker.CircuitBreaker
	name           string
}

type serviceBuilder interface {
	BuildHTTP(ctx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error)
}

// New creates a new circuit breaker middleware.
func New(ctx context.Context, next http.Handler, confCircuitBreaker dynamic.CircuitBreaker, serviceBuilder serviceBuilder, name string) (http.Handler, error) {
	expression := confCircuitBreaker.Expression

	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName))
	logger.Debug("Creating middleware")
	logger.Debug("Setting up with expression: %s", expression)

	oxyCircuitBreaker, err := cbreaker.New(next, expression, createCircuitBreakerOptions(ctx, logger, confCircuitBreaker, serviceBuilder, expression))
	if err != nil {
		return nil, err
	}
	return &circuitBreaker{
		circuitBreaker: oxyCircuitBreaker,
		name:           name,
	}, nil
}

// NewCircuitBreakerOptions returns a new CircuitBreakerOption.
func createCircuitBreakerOptions(ctx context.Context, logger log.Logger, confCircuitBreaker dynamic.CircuitBreaker, serviceBuilder serviceBuilder, expression string) cbreaker.CircuitBreakerOption {
	if len(confCircuitBreaker.Service) > 0 {
		h, err := serviceBuilder.BuildHTTP(ctx, confCircuitBreaker.Service, nil)

		if err != nil {
			logger.Error(err)
		} else {
			return cbreaker.Fallback(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				tracing.SetErrorWithEvent(req, "forwarded by circuit-breaker (%q) to %s", expression, confCircuitBreaker.Service)

				h.ServeHTTP(rw, req)
			}))
		}
	}

	return cbreaker.Fallback(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		tracing.SetErrorWithEvent(req, "blocked by circuit-breaker (%q)", expression)
		rw.WriteHeader(http.StatusServiceUnavailable)

		if _, err := rw.Write([]byte(http.StatusText(http.StatusServiceUnavailable))); err != nil {
			log.FromContext(req.Context()).Error(err)
		}
	}))
}

func (c *circuitBreaker) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func (c *circuitBreaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c.circuitBreaker.ServeHTTP(rw, req)
}
