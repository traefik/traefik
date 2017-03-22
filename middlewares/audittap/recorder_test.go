package audittap

import (
	. "github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuditResponseWriter_no_body(t *testing.T) {
	TheClock = T0

	recorder := httptest.NewRecorder()
	w := NewAuditResponseWriter(recorder, MaximumEntityLength)
	w.WriteHeader(204)
	assert.Equal(t, 204, w.SummariseResponse()[Status])
	assert.Equal(t, 0, w.SummariseResponse()[Size])
}

func TestAuditResponseWriter_with_body(t *testing.T) {
	TheClock = T0

	recorder := httptest.NewRecorder()
	w := NewAuditResponseWriter(recorder, MaximumEntityLength)
	w.WriteHeader(200)
	w.Write([]byte("hello"))
	w.Write([]byte("world"))
	assert.Equal(t, 200, w.SummariseResponse()[Status])
	assert.Equal(t, 10, w.SummariseResponse()[Size])
}

func TestAuditResponseWriter_headers(t *testing.T) {
	TheClock = T0

	recorder := httptest.NewRecorder()
	w := NewAuditResponseWriter(recorder, MaximumEntityLength)

	// hop-by-hop headers should be dropped
	w.Header().Set("Keep-Alive", "true")
	w.Header().Set("Connection", "1")
	w.Header().Set("Proxy-Authenticate", "1")
	w.Header().Set("Proxy-Authorization", "1")
	w.Header().Set("TE", "1")
	w.Header().Set("Trailers", "1")
	w.Header().Set("Transfer-Encoding", "1")
	w.Header().Set("Upgrade", "1")

	// other headers should be retainedd
	w.Header().Set("Content-Length", "123")
	w.Header().Set("Request-ID", "abc123")
	w.Header().Add("Cookie", "a=1; b=2")
	w.Header().Add("Cookie", "c=3")

	assert.Equal(t,
		DataMap{
			"hdr-content-length": "123",
			"hdr-request-id":     "abc123",
			"hdr-cookie":         []string{"a=1", "b=2", "c=3"},
			CompletedAt:          T0.Now().UTC(),
			Status:               0,
			Size:                 0,
			Entity:               []byte{},
		},
		w.SummariseResponse())
}

type fixedClock time.Time

func (c fixedClock) Now() time.Time {
	return time.Time(c)
}

var T0 = fixedClock(time.Unix(1000000000, 0))
