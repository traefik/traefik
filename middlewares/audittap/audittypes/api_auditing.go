package audittypes

import (
	"net/http"
	"strings"

	"github.com/containous/traefik/middlewares/audittap/types"
)

// APIAuditEvent is the audit event created for API calls
type APIAuditEvent struct {
	AuditEvent
	AuthorisationToken string `json:"authorisationToken,omitempty"`
}

// AppendRequest appends information about the request to the audit event
func (ev *APIAuditEvent) AppendRequest(ctx *RequestContext, auditSpec *AuditSpecification) {
	appendCommonRequestFields(&ev.AuditEvent, ctx)
	ev.AuthorisationToken = ctx.FlatHeaders.GetString("authorization")
	if body, _, err := copyRequestBody(ctx.Req); err == nil {
		if &auditSpec.AuditObfuscation != nil {
			ct := ctx.FlatHeaders.GetString("content-type")
			var fnObfuscate func(b []byte) ([]byte, error)
			if ct == "application/x-www-form-urlencoded" {
				fnObfuscate = auditSpec.AuditObfuscation.ObfuscateURLEncoded
			} else if ct == "application/json" {
				fnObfuscate = auditSpec.AuditObfuscation.ObfuscateJSON
			}
			if fnObfuscate != nil {
				if sanitised, err := fnObfuscate(body); err == nil {
					ev.addRequestPayloadContents(strings.TrimSpace(string(sanitised)))
					ev.RequestPayload[keyPayloadLength] = len(body)
				}
			} else {
				ev.addRequestPayloadContents(string(body))
			}
		}
	}
}

// AppendResponse appends information about the response to the audit event
func (ev *APIAuditEvent) AppendResponse(responseHeaders http.Header, respInfo types.ResponseInfo, auditSpec *AuditSpecification) {
	appendCommonResponseFields(&ev.AuditEvent, responseHeaders, respInfo)
	ev.addResponsePayloadContents(strings.TrimSpace(string(respInfo.Entity)))
}

// EnforceConstraints ensures the audit event satisfies constraints
func (ev *APIAuditEvent) EnforceConstraints(constraints AuditConstraints) bool {
	enforcePrecedentConstraints(&ev.AuditEvent, constraints)
	return true
}

// ToEncoded transforms the event into an Encoded
func (ev *APIAuditEvent) ToEncoded() types.Encoded {
	return types.ToEncoded(ev)
}

// NewAPIAuditEvent creates a new APIAuditEvent with the provided auditSource and auditType
func NewAPIAuditEvent(auditSource string, auditType string) Auditer {
	ev := APIAuditEvent{}
	ev.AuditEvent = AuditEvent{AuditSource: auditSource, AuditType: auditType}
	return &ev
}
