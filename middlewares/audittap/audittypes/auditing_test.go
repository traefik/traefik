package audittypes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"strconv"

	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestAppendCommonRequestFields(t *testing.T) {

	types.TheClock = T0

	var ev = &AuditEvent{}
	var req = httptest.NewRequest("POST", "/some/resource?qz=abc", nil)
	req.Header.Set("X-Request-Id", "req321")
	req.Header.Set("True-Client-IP", "101.1.101.1")
	req.Header.Set("True-Client-Port", "5005")
	req.Header.Set("X-Source", "202.2.202.2")
	req.Header.Set("Request-ID", "R123")
	req.Header.Set("Session-ID", "S123")
	req.Header.Set("Akamai-Test-Hdr", "Ak999")

	returned := appendCommonRequestFields(ev, req)

	assert.NotEmpty(t, ev.EventID)
	assert.Equal(t, "2001-09-09T01:46:40.000Z", ev.GeneratedAt)
	assert.Equal(t, "1", ev.Version)
	assert.Equal(t, "POST", ev.Method)
	assert.Equal(t, "/some/resource", ev.Path)
	assert.Equal(t, "qz=abc", ev.QueryString)
	assert.Equal(t, "req321", ev.RequestID)
	assert.Equal(t, "101.1.101.1", ev.ClientIP)
	assert.Equal(t, "5005", ev.ClientPort)
	assert.Equal(t, "202.2.202.2", ev.ReceivingIP)
	assert.Equal(t, types.DataMap{"session-id": "S123", "request-id": "R123"}, ev.ClientHeaders)
	assert.Equal(t, types.DataMap{"akamai-test-hdr": "Ak999"}, ev.RequestHeaders)
	assert.Empty(t, ev.ResponseStatus)
	assert.Nil(t, ev.ResponseHeaders)
	assert.Nil(t, ev.ResponsePayload)
	assert.Len(t, returned, 7)
	assert.Equal(t, "202.2.202.2", returned.GetString("x-source"))
}

func TestAppendCommonResponseFields(t *testing.T) {

	var ev = &AuditEvent{}
	var respHeaders = http.Header{}
	respHeaders.Set("Content-Type", "text/plain")
	respHeaders.Set("Some-Hdr-X", "X99")
	respHeaders.Add("Cookie", "a=1; b=2")
	respHeaders.Add("Cookie", "c=3")

	respInfo := types.ResponseInfo{200, 99, nil, 2048}

	returned := appendCommonResponseFields(ev, respHeaders, respInfo)

	expectRespHdrs := types.DataMap{
		"cookie":     []string{"a=1", "b=2", "c=3"},
		"some-hdr-x": "X99",
	}

	assert.Equal(t, strconv.Itoa(respInfo.Status), ev.ResponseStatus)
	assert.Equal(t, types.DataMap{"type": "text/plain"}, ev.ResponsePayload)
	assert.Equal(t, expectRespHdrs, ev.ResponseHeaders)
	assert.Equal(t, types.DataMap{"cookie": []string{"a=1", "b=2", "c=3"}, "content-type": "text/plain", "some-hdr-x": "X99"}, returned)
}

func TestAuditResponseHeaders(t *testing.T) {

	var ev = &AuditEvent{}
	var respHeaders = http.Header{}

	// hop-by-hop headers should be retained
	respHeaders.Set("Keep-Alive", "true")
	respHeaders.Set("Connection", "1")
	respHeaders.Set("Proxy-Authenticate", "1")
	respHeaders.Set("Proxy-Authorization", "1")
	respHeaders.Set("TE", "1")
	respHeaders.Set("Trailers", "1")
	respHeaders.Set("Transfer-Encoding", "1")
	respHeaders.Set("Upgrade", "1")

	// other headers should be retained
	respHeaders.Set("Content-Length", "123")
	respHeaders.Set("Request-ID", "abc123")
	respHeaders.Add("Cookie", "a=1; b=2")
	respHeaders.Add("Cookie", "c=3")

	// content-type should be set under responsePayload
	respHeaders.Add("Content-Type", "application/json")

	appendCommonResponseFields(ev, respHeaders, types.ResponseInfo{200, 99, nil, 2048})

	expectRespHdrs := types.DataMap{
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
		"keep-alive":          "true"}

	assert.Equal(t, types.DataMap{"type": "application/json"}, ev.ResponsePayload)
	assert.Equal(t, expectRespHdrs, ev.ResponseHeaders)
}

