package middlewares

import (
	"net/http"

	"github.com/containous/traefik/middlewares/common"
	"github.com/vulcand/oxy/cbreaker"
)

var _ common.Middleware = &CircuitBreaker{}

// CircuitBreaker holds the oxy circuit breaker.
type CircuitBreaker struct {
	common.BasicMiddleware
	circuitBreaker *cbreaker.CircuitBreaker
}

// NewCircuitBreaker returns a new CircuitBreaker.
func NewCircuitBreaker(next http.Handler, expression string, options ...cbreaker.CircuitBreakerOption) (common.Middleware, error) {
	circuitBreaker, err := cbreaker.New(next, expression, options...)
	if err != nil {
		return nil, err
	}
	return &CircuitBreaker{common.NewMiddleware(next), circuitBreaker}, nil
}

func (cb *CircuitBreaker) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	cb.circuitBreaker.ServeHTTP(rw, r)
}
