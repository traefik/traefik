package audittypes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestMdtpAuditEvent(t *testing.T) {

	types.TheClock = T0

	ev := &MdtpAuditEvent{}
	requestBody := "say=Hi&to=Dave"
	req := httptest.NewRequest("POST", "http://my-mdtp-app.public.mdtp/some/resource?p1=v1", strings.NewReader(requestBody))
	req.Header.Set("Authorization", "auth456")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseBody := "Some response message"
	respHdrs := http.Header{}
	respHdrs.Set("Content-Type", "text/plain")
	respInfo := types.ResponseInfo{200, 101, []byte(responseBody), 2048}

	ev.AppendRequest(req)
	ev.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "my-mdtp-app", ev.AuditSource)
	assert.Equal(t, "POST", ev.Method)
	assert.Equal(t, "/some/resource", ev.Path)
	assert.Equal(t, "p1=v1", ev.QueryString)
	assert.Equal(t, "auth456", ev.AuthorisationToken)

	assert.EqualValues(t, len(requestBody), ev.RequestPayload.Get("length"))
	assert.Equal(t, string(requestBody), ev.RequestPayload["contents"])

	assert.EqualValues(t, len(responseBody), ev.ResponsePayload.Get("length"))
	assert.Equal(t, string(responseBody), ev.ResponsePayload["contents"])

	assert.Equal(t, "200", ev.ResponseStatus)

	assert.True(t, ev.EnforceConstraints(AuditConstraints{}))
}

func TestAuditSourceDerivation(t *testing.T) {
	assert.Equal(t, "my-app", deriveAuditSource("my-app.service"))
	assert.Equal(t, "my-app", deriveAuditSource("  my-app.service  "))
	assert.Equal(t, "my-app", deriveAuditSource("my-app.protected.mdtp"))
	assert.Equal(t, "my-app", deriveAuditSource("my-app.public.mdtp"))
}
