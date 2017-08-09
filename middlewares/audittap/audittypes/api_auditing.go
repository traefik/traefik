package audittypes

import (
	"github.com/containous/traefik/middlewares/audittap/types"
	"net/http"
)

// APIAuditEvent is the audit event created for API calls
type APIAuditEvent struct {
	AuditEvent
	AuthorisationToken string `json:"authorisationToken,omitempty"`
}

// AppendRequest appends information about the request to the audit event
func (ev *APIAuditEvent) AppendRequest(req *http.Request) {
	reqHeaders := appendCommonRequestFields(&ev.AuditEvent, req)
	ev.AuthorisationToken = reqHeaders.GetString("authorization")
}

// AppendResponse appends information about the response to the audit event
func (ev *APIAuditEvent) AppendResponse(responseHeaders http.Header, respInfo types.ResponseInfo) {
	appendCommonResponseFields(&ev.AuditEvent, responseHeaders, respInfo)
}

// ToEncoded transforms the event into an Encoded
func (ev *APIAuditEvent) ToEncoded() types.Encoded {
	return EncodeToJSON(ev)
}

// NewAPIAuditEvent creates a new APIAuditEvent with the provided auditSource and auditType
func NewAPIAuditEvent(auditSource string, auditType string) Auditer {
	ev := APIAuditEvent{}
	ev.AuditEvent = AuditEvent{AuditSource: auditSource, AuditType: auditType}
	return &ev
}
