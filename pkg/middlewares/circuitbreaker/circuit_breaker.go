package circuitbreaker

import (
	"context"
	"net/http"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/tracing"
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

// New creates a new circuit breaker middleware.
func New(ctx context.Context, next http.Handler, confCircuitBreaker config.CircuitBreaker, name string) (http.Handler, error) {
	expression := confCircuitBreaker.Expression

	logger := middlewares.GetLogger(ctx, name, typeName)
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

// NewCircuitBreakerOptions returns a new CircuitBreakerOption
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
	middlewares.GetLogger(req.Context(), c.name, typeName).Debug("Entering middleware")
	c.circuitBreaker.ServeHTTP(rw, req)
}
