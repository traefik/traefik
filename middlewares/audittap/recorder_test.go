package audittap

import (
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/stretchr/testify/assert"
)

func TestAuditResponseWriter_with_body(t *testing.T) {
	TheClock = T0

	recorder := httptest.NewRecorder()
	w := NewAuditResponseWriter(recorder, MaximumEntityLength)
	w.WriteHeader(200)
	w.Write([]byte("hello"))
	w.Write([]byte("world"))

	var s = Summary{}
	w.SummariseResponse(&s)

	assert.Equal(t, "200", s.ResponseStatus)
}

func TestAuditResponseWriter_headers(t *testing.T) {
	TheClock = T0

	recorder := httptest.NewRecorder()
	w := NewAuditResponseWriter(recorder, MaximumEntityLength)

	// hop-by-hop headers should be retained
	w.Header().Set("Keep-Alive", "true")
	w.Header().Set("Connection", "1")
	w.Header().Set("Proxy-Authenticate", "1")
	w.Header().Set("Proxy-Authorization", "1")
	w.Header().Set("TE", "1")
	w.Header().Set("Trailers", "1")
	w.Header().Set("Transfer-Encoding", "1")
	w.Header().Set("Upgrade", "1")

	// other headers should be retained
	w.Header().Set("Content-Length", "123")
	w.Header().Set("Request-ID", "abc123")
	w.Header().Add("Cookie", "a=1; b=2")
	w.Header().Add("Cookie", "c=3")

	// content-type should be set under responsePayload
	w.Header().Add("Content-Type", "application/json")

	var s = Summary{}
	w.SummariseResponse(&s)

	assert.Equal(t,
		Summary{
			AuditSource:        "",
			AuditType:          "",
			EventID:            "",
			GeneratedAt:        "",
			Version:            "",
			RequestID:          "",
			Method:             "",
			Path:               "",
			QueryString:        "",
			ClientIP:           "",
			ClientPort:         "",
			ReceivingIP:        "",
			AuthorisationToken: "",
			ResponseStatus:     "0",
			ResponsePayload:    DataMap{"type": "application/json"},
			ClientHeaders:      nil,
			RequestHeaders:     nil,
			RequestPayload:     nil,
			ResponseHeaders: DataMap{
				"trailers":            "1",
				"proxy-authenticate":  "1",
				"cookie":              []string{"a=1", "b=2", "c=3"},
				"te":                  "1",
				"request-id":          "abc123",
				"content-length":      "123",
				"transfer-encoding":   "1",
				"proxy-authorization": "1",
				"connection":          "1",
				"upgrade":             "1",
				"keep-alive":          "true"},
		},
		s)
}

type fixedClock time.Time

func (c fixedClock) Now() time.Time {
	return time.Time(c)
}

var T0 = fixedClock(time.Unix(1000000000, 0))
