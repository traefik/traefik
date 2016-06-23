package middlewares

import (
	"net/http"

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

func (cb *CircuitBreaker) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cb.circuitBreaker.ServeHTTP(rw, r)
}
