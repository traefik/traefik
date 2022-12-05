package circuitbreaker

import (
	"context"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"github.com/vulcand/oxy/v2/cbreaker"
)

const typeName = "CircuitBreaker"

type circuitBreaker struct {
	circuitBreaker *cbreaker.CircuitBreaker
	name           string
}

// New creates a new circuit breaker middleware.
func New(ctx context.Context, next http.Handler, confCircuitBreaker dynamic.CircuitBreaker, name string) (http.Handler, error) {
	expression := confCircuitBreaker.Expression

	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName))
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up with expression: %s", expression)

	cbOpts := []cbreaker.Option{
		cbreaker.Fallback(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			tracing.SetErrorWithEvent(req, "blocked by circuit-breaker (%q)", expression)
			rw.WriteHeader(http.StatusServiceUnavailable)

			if _, err := rw.Write([]byte(http.StatusText(http.StatusServiceUnavailable))); err != nil {
				log.FromContext(req.Context()).Error(err)
			}
		})),
	}

	if confCircuitBreaker.CheckPeriod > 0 {
		cbOpts = append(cbOpts, cbreaker.CheckPeriod(time.Duration(confCircuitBreaker.CheckPeriod)))
	}

	if confCircuitBreaker.FallbackDuration > 0 {
		cbOpts = append(cbOpts, cbreaker.FallbackDuration(time.Duration(confCircuitBreaker.FallbackDuration)))
	}

	if confCircuitBreaker.RecoveryDuration > 0 {
		cbOpts = append(cbOpts, cbreaker.RecoveryDuration(time.Duration(confCircuitBreaker.RecoveryDuration)))
	}

	oxyCircuitBreaker, err := cbreaker.New(next, expression, cbOpts...)
	if err != nil {
		return nil, err
	}
	return &circuitBreaker{
		circuitBreaker: oxyCircuitBreaker,
		name:           name,
	}, nil
}

func (c *circuitBreaker) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func (c *circuitBreaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c.circuitBreaker.ServeHTTP(rw, req)
}
