package audittypes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestApiAuditEvent(t *testing.T) {

	body := []byte("SomeData")
	ev := APIAuditEvent{}
	req := httptest.NewRequest("POST", "/some/api/resource?p1=v1", bytes.NewReader(body))
	req.Header.Set("Authorization", "auth456")

	respHdrs := http.Header{}
	respHdrs.Set("Content-Type", "text/plain")
	respInfo := types.ResponseInfo{404, 101, nil, 2048}

	ev.AppendRequest(req)
	ev.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "POST", ev.Method)
	assert.Equal(t, "/some/api/resource", ev.Path)
	assert.Equal(t, "p1=v1", ev.QueryString)
	assert.Equal(t, "auth456", ev.AuthorisationToken)
	assert.EqualValues(t, len(body), ev.RequestPayload.Get("length"))

	assert.Equal(t, "404", ev.ResponseStatus)

	assert.True(t, ev.EnforceConstraints(AuditConstraints{}))
}

func TestNewApiAudit(t *testing.T) {
	auditer := NewAPIAuditEvent("ping", "pong")
	if api, ok := auditer.(*APIAuditEvent); ok {
		assert.Equal(t, "ping", api.AuditSource)
		assert.Equal(t, "pong", api.AuditType)
	} else {
		assert.Fail(t, "Was not an APIAuditEvent")
	}
}
