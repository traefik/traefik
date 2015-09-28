/*
Copyright
*/
package middlewares

import (
	"net/http"

	"github.com/mailgun/oxy/cbreaker"
)

type CircuitBreaker struct {
	circuitBreaker *cbreaker.CircuitBreaker
}

func NewCircuitBreaker(next http.Handler, expression string, options ...cbreaker.CircuitBreakerOption) *CircuitBreaker {
	circuitBreaker, _ := cbreaker.New(next, expression, options...)
	return &CircuitBreaker{circuitBreaker}
}

func (cb *CircuitBreaker) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cb.circuitBreaker.ServeHTTP(rw, r)
}