func TestRequestContentsOmittedWhenTooLong(t *testing.T) {
	max := 20
	ev := AuditEvent{RequestPayload: types.DataMap{}}
	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(max)}
	ev.RequestPayload[keyPayloadContents] = types.DataMap{"Key1": "MoreThan20Bytes"}
	ev.RequestPayload[keyPayloadLength] = max + 1
	enforcePrecedentConstraints(&ev, constraints)
	assert.Nil(t, ev.RequestPayload[keyPayloadContents])
}

func TestResponseContentsOmittedWhenTooLong(t *testing.T) {
	max := 20
	ev := AuditEvent{ResponsePayload: types.DataMap{}}
	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(max)}
	ev.ResponsePayload[keyPayloadContents] = types.DataMap{"Key1": "MoreThan20Bytes"}
	ev.ResponsePayload[keyPayloadLength] = max + 1
	enforcePrecedentConstraints(&ev, constraints)
	assert.Nil(t, ev.ResponsePayload[keyPayloadContents])
}

func TestResponseContentsOmittedWhenResponseAndRequestTooLong(t *testing.T) {
	requestPayload := types.DataMap{"Key1": "AllowedRequestSize"}
	max := 40
	ev := AuditEvent{RequestPayload: types.DataMap{}, ResponsePayload: types.DataMap{}}
	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(max)}
	ev.RequestPayload[keyPayloadContents] = requestPayload
	ev.RequestPayload[keyPayloadLength] = max / 2
	ev.ResponsePayload[keyPayloadContents] = types.DataMap{"Key1": "DisallowedResponseSize"}
	ev.ResponsePayload[keyPayloadLength] = (max / 2) + 1
	enforcePrecedentConstraints(&ev, constraints)
	assert.Nil(t, ev.ResponsePayload[keyPayloadContents])
	assert.Equal(t, requestPayload, ev.RequestPayload[keyPayloadContents])
}

func TestRequestContentsRetained(t *testing.T) {
	max := 1000
	contents := types.DataMap{"Key1": "MoreThan20Bytes"}
	ev := RATEAuditEvent{}
	ev.AuditEvent = AuditEvent{RequestPayload: types.DataMap{}}
	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(max)}
	ev.RequestPayload[keyPayloadContents] = contents
	ev.RequestPayload[keyPayloadLength] = max - 1
	enforcePrecedentConstraints(&ev.AuditEvent, constraints)
	assert.Equal(t, contents, ev.RequestPayload[keyPayloadContents])
	assert.Nil(t, ev.ResponsePayload[keyPayloadContents])
}

func TestResponseContentsRetained(t *testing.T) {
	max := 1000
	contents := types.DataMap{"Key1": "MoreThan20Bytes"}
	ev := RATEAuditEvent{}
	ev.AuditEvent = AuditEvent{RequestPayload: types.DataMap{}, ResponsePayload: types.DataMap{}}
	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(max)}
	ev.RequestPayload[keyPayloadLength] = max + 1
	ev.ResponsePayload[keyPayloadContents] = contents
	ev.ResponsePayload[keyPayloadLength] = max - 1
	enforcePrecedentConstraints(&ev.AuditEvent, constraints)
	assert.Nil(t, ev.RequestPayload[keyPayloadContents])
	assert.Equal(t, contents, ev.ResponsePayload[keyPayloadContents])
}

func TestAuditObfuscateUrlEncoded(t *testing.T) {
	obs := AuditObfuscation{MaskValue: "@++@", MaskFields: []string{"x1"}}

	masked, err := obs.ObfuscateURLEncoded([]byte("start=s1&x1=kkdw09dkwad&def=d1&x1=wdoueqoi2ej&end=e1"))
	assert.NoError(t, err)
	assert.Equal(t, "start=s1&x1=@++@&def=d1&x1=@++@&end=e1", string(masked))

	masked, err = obs.ObfuscateURLEncoded([]byte("x1=aefaef&d1=dere%20e&x1=wdawdwwd&d2=ziefjef&x1=brerber"))
	assert.NoError(t, err)
	assert.Equal(t, "x1=@++@&d1=dere%20e&x1=@++@&d2=ziefjef&x1=@++@", string(masked))
}

type fixedClock time.Time

func (c fixedClock) Now() time.Time {
	return time.Time(c)
}

var T0 = fixedClock(time.Unix(1000000000, 0))
