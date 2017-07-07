package audittap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
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
	cfg := &types.AuditSink{
		AuditSource: "testSource",
		AuditType:   "testType",
	}
	tap, err := NewAuditTap(cfg, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
	tap.AuditStreams = []audittypes.AuditStream{capture}

	req := httptest.NewRequest("", "/a/b/c?d=1&e=2", nil)
	req.RemoteAddr = "101.102.103.104:1234"
	req.Host = "example.co.uk"
	req.Header.Set("Request-ID", "R123")
	req.Header.Set("Session-ID", "S123")
	res := httptest.NewRecorder()

	tap.ServeHTTP(res, req)

	capture.events[0].EventID = ""
	capture.events[0].RequestID = ""

	assert.Equal(t, 1, len(capture.events))
	assert.Equal(t,
		audittypes.Summary{
			AuditSource:        "testSource",
			AuditType:          "testType",
			GeneratedAt:        "2001-09-09T01:46:40.000Z",
			Version:            "1",
			RequestID:          "",
			Method:             "GET",
			Path:               "/a/b/c",
			QueryString:        "d=1&e=2",
			ClientIP:           "",
			ClientPort:         "",
			ReceivingIP:        "",
			AuthorisationToken: "",
			ResponseStatus:     "404",
			ResponsePayload: audittypes.DataMap{
				"type": "text/plain; charset=utf-8",
			},
			ClientHeaders: audittypes.DataMap{
				"session-id": "S123",
				"request-id": "R123",
			},
			RequestHeaders: audittypes.DataMap{},
			RequestPayload: audittypes.DataMap{
				"type": "",
			},
			ResponseHeaders: audittypes.DataMap{
				"x-content-type-options": "nosniff",
			},
		}, capture.events[0])
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
