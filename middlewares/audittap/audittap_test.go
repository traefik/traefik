package audittap

import (
	"fmt"
	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type noopAuditStream struct {
	events []audittypes.Summary
}

func (as *noopAuditStream) Audit(summary audittypes.Summary) error {
	as.events = append(as.events, summary)
	return nil
}

func (as *noopAuditStream) Close() error {
	return nil
}

func TestAuditTap_noop(t *testing.T) {
	audittypes.TheClock = T0

	capture := &noopAuditStream{}
	cfg := &types.AuditSink{}
	tap, err := NewAuditTap(cfg, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	tap.AuditStreams = []audittypes.AuditStream{capture}
	assert.NoError(t, err)

	req := httptest.NewRequest("", "/a/b/c?d=1&e=2", nil)
	req.RemoteAddr = "101.102.103.104:1234"
	req.Host = "example.co.uk"
	req.Header.Set("Request-ID", "R123")
	req.Header.Set("Session-ID", "S123")
	res := httptest.NewRecorder()

	tap.ServeHTTP(res, req)

	assert.Equal(t, 1, len(capture.events))
	assert.Equal(t,
		audittypes.Summary{
			"backend1",
			audittypes.DataMap{
				audittypes.Host:       "example.co.uk",
				audittypes.Method:     "GET",
				audittypes.Path:       "/a/b/c",
				audittypes.Query:      "d=1&e=2",
				audittypes.RemoteAddr: "101.102.103.104:1234",
				"hdr-request-id":      "R123",
				"hdr-session-id":      "S123",
				audittypes.BeganAt:    audittypes.TheClock.Now().UTC(),
			},
			audittypes.DataMap{
				audittypes.Status:            404,
				"hdr-x-content-type-options": "nosniff",
				"hdr-content-type":           "text/plain; charset=utf-8",
				audittypes.Size:              19,
				audittypes.Entity:            []byte("404 page not found\n"),
				audittypes.CompletedAt:       audittypes.TheClock.Now().UTC(),
			},
		},
		capture.events[0])
}

// simpleHandler replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be plain text.
func simpleHandler(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, error)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	simpleHandler(w, "404 page not found", http.StatusNotFound)
}
