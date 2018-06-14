package audittypes

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/containous/traefik/middlewares/audittap/types"
)

// MdtpAuditEvent is the audit event created for API calls
type MdtpAuditEvent struct {
	AuditEvent
	AuthorisationToken string
}

// AppendRequest appends information about the request to the audit event
func (ev *MdtpAuditEvent) AppendRequest(req *http.Request) {
	//reqHeaders := appendCommonRequestFields(&ev.AuditEvent, req)
	ev.AuditSource = deriveAuditSource(req.Host)
	ev.AuditType = "RequestReceived"
	reqHeaders := appendCommonRequestFields(&ev.AuditEvent, req)
	ev.AuthorisationToken = reqHeaders.GetString("authorization")
	if body, _, err := copyRequestBody(req); err == nil {
		ev.addRequestPayloadContents(string(body))
	}
}

// AppendResponse appends information about the response to the audit event
func (ev *MdtpAuditEvent) AppendResponse(responseHeaders http.Header, respInfo types.ResponseInfo) {
	appendCommonResponseFields(&ev.AuditEvent, responseHeaders, respInfo)
	ev.addResponsePayloadContents(strings.TrimSpace(string(respInfo.Entity)))
}

// EnforceConstraints ensures the audit event satisfies constraints
func (ev *MdtpAuditEvent) EnforceConstraints(constraints AuditConstraints) bool {
	enforcePrecedentConstraints(&ev.AuditEvent, constraints)
	return true
}

// ToEncoded transforms the event into an Encoded
func (ev *MdtpAuditEvent) ToEncoded() types.Encoded {
	return types.ToEncoded(ev)
}

// NewAPIAuditEvent creates a new APIAuditEvent with the provided auditSource and auditType
func NewMdtpAuditEvent() Auditer {
	ev := APIAuditEvent{AuditEvent: AuditEvent{}}
	return &ev
}

func deriveAuditSource(s string) string {
	re := regexp.MustCompile("^(.*)\\.(service|public\\.mdtp|protected\\.mdtp)$")
	matches := re.FindStringSubmatch(strings.TrimSpace(s))
	auditSource := s
	if len(matches) == 3 {
		auditSource = matches[1]
	}
	return auditSource
}
