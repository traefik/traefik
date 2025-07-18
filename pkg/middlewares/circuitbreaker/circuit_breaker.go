package circuitbreaker

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
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

	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")
	logger.Debug().Msgf("Setting up with expression: %s", expression)

	responseCode := confCircuitBreaker.ResponseCode

	cbOpts := []cbreaker.Option{
		cbreaker.Fallback(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			observability.SetStatusErrorf(req.Context(), "blocked by circuit-breaker (%q)", expression)
			rw.WriteHeader(responseCode)

			if _, err := rw.Write([]byte(http.StatusText(responseCode))); err != nil {
				log.Ctx(req.Context()).Error().Err(err).Send()
			}
		})),
		cbreaker.Logger(logs.NewOxyWrapper(*logger)),
		cbreaker.Verbose(logger.GetLevel() == zerolog.TraceLevel),
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

func (c *circuitBreaker) GetTracingInformation() (string, string) {
	return c.name, typeName
}

func (c *circuitBreaker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c.circuitBreaker.ServeHTTP(rw, req)
}
