package audittypes

import (
	"github.com/containous/traefik/middlewares/audittap/types"
	"net/http"
)

type ApiAuditEvent struct {
	AuditEvent
	AuthorisationToken string `json:"authorisationToken,omitempty"`
}

func (ev *ApiAuditEvent) AppendRequest(req *http.Request) {
	reqHeaders := appendCommonRequestFields(&ev.AuditEvent, req)
	ev.AuthorisationToken = reqHeaders.GetString("authorization")
}

func (ev *ApiAuditEvent) AppendResponse(responseHeaders http.Header, respInfo types.ResponseInfo) {
	appendCommonResponseFields(&ev.AuditEvent, responseHeaders, respInfo)
}

func (ev *ApiAuditEvent) ToEncoded() types.Encoded {
	return EncodeToJSON(ev)
}

func NewApiAuditEvent(auditSource string, auditType string) Auditer {
	ev := ApiAuditEvent{}
	ev.AuditEvent = AuditEvent{AuditSource: auditSource, AuditType: auditType}
	return &ev
}
