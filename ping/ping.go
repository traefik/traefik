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
func (g *Handler) WithContext(ctx context.Context) {
	go func() {
		<-ctx.Done()
		g.terminating = true
	}()
}

// AddRoutes add ping routes on a router
func (g *Handler) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet, http.MethodHead).Path("/ping").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			statusCode := http.StatusOK
			if g.terminating {
				statusCode = http.StatusServiceUnavailable
			}
			response.WriteHeader(statusCode)
			fmt.Fprint(response, http.StatusText(statusCode))
		})
}
