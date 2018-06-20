package middlewares

import (
	"net/http"

	"github.com/containous/traefik/middlewares/tracing"
	"github.com/vulcand/oxy/cbreaker"
)

// CircuitBreaker holds the oxy circuit breaker.
type CircuitBreaker struct {
	circuitBreaker *cbreaker.CircuitBreaker
}

// NewCircuitBreaker returns a new CircuitBreaker.
func NewCircuitBreaker(next http.Handler, expression string, options ...cbreaker.CircuitBreakerOption) (*CircuitBreaker, error) {
	circuitBreaker, err := cbreaker.New(next, expression, options...)
	if err != nil {
		return nil, err
	}
	return &CircuitBreaker{circuitBreaker}, nil
}

// NewCircuitBreakerOptions returns a new CircuitBreakerOption
func NewCircuitBreakerOptions(expression string) cbreaker.CircuitBreakerOption {
	return cbreaker.Fallback(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracing.LogEventf(r, "blocked by circuit-breaker (%q)", expression)

		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
	}))
}

func (cb *CircuitBreaker) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	cb.circuitBreaker.ServeHTTP(rw, r)
}
