package middlewares

import (
	"net/http"

	"github.com/containous/traefik/healthcheck"
)

// EmptyBackendHandler is a middlware that checks whether the current Backend
// has at least one active Server in respect to the healthchecks and if this
// is not the case, it will stop the middleware chain and respond with 503.
type EmptyBackendHandler struct {
	next healthcheck.BalancerHandler
}

// NewEmptyBackendHandler creates a new EmptyBackendHandler instance.
func NewEmptyBackendHandler(lb healthcheck.BalancerHandler) *EmptyBackendHandler {
	return &EmptyBackendHandler{next: lb}
}

// ServeHTTP responds with 503 when there is no active Server and otherwise
// invokes the next handler in the middleware chain.
func (h *EmptyBackendHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if len(h.next.Servers()) == 0 {
		rw.WriteHeader(http.StatusServiceUnavailable)
		rw.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
	} else {
		h.next.ServeHTTP(rw, r)
	}
}
