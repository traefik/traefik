package streams

import (
	. "github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func testSummary(clk Clock) Summary {
	t := clk.Now().UTC()
	return Summary{
		Source: "source1",
		Request: DataMap{
			Host:       "host.com",
			Method:     "GET",
			Path:       "/a/b/c",
			Query:      "?z=00",
			RemoteAddr: "10.11.12.13:12345",
			BeganAt:    t,
		},
		Response: DataMap{
			Status:      200,
			Size:        123,
			CompletedAt: t.Add(time.Millisecond),
		},
	}
}

//-------------------------------------------------------------------------------------------------

func TestDirectJSONRenderer(t *testing.T) {
	enc := DirectJSONRenderer(testSummary(T0))
	assert.NoError(t, enc.Err)

	str := string(enc.Bytes)
	assert.True(t, strings.HasPrefix(str, `{"auditSource":"source1","request":{`), str)
	assert.True(t, strings.HasSuffix(str, `}}`), str)
	assert.True(t, strings.Contains(str, `,"response":`), str)
	p := strings.Index(str, `"response":`)
	request := str[0:p]
	response := str[p:]

	// the order of the map keys is unspecified, so check each item one by one
	assert.True(t, strings.Contains(request, `"host":"host.com"`), request)
	assert.True(t, strings.Contains(request, `"method":"GET"`), request)
	assert.True(t, strings.Contains(request, `"path":"/a/b/c"`), request)
	assert.True(t, strings.Contains(request, `"query":"?z=00"`), request)
	assert.True(t, strings.Contains(request, `"remoteAddr":"10.11.12.13:12345"`), request)
	assert.True(t, strings.Contains(request, `"beganAt":"2001-09-09T01:46:40Z"`), request)

	assert.True(t, strings.Contains(response, `"status":200`), response)
	assert.True(t, strings.Contains(response, `"size":123`), response)
	assert.True(t, strings.Contains(response, `"completedAt":"2001-09-09T01:46:40.001Z"`), response)
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
