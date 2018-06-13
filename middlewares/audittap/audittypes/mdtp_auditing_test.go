package audittypes

import (
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestAppendRequestFields(t *testing.T) {

	types.TheClock = T0

	ev := &MdtpAuditEvent{}
	req := httptest.NewRequest("POST", "http://my-mdtp-app.public.mdtp/some/resource?qz=abc", nil)
	req.Header.Set("X-Request-Id", "req321")
	req.Header.Set("True-Client-IP", "101.1.101.1")
	req.Header.Set("True-Client-Port", "5005")
	req.Header.Set("X-Source", "202.2.202.2")
	req.Header.Set("Request-ID", "R123")
	req.Header.Set("Session-ID", "S123")
	req.Header.Set("Akamai-Test-Hdr", "Ak999")

	ev.AppendRequest(req)

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
	assert.Equal(t, "my-mdtp-app", ev.AuditSource)
}

func TestAuditSourceDerivation(t *testing.T) {
	assert.Equal(t, "my-app", deriveAuditSource("my-app.service"))
	assert.Equal(t, "my-app", deriveAuditSource("  my-app.service  "))
	assert.Equal(t, "my-app", deriveAuditSource("my-app.protected.mdtp"))
	assert.Equal(t, "my-app", deriveAuditSource("my-app.public.mdtp"))
}
