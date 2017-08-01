package audittypes

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/containous/traefik/middlewares/audittap/types"
	"gopkg.in/beevik/etree.v0"
)

type RateAuditDetail struct {
	CorrelationID   string `json:"correlationID,omitempty"`
	Email           string `json:"email,omitempty"`
	RequestType     string `json:"requestType,omitempty"`
	Role            string `json:"role,omitempty"`
	SenderID        string `json:"senderID,omitempty"`
	SoftwareFamily  string `json:"softwareFamily,omitempty"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
	TransactionID   string `json:"transactionID,omitempty"`
	UserType        string `json:"userType,omitempty"`
}

type RateAuditEvent struct {
	AuditEvent
	Detail      RateAuditDetail `json:"detail,omitempty"`
	Identifiers types.DataMap   `json:"identifiers,omitempty"`
	Enrolments  types.DataMap   `json:"enrolments,omitempty"`
}

type XMLFragment struct {
	InnerXML []byte `xml:",innerxml"`
}

type PartialGovTalkMessage struct {
	Header  *etree.Document
	Details *etree.Document
}

func (ev *RateAuditEvent) AppendRequest(req *http.Request) {
	appendCommonRequestFields(&ev.AuditEvent, req)
	if partialMsg, err := getMessageParts(req.Body); err == nil {
		extractIfPresent(partialMsg.Header, xpClass, &ev.AuditType)
		ev.populateDetail(partialMsg)
		ev.populateIdentifiers(partialMsg)
		ev.populateEnrolments(partialMsg)
	}
}

func (ev *RateAuditEvent) AppendResponse(responseHeaders http.Header, respInfo types.ResponseInfo) {
	appendCommonResponseFields(&ev.AuditEvent, responseHeaders, respInfo)
}

func (ra *RateAuditEvent) ToEncoded() types.Encoded {
	return EncodeToJSON(ra)
}

func NewRateAuditEvent() Auditer {
	ev := RateAuditEvent{}
	ev.AuditEvent = AuditEvent{AuditSource: "transaction-engine-frontend"}
	return &ev
}

func getMessageParts(body io.Reader) (*PartialGovTalkMessage, error) {
	decoder := xml.NewDecoder(body)
	var header XMLFragment
	var details XMLFragment
	haveHeader := false
	haveDetails := false

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Header" {
				if err := decoder.DecodeElement(&header, &se); err == nil {
					haveHeader = true
				}
			} else if se.Name.Local == "GovTalkDetails" {
				if err := decoder.DecodeElement(&details, &se); err == nil {
					haveDetails = true
				}
			}
		}

		// Stop parsing when required data is obtained
		if haveHeader && haveDetails {
			partial := PartialGovTalkMessage{}
			headerDoc := etree.NewDocument()
			if err := headerDoc.ReadFromBytes(header.InnerXML); err != nil {
				return nil, err
			}
			detailsDoc := etree.NewDocument()
			if err := detailsDoc.ReadFromBytes(details.InnerXML); err != nil {
				return nil, err
			}
			partial.Header = headerDoc
			partial.Details = detailsDoc
			return &partial, nil
		}
	}

	return nil, errors.New("Unexpected message structure. Headers/GovTalkDetails not present")
}

// Headers XPaths
var xpClass = etree.MustCompilePath("./MessageDetails/Class")
var xpCorrelationID = etree.MustCompilePath("./MessageDetails/CorrelationID")
var xpTransactionID = etree.MustCompilePath("./MessageDetails/TransactionID")
var xpEmailAddress = etree.MustCompilePath("./SenderDetails/EmailAddress")
var xpSenderID = etree.MustCompilePath("./SenderDetails/IDAuthentication/SenderID")

// GovTalkDetails XPaths
var xpAgentGroupCode = etree.MustCompilePath("./GatewayAdditions/Submitter/AgentDetails/AgentGroupCode")
var xpAgentEnrolments = etree.MustCompilePath("./GatewayAdditions/Submitter/AgentDetails/AgentEnrolments/Enrolment/ServiceName")
var xpAgentEnrolmentIds = etree.MustCompilePath("./Identifiers/Identifier")
var xpCredentialID = etree.MustCompilePath("./GatewayAdditions/Submitter/SubmitterDetails/CredentialIdentifier")
var xpSoftwareFamily = etree.MustCompilePath("./ChannelRouting/Channel/Product")
var xpSoftwareVersion = etree.MustCompilePath("./ChannelRouting/Channel/Version")
var xpIdentifiers = etree.MustCompilePath("./Keys/Key")
var xpRole = etree.MustCompilePath("./GatewayAdditions/Submitter/SubmitterDetails/CredentialRole")
var xpUserType = etree.MustCompilePath("./GatewayAdditions/Submitter/SubmitterDetails/RegistrationCategory")

func (ev *RateAuditEvent) populateDetail(partial *PartialGovTalkMessage) {
	extractIfPresent(partial.Header, xpCorrelationID, &ev.Detail.CorrelationID)
	extractIfPresent(partial.Header, xpTransactionID, &ev.Detail.TransactionID)
	extractIfPresent(partial.Header, xpSenderID, &ev.Detail.SenderID)
	extractIfPresent(partial.Header, xpEmailAddress, &ev.Detail.Email)
	extractIfPresent(partial.Details, xpRole, &ev.Detail.Role)
	extractIfPresent(partial.Details, xpSoftwareFamily, &ev.Detail.SoftwareFamily)
	extractIfPresent(partial.Details, xpSoftwareVersion, &ev.Detail.SoftwareVersion)
	extractIfPresent(partial.Details, xpUserType, &ev.Detail.UserType)

	ev.Detail.RequestType = translateRequestType(partial)

}

func (ev *RateAuditEvent) populateIdentifiers(partial *PartialGovTalkMessage) {

	ev.Identifiers = types.DataMap{}

	if node := partial.Details.FindElementPath(xpCredentialID); node != nil {
		ev.Identifiers["credID"] = node.Text()
	}

	if node := partial.Details.FindElementPath(xpAgentGroupCode); node != nil {
		ev.Identifiers["agentGroupCode"] = node.Text()
	}

	if ids := partial.Details.FindElementsPath(xpIdentifiers); len(ids) > 0 {
		for _, el := range ids {
			ev.Identifiers[el.SelectAttr("Type").Value] = el.Text()
		}
	}

}

func (ev *RateAuditEvent) populateEnrolments(partial *PartialGovTalkMessage) {

	if nodes := partial.Details.FindElementsPath(xpAgentEnrolments); len(nodes) > 0 {
		ev.Enrolments = types.DataMap{}
		for _, service := range nodes {
			enrolmentIds := types.DataMap{}
			for _, ids := range service.Parent().FindElementsPath(xpAgentEnrolmentIds) {
				enrolmentIds[ids.SelectAttr("IdentifierType").Value] = ids.Text()
			}
			ev.Enrolments[service.Text()] = enrolmentIds
		}
	}

}

func extractIfPresent(doc *etree.Document, path etree.Path, dest *string) {
	if node := doc.FindElementPath(path); node != nil {
		*dest = node.Text()
	}
}

var pathFunction = etree.MustCompilePath("./MessageDetails/Function")
var pathQualifier = etree.MustCompilePath("./MessageDetails/Qualifier")

func translateRequestType(partial *PartialGovTalkMessage) string {
	var function string
	var qualifier string
	extractIfPresent(partial.Header, pathFunction, &function)
	extractIfPresent(partial.Header, pathQualifier, &qualifier)

	if function != "" && qualifier != "" {
		var translated = function
		if translated == "submit" {
			translated = "submission"
		} else if translated == "list" {
			translated = "data"
		}
		return strings.ToUpper(translated + "_" + qualifier)
	}

	return ""
}
