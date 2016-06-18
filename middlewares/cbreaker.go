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
func NewCircuitBreaker(next http.Handler, expression string, options ...cbreaker.CircuitBreakerOption) *CircuitBreaker {
	circuitBreaker, _ := cbreaker.New(next, expression, options...)
	return &CircuitBreaker{circuitBreaker}
}

func (cb *CircuitBreaker) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cb.circuitBreaker.ServeHTTP(rw, r)
}
