package middlewares

import (
	"net/http"

	"github.com/containous/traefik/healthcheck"
)

// EmptyBackendHandler is a middlware that checks whether the current Backend
// has at least one active Server in respect to the healthchecks and if this
// is not the case, it will stop the middleware chain and respond with 503.
type EmptyBackendHandler struct {
	lb   healthcheck.LoadBalancer
	next http.Handler
}

// NewEmptyBackendHandler creates a new EmptyBackendHandler instance.
func NewEmptyBackendHandler(lb healthcheck.LoadBalancer, next http.Handler) *EmptyBackendHandler {
	return &EmptyBackendHandler{lb: lb, next: next}
}

// ServeHTTP responds with 503 when there is no active Server and otherwise
// invokes the next handler in the middleware chain.
func (h *EmptyBackendHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if len(h.lb.Servers()) == 0 {
		rw.WriteHeader(http.StatusServiceUnavailable)
		rw.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
	} else {
		h.next.ServeHTTP(rw, r)
	}
}
