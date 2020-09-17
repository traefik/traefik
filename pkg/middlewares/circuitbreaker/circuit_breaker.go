package circuitbreaker

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"github.com/vulcand/oxy/cbreaker"
)

const (
	typeName = "CircuitBreaker"
)

type circuitBreaker struct {
	circuitBreaker *cbreaker.CircuitBreaker
	name           string
}

// New creates a new circuit breaker middleware.
func New(ctx context.Context, next http.Handler, confCircuitBreaker dynamic.CircuitBreaker, name string) (http.Handler, error) {
	expression := confCircuitBreaker.Expression

	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName))
	logger.Debug("Creating middleware")
	logger.Debug("Setting up with expression: %s", expression)

	oxyCircuitBreaker, err := cbreaker.New(next, expression, createCircuitBreakerOptions(expression))
	if err != nil {
		return nil, err
	}
	return &circuitBreaker{
		circuitBreaker: oxyCircuitBreaker,
		name:           name,
	}, nil
}

// NewCircuitBreakerOptions returns a new CircuitBreakerOption.
func createCircuitBreakerOptions(expression string) cbreaker.CircuitBreakerOption {
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
