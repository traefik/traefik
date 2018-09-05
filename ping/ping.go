package ping

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containous/mux"
)

// Handler expose ping routes
type Handler struct {
	EntryPoint  string `description:"Ping entryPoint" export:"true"`
	terminating bool
}

// WithContext causes the ping endpoint to serve non 200 responses.
func (h *Handler) WithContext(ctx context.Context) {
	go func() {
		<-ctx.Done()
		h.terminating = true
	}()
}

// AddRoutes add ping routes on a router
func (h *Handler) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet, http.MethodHead).Path("/ping").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			statusCode := http.StatusOK
			if h.terminating {
				statusCode = http.StatusServiceUnavailable
			}
			response.WriteHeader(statusCode)
			fmt.Fprint(response, http.StatusText(statusCode))
		})
}
