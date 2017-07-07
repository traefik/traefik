package streams

import (
	"strings"
	"testing"
	"time"

	. "github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/stretchr/testify/assert"
)

func testSummary(clk Clock) Summary {
	t := clk.Now().UTC()
	return Summary{
		AuditSource:        "source1",
		AuditType:          "type1",
		EventID:            "1234",
		GeneratedAt:        t.Format("2006-01-02T15:04:05.000Z07:00"),
		Version:            "1",
		RequestID:          "5678",
		Method:             "GET",
		Path:               "/a/b/c",
		QueryString:        "?z=00",
		ClientIP:           "ip",
		ClientPort:         "9999",
		ReceivingIP:        "1.2.3.4",
		AuthorisationToken: "Basic foo",
		ResponseStatus:     "200",
		ResponsePayload: DataMap{
			"type": "text/plain",
		},
		ClientHeaders: DataMap{
			"accept": "*/*",
		},
		RequestHeaders: DataMap{
			"x-forwarded-for": "zzz",
		},
		ResponseHeaders: DataMap{
			"content-length": "123",
		},
		RequestPayload: DataMap{
			"type": "text/plain",
		},
	}
}

//-------------------------------------------------------------------------------------------------

func TestDirectJSONRenderer(t *testing.T) {
	enc := DirectJSONRenderer(testSummary(T0))
	assert.NoError(t, enc.Err)

	str := string(enc.Bytes)
	assert.True(t, strings.Contains(str, `"auditSource":"source1"`), str)
	assert.True(t, strings.Contains(str, `"auditType":"type1"`), str)
	assert.True(t, strings.Contains(str, `"eventID":"1234"`), str)
	assert.True(t, strings.Contains(str, `"generatedAt":"2001-09-09T01:46:40.000Z"`), str)
	assert.True(t, strings.Contains(str, `"version":"1"`), str)
	assert.True(t, strings.Contains(str, `"requestID":"5678"`), str)
	assert.True(t, strings.Contains(str, `"method":"GET"`), str)
	assert.True(t, strings.Contains(str, `"path":"/a/b/c"`), str)
	assert.True(t, strings.Contains(str, `"queryString":"?z=00"`), str)
	assert.True(t, strings.Contains(str, `"clientIP":"ip"`), str)
	assert.True(t, strings.Contains(str, `"clientPort":"9999"`), str)
	assert.True(t, strings.Contains(str, `"authorisationToken":"Basic foo"`), str)
	assert.True(t, strings.Contains(str, `"receivingIP":"1.2.3.4"`), str)
	assert.True(t, strings.Contains(str, `"responseStatus":"200"`), str)
	assert.True(t, strings.Contains(str, `"responsePayload":{"type":"text/plain"}`), str)
	assert.True(t, strings.Contains(str, `"clientHeaders":{"accept":"*/*"}`), str)
	assert.True(t, strings.Contains(str, `"requestHeaders":{"x-forwarded-for":"zzz"}`), str)
	assert.True(t, strings.Contains(str, `"requestPayload":{"type":"text/plain"}`), str)
	assert.True(t, strings.Contains(str, `"responseHeaders":{"content-length":"123"}}`), str)
}

func noopRenderer(ignored Summary) Encoded {
	return encodedJSONSample
}

func TestAuditStream(t *testing.T) {
	sink := &noopSink{0, 0}
	as := NewAuditStream(noopRenderer, sink)

	err := as.Audit(testSummary(T0))
	assert.NoError(t, err)
	assert.Equal(t, 1, sink.audits)

	err = as.Close()
	assert.NoError(t, err)
	assert.Equal(t, 1, sink.closes)
}

//-------------------------------------------------------------------------------------------------

type noopSink struct {
	audits, closes int
}

func (ns *noopSink) Audit(encoded Encoded) error {
	ns.audits++
	return nil
}

func (ns *noopSink) Close() error {
	ns.closes++
	return nil
}

//-------------------------------------------------------------------------------------------------

type fixedClock time.Time

func (c fixedClock) Now() time.Time {
	return time.Time(c)
}

var T0 = fixedClock(time.Unix(1000000000, 0))
