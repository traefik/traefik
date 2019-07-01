package ping

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containous/mux"
)

// Handler expose ping routes.
type Handler struct {
	EntryPoint  string   `description:"Ping entryPoint." json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty" export:"true"`
	Middlewares []string `description:"Middleware list." json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	terminating bool
}

// SetDefaults sets the default values.
func (h *Handler) SetDefaults() {
	h.EntryPoint = "traefik"
}

// WithContext causes the ping endpoint to serve non 200 responses.
func (h *Handler) WithContext(ctx context.Context) {
	go func() {
		<-ctx.Done()
		h.terminating = true
	}()
}

// Append adds ping routes on a router.
func (h *Handler) Append(router *mux.Router) {
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
