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

func NewCircuitBreaker(next http.Handler, options ...cbreaker.CircuitBreakerOption) *CircuitBreaker {
	circuitBreaker, _ := cbreaker.New(next, "NetworkErrorRatio() > 0.5", options...)
	return &CircuitBreaker{circuitBreaker}
}

func (cb *CircuitBreaker) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cb.circuitBreaker.ServeHTTP(rw, r)
}
