package ready

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Handler expose ready routes.
type Handler struct {
	EntryPoint    string `description:"EntryPoint" json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty" export:"true"`
	ManualRouting bool   `description:"Manual routing" json:"manualRouting,omitempty" toml:"manualRouting,omitempty" yaml:"manualRouting,omitempty" export:"true"`
	ready         atomic.Bool
}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// SetDefaults sets the default values.
func (h *Handler) SetDefaults() {
	h.EntryPoint = "traefik"
}

// SetReady sets the handler as ready.
func (h *Handler) SetReady() {
	h.ready.Store(true)
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	statusCode := http.StatusServiceUnavailable
	if h.ready.Load() {
		statusCode = http.StatusOK
	}
	response.WriteHeader(statusCode)
	fmt.Fprint(response, http.StatusText(statusCode))
}
