package audittap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/containous/traefik/middlewares/audittap/configuration"
	atypes "github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

type noopAuditStream struct {
	events []interface{}
}

func (as *noopAuditStream) Audit(event atypes.Encoded) error {
	as.events = append(as.events, event)
	return nil
}

func (as *noopAuditStream) Close() error {
	return nil
}

func TestAuditTap_noop(t *testing.T) {
	atypes.TheClock = T0

	capture := &noopAuditStream{}
	cfg := &configuration.AuditSink{
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
	if apiAudit, err := toAPIAudit(capture.events[0]); err == nil {
		assert.Equal(t, "testSource", apiAudit.AuditSource)
		assert.Equal(t, "testType", apiAudit.AuditType)
		assert.Equal(t, "auth789", apiAudit.AuthorisationToken)
	} else {
		assert.Fail(t, "Audit is not an Api Audit")
	}
}

func TestInvalidProxyingForRequired(t *testing.T) {
	capture := &noopAuditStream{}
	_, err := NewAuditTap(&configuration.AuditSink{}, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.Error(t, err)
	_, err = NewAuditTap(&configuration.AuditSink{ProxyingFor: "IPA"}, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.Error(t, err)
}

func TestRateProxyingFor(t *testing.T) {
	capture := &noopAuditStream{}
	_, err := NewAuditTap(&configuration.AuditSink{ProxyingFor: "Rate"}, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
}

func TestApiProxyingFor(t *testing.T) {

	capture := &noopAuditStream{}
	cfg := &configuration.AuditSink{
		ProxyingFor: "API",
		AuditSource: "testSource",
		AuditType:   "testType",
	}
	_, err := NewAuditTap(cfg, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
}

func TestMdtpProxyingFor(t *testing.T) {
	capture := &noopAuditStream{}
	_, err := NewAuditTap(&configuration.AuditSink{ProxyingFor: "MDTP"}, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
}

func TestAuditConstraintDefaults(t *testing.T) {
	capture := &noopAuditStream{}
	tap, err := NewAuditTap(&configuration.AuditSink{ProxyingFor: "Rate"}, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
	assert.Equal(t, int64(100000), tap.AuditConfig.AuditConstraints.MaxAuditLength)
	assert.Equal(t, int64(96000), tap.AuditConfig.AuditConstraints.MaxPayloadContentsLength)
}

func TestAuditConstraintsAssigned(t *testing.T) {
	capture := &noopAuditStream{}
	conf := configuration.AuditSink{ProxyingFor: "Rate", MaxAuditLength: "3M", MaxPayloadContentsLength: "39k"}
	tap, err := NewAuditTap(&conf, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
	assert.Equal(t, int64(3000000), tap.AuditConfig.AuditConstraints.MaxAuditLength)
	assert.Equal(t, int64(39000), tap.AuditConfig.AuditConstraints.MaxPayloadContentsLength)
}

func TestOversizedAuditDropped(t *testing.T) {
	capture := &noopAuditStream{}
	cfg := &configuration.AuditSink{
		ProxyingFor:    "API",
		AuditSource:    "testSource",
		AuditType:      "testType",
		EncryptSecret:  "",
		MaxAuditLength: "50", // bytes
	}
	tap, err := NewAuditTap(cfg, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
	audit := &audittypes.APIAuditEvent{}
	payload := atypes.DataMap{"SomeKey": "IAmLongerThan10Bytes"}
	audit.RequestPayload = payload

	err = tap.submitAudit(audit)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(capture.events))
}

func TestEnforceConstraintsFailDropsAudit(t *testing.T) {
	capture := &noopAuditStream{}
	conf := configuration.AuditSink{ProxyingFor: "Rate", MaxAuditLength: "3M", MaxPayloadContentsLength: "1M"}
	tap, err := NewAuditTap(&conf, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)

	vatDecl, err := ioutil.ReadFile("audittypes/testdata/HMRC-SA-SA100-TIL.xml") // Test In Live event
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("", "/pathsegment?d=1&e=2", bytes.NewReader([]byte(vatDecl)))
	tap.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, 0, len(capture.events))
}

func TestAuditExclusion(t *testing.T) {

	capture := &noopAuditStream{}

	excludes := configuration.Exclusions{
		"Ex1": &configuration.FilterOption{HeaderName: "Host", Contains: []string{"aaaignorehost1bbb", "hostignore"}},
	}

	cfg := &configuration.AuditSink{
		ProxyingFor: "api",
		AuditSource: "as1",
		AuditType:   "at1",
		Exclusions:  excludes,
	}

	tap, err := NewAuditTap(cfg, []audittypes.AuditStream{capture}, "backend1", http.HandlerFunc(notFound))
	assert.NoError(t, err)
	tap.AuditStreams = []audittypes.AuditStream{capture}

	excReq := httptest.NewRequest("", "/pathsegment?d=1&e=2", nil)
	excReq.Host = "abchostignoredef.somedomain"
	tap.ServeHTTP(httptest.NewRecorder(), excReq)

	incReq := httptest.NewRequest("", "/includeme?d=1&e=2", nil)
	incReq.Host = "iamok.somedomain"
	tap.ServeHTTP(httptest.NewRecorder(), incReq)

	assert.Equal(t, 1, len(capture.events))
	if apiAudit, err := toAPIAudit(capture.events[0]); err == nil {
		assert.Equal(t, "as1", apiAudit.AuditSource)
		assert.Equal(t, "at1", apiAudit.AuditType)
		assert.Equal(t, "/includeme", apiAudit.Path)
	} else {
		assert.Fail(t, "Audit is not an Api Audit")
	}
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

func toAPIAudit(obj interface{}) (*audittypes.APIAuditEvent, error) {
	if enc, ok := obj.(atypes.Encoded); ok {
		audit := &audittypes.APIAuditEvent{}
		err := json.Unmarshal(enc.Bytes, audit)
		return audit, err
	}
	return nil, errors.New("obj is expected to be type Encoded")
}
