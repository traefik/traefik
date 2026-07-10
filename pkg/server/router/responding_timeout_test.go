package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/middlewares/recovery"
	"github.com/traefik/traefik/v3/pkg/observability/tracing"
	"go.opentelemetry.io/otel/trace/noop"
)

type deadlineRecorder struct {
	http.ResponseWriter

	readDeadline  time.Time
	writeDeadline time.Time
}

func (d *deadlineRecorder) SetReadDeadline(deadline time.Time) error {
	d.readDeadline = deadline
	return nil
}

func (d *deadlineRecorder) SetWriteDeadline(deadline time.Time) error {
	d.writeDeadline = deadline
	return nil
}

// TestRespondingTimeoutDeadlineReachesConnection ensures the router respondingTimeouts can arm the
// connection deadlines through the whole writer chain sitting above its insertion point.
// The chain is assembled with tracing enabled, the configuration that engages the observability
// statusCodeRecorder: a wrapper dropping Unwrap would make http.ResponseController fail silently.
func TestRespondingTimeoutDeadlineReachesConnection(t *testing.T) {
	base := &deadlineRecorder{ResponseWriter: httptest.NewRecorder()}
	deadline := time.Now().Add(time.Minute)

	// The router handler position, where the respondingtimeout middleware arms the deadlines.
	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rc := http.NewResponseController(rw)
		assert.NoError(t, rc.SetReadDeadline(deadline))
		assert.NoError(t, rc.SetWriteDeadline(deadline))
	})

	// recovery wraps the whole muxer, while capture and the tracing entry point come from the per-router
	// observability chain: together they are exactly the writer chain above the respondingtimeout insertion point.
	chain := alice.New(
		func(next http.Handler) (http.Handler, error) { return recovery.New(t.Context(), next) },
		capture.Wrap,
		observability.EntryPointHandler(t.Context(), tracing.NewTracer(noop.Tracer{}, nil, nil, nil), "web"),
	)

	handler, err := chain.Then(next)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://localhost/", http.NoBody)
	// Tracing must be enabled for the entry point handler to engage the statusCodeRecorder.
	req = req.WithContext(observability.WithObservability(req.Context(), observability.Observability{TracingEnabled: true}))

	handler.ServeHTTP(base, req)

	assert.Equal(t, deadline, base.readDeadline)
	assert.Equal(t, deadline, base.writeDeadline)
}
