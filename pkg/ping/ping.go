package ping

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
)

// Handler expose ping routes.
type Handler struct {
	EntryPoint            string `description:"EntryPoint" json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty" export:"true"`
	ManualRouting         bool   `description:"Manual routing" json:"manualRouting,omitempty" toml:"manualRouting,omitempty" yaml:"manualRouting,omitempty" export:"true"`
	TerminatingStatusCode int    `description:"Terminating status code" json:"terminatingStatusCode,omitempty" toml:"terminatingStatusCode,omitempty" yaml:"terminatingStatusCode,omitempty" export:"true"`
	terminating           atomic.Bool
}

// SetDefaults sets the default values.
func (h *Handler) SetDefaults() {
	h.EntryPoint = "traefik"
	h.TerminatingStatusCode = http.StatusServiceUnavailable
}

// WithContext causes the ping endpoint to serve non 200 responses.
func (h *Handler) WithContext(ctx context.Context) {
	go func() {
		<-ctx.Done()
		h.terminating.Store(true)
	}()
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	statusCode := http.StatusOK
	if h.terminating.Load() {
		statusCode = h.TerminatingStatusCode
	}
	response.WriteHeader(statusCode)
	fmt.Fprint(response, http.StatusText(statusCode))
}
