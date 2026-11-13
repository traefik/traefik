// Package ready exposes an HTTP handler that reports the readiness of Traefik.
// Unlike the ping endpoint, which only reports that the process is alive,
// the ready endpoint returns 200 OK only after every enabled provider has
// loaded its initial configuration at least once.
package ready

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
)

// Handler exposes the ready route.
type Handler struct {
	EntryPoint            string `description:"EntryPoint" json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty" export:"true"`
	ManualRouting         bool   `description:"Manual routing" json:"manualRouting,omitempty" toml:"manualRouting,omitempty" yaml:"manualRouting,omitempty" export:"true"`
	TerminatingStatusCode int    `description:"Terminating status code" json:"terminatingStatusCode,omitempty" toml:"terminatingStatusCode,omitempty" yaml:"terminatingStatusCode,omitempty" export:"true"`

	ready       atomic.Bool
	terminating atomic.Bool
}

// SetDefaults sets the default values.
func (h *Handler) SetDefaults() {
	h.EntryPoint = "traefik"
	h.TerminatingStatusCode = http.StatusServiceUnavailable
}

// WithContext causes the ready endpoint to serve non-200 responses once ctx is canceled.
func (h *Handler) WithContext(ctx context.Context) {
	go func() {
		<-ctx.Done()
		h.terminating.Store(true)
	}()
}

// SetReady marks the handler as ready. Once called, the endpoint returns 200 OK
// (until the terminating flag is set).
func (h *Handler) SetReady() {
	h.ready.Store(true)
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, _ *http.Request) {
	statusCode := http.StatusOK
	switch {
	case h.terminating.Load():
		statusCode = h.TerminatingStatusCode
	case !h.ready.Load():
		statusCode = http.StatusServiceUnavailable
	}
	response.WriteHeader(statusCode)
	fmt.Fprint(response, http.StatusText(statusCode))
}
