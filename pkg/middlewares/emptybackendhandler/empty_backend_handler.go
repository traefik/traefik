package emptybackendhandler

import (
	"net/http"

	"github.com/traefik/traefik/v2/pkg/healthcheck"
)

// EmptyBackend is a middleware that checks whether the current Backend
// has at least one active Server in respect to the healthchecks and if this
// is not the case, it will stop the middleware chain and respond with 503.
type emptyBackend struct {
	next healthcheck.BalancerHandler
}

// New creates a new EmptyBackend middleware.
func New(lb healthcheck.BalancerHandler) http.Handler {
	return &emptyBackend{next: lb}
}

// ServeHTTP responds with 503 when there is no active Server and otherwise
// invokes the next handler in the middleware chain.
func (e *emptyBackend) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(e.next.Servers()) == 0 {
		rw.WriteHeader(http.StatusServiceUnavailable)
		_, err := rw.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	} else {
		e.next.ServeHTTP(rw, req)
	}
}
