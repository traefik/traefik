package audittap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/middlewares/audittap/audittypes"
	atypes "github.com/containous/traefik/middlewares/audittap/types"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

type noopAuditStream struct {
	events []interface{}
}

func (as *noopAuditStream) Audit(event atypes.Encodeable) error {
	as.events = append(as.events, event)
	return nil
}

func (as *noopAuditStream) Close() error {
	return nil
}

func TestAuditTap_noop(t *testing.T) {
	atypes.TheClock = T0

	capture := &noopAuditStream{}
	cfg := &types.AuditSink{
		ProxyingFor:   "API",
		AuditSource:   "testSource",
		AuditType:     "testType",
		EncryptSecret: "",
	}
	tap, err := NewAuditTap(cfg, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
	tap.AuditStreams = []audittypes.AuditStream{capture}

	req := httptest.NewRequest("", "/a/b/c?d=1&e=2", nil)
	req.Header.Set("Authorization", "auth789")

	res := httptest.NewRecorder()

	tap.ServeHTTP(res, req)

	assert.Equal(t, 1, len(capture.events))
	if apiAudit, ok := capture.events[0].(*audittypes.APIAuditEvent); ok {
		assert.Equal(t, "testSource", apiAudit.AuditSource)
		assert.Equal(t, "testType", apiAudit.AuditType)
		assert.Equal(t, "auth789", apiAudit.AuthorisationToken)
	} else {
		assert.Fail(t, "Audit is not an Api Audit")
	}
}

func TestInvalidProxyingForRequired(t *testing.T) {
	capture := &noopAuditStream{}
	_, err := NewAuditTap(&types.AuditSink{}, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.Error(t, err)
	_, err = NewAuditTap(&types.AuditSink{ProxyingFor: "IPA"}, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.Error(t, err)
}

func TestRateProxyingFor(t *testing.T) {
	capture := &noopAuditStream{}
	_, err := NewAuditTap(&types.AuditSink{ProxyingFor: "Rate"}, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
}

func TestApiProxyingFor(t *testing.T) {

	capture := &noopAuditStream{}
	cfg := &types.AuditSink{
		ProxyingFor: "API",
		AuditSource: "testSource",
		AuditType:   "testType",
	}
	_, err := NewAuditTap(cfg, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
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
