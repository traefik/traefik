package ping

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/containous/mux"
)

// Handler expose ping routes
type Handler struct {
	EntryPoint  string `description:"Ping entryPoint" export:"true"`
	terminating bool
	lock        sync.RWMutex
}

// SetTerminating causes the ping endpoint to serve non 200 responses.
func (g *Handler) SetTerminating() {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.terminating = true
}

// AddRoutes add ping routes on a router
func (g *Handler) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet, http.MethodHead).Path("/ping").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			g.lock.RLock()
			defer g.lock.RUnlock()

			statusCode := http.StatusOK
			if g.terminating {
				statusCode = http.StatusServiceUnavailable
			}
			response.WriteHeader(statusCode)
			fmt.Fprint(response, http.StatusText(statusCode))
		})
}
