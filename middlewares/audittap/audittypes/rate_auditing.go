package audittypes

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"fmt"

	"github.com/beevik/etree"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/containous/traefik/middlewares/audittap/xmlutils"
)

// RATEAuditDetail is the detail section of the event
type RATEAuditDetail struct {
	CorrelationID   string `json:"correlationID,omitempty"`
	Email           string `json:"email,omitempty"`
	RequestType     string `json:"requestType,omitempty"`
	Role            string `json:"role,omitempty"`
	SenderID        string `json:"senderID,omitempty"`
	SoftwareFamily  string `json:"softwareFamily,omitempty"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
	TransactionID   string `json:"transactionID,omitempty"`
	UserType        string `json:"userType,omitempty"`
	IsRepayment     string `json:"isRepayment,omitempty"` // Boolean values as string so it can be empty
}

// RATEAuditEvent is the audit event created for RATE calls
type RATEAuditEvent struct {
	AuditEvent
	Detail      RATEAuditDetail `json:"detail,omitempty"`
	Identifiers types.DataMap   `json:"identifiers,omitempty"`
	Enrolments  types.DataMap   `json:"enrolments,omitempty"`
}

type contentExtractor interface {
	populateAuditEvent(ev *AuditEvent)
	populateDetail(ev *RATEAuditDetail)
	populateIdentifiers(ev *RATEAuditEvent)
	populateEnrolments(ev *RATEAuditEvent)
	populateMessageSpecificInfo(ev *RATEAuditEvent)
}

// AppendRequest appends information about the request to the audit event
func (ev *RATEAuditEvent) AppendRequest(req *http.Request) {
	appendCommonRequestFields(&ev.AuditEvent, req)
	appendMessageContent(ev, req)
}

func appendMessageContent(ev *RATEAuditEvent, req *http.Request) {

	body, err := copyBody(req)
	url, err := url.ParseRequestURI(req.RequestURI)

	if err != nil {
		log.Errorf("Error reading request body: %v", err)
	}

	decoder := xml.NewDecoder(body)
	if root, err := xmlutils.ScrollToNextElement(decoder); err == nil {
		var extractor contentExtractor
		switch root.Name.Local {
		case "GovTalkMessage":
			extractor, err = gtmGetMessageParts(decoder, url.Path)
		case "ChRISEnvelope":
			extractor, err = ceGetMessageParts(decoder)
		default:
			err = fmt.Errorf("Unhandled XML content %s", root.Name.Local)
		}

		if err == nil {
			extractor.populateAuditEvent(&ev.AuditEvent)
			extractor.populateDetail(&ev.Detail)
			extractor.populateIdentifiers(ev)
			extractor.populateEnrolments(ev)
			extractor.populateMessageSpecificInfo(ev)
		} else {
			log.Debugf("Error processing RATE message: %v", err)
		}
	} else {
		log.Debugf("Error processing RATE message: %v", err)
	}

	if ev.AuditType == "" {
		ev.AuditType = types.UnclassifiedRequest
	}
}

// AppendResponse appends information about the response to the audit event
func (ev *RATEAuditEvent) AppendResponse(responseHeaders http.Header, respInfo types.ResponseInfo) {
	appendCommonResponseFields(&ev.AuditEvent, responseHeaders, respInfo)
}

// ToEncoded transforms the event into an Encoded
func (ev *RATEAuditEvent) ToEncoded() types.Encoded {
	return EncodeToJSON(ev)
}

// NewRATEAuditEvent creates a new APIAuditEvent with the provided auditSource and auditType
func NewRATEAuditEvent() Auditer {
	ev := RATEAuditEvent{}
	ev.AuditEvent = AuditEvent{AuditSource: "transaction-engine-frontend"}
	return &ev
}

// Need to take a copy of the body contents so a fresh reader for body is available to subsequent handlers
func copyBody(req *http.Request) (io.ReadCloser, error) {
	buf, err := ioutil.ReadAll(req.Body)
	if err == nil {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		return ioutil.NopCloser(bytes.NewBuffer(buf)), nil
	}
	return nil, err
}

// GovTalkMessage Processing

type partialGovTalkMessage struct {
	Header  *etree.Document
	Details *etree.Document
	Body    *etree.Document
}

func gtmGetMessageParts(decoder *xml.Decoder, path string) (*partialGovTalkMessage, error) {

	partial := partialGovTalkMessage{}
	isSaSubmission := false

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Header" {
				if doc, err := xmlutils.ElementInnerToDocument(&se, decoder); err == nil {
					partial.Header = doc
					if el := doc.FindElementPath(gtmClass); el != nil {
						isSaSubmission = strings.HasPrefix(el.Text(), "HMRC-SA-") && path == "/submission"
					}
				}
			} else if se.Name.Local == "GovTalkDetails" {
				if doc, err := xmlutils.ElementInnerToDocument(&se, decoder); err == nil {
					partial.Details = doc
				}
			} else if isSaSubmission && se.Name.Local == "Body" {
				if doc, err := xmlutils.ElementInnerToDocument(&se, decoder); err == nil {
					partial.Body = doc
				}
			}
		}

		// For non SA events halt parsing when required data is obtained otherwise need to continue parsing to the end
		if !isSaSubmission && partial.Header != nil && partial.Details != nil {
			break // For non SA events halt parsing when required data is obtained
		}
	}

	if partial.Header != nil && partial.Details != nil {
		return &partial, nil
	}
	return nil, errors.New("Unexpected message structure. Headers/GovTalkDetails not present")
}

// Headers XPaths
var gtmClass = etree.MustCompilePath("./Header/MessageDetails/Class")
var gtmPathFunction = etree.MustCompilePath("./Header/MessageDetails/Function")
var gtmPathQualifier = etree.MustCompilePath("./Header/MessageDetails/Qualifier")
var gtmCorrelationID = etree.MustCompilePath("./Header/MessageDetails/CorrelationID")
var gtmTransactionID = etree.MustCompilePath("./Header/MessageDetails/TransactionID")
var gtmEmailAddress = etree.MustCompilePath("./Header/SenderDetails/EmailAddress")
var gtmSenderID = etree.MustCompilePath("./Header/SenderDetails/IDAuthentication/SenderID")

// GovTalkDetails XPaths
var gtmAgentGroupCode = etree.MustCompilePath("./GovTalkDetails/GatewayAdditions/Submitter/AgentDetails/AgentGroupCode")
var gtmAgentEnrolments = etree.MustCompilePath("./GovTalkDetails/GatewayAdditions/Submitter/AgentDetails/AgentEnrolments/Enrolment/ServiceName")
var gtmAgentEnrolmentIds = etree.MustCompilePath("./Identifiers/Identifier")
var gtmCredentialID = etree.MustCompilePath("./GovTalkDetails/GatewayAdditions/Submitter/SubmitterDetails/CredentialIdentifier")
var gtmSoftwareFamily = etree.MustCompilePath("./GovTalkDetails/ChannelRouting/Channel/Product")
var gtmSoftwareVersion = etree.MustCompilePath("./GovTalkDetails/ChannelRouting/Channel/Version")
var gtmIdentifiers = etree.MustCompilePath("./GovTalkDetails/Keys/Key")
var gtmRole = etree.MustCompilePath("./GovTalkDetails/GatewayAdditions/Submitter/SubmitterDetails/CredentialRole")
var gtmUserType = etree.MustCompilePath("./GovTalkDetails/GatewayAdditions/Submitter/SubmitterDetails/RegistrationCategory")

// SA Specific Data
var gtmSa110Repayment = etree.MustCompilePath("./Body/IRenvelope/MTR/SA110/SelfAssessment/TotalTaxEtcDue")
var gtmSa900Claim = etree.MustCompilePath("./Body/IRenvelope/SAtrust/TrustEstate/TaxCalculation/ClaimRepaymentForNextYear")
var gtmSa900Repayment = etree.MustCompilePath("./Body/IRenvelope/SAtrust/TrustEstate/TaxCalculation/RepaymentForNextYear")

func (partial *partialGovTalkMessage) populateAuditEvent(ae *AuditEvent) {
	extractIfPresent(partial.Header, gtmClass, &ae.AuditType)
}

func (partial *partialGovTalkMessage) populateDetail(ev *RATEAuditDetail) {
	extractIfPresent(partial.Header, gtmCorrelationID, &ev.CorrelationID)
	extractIfPresent(partial.Header, gtmTransactionID, &ev.TransactionID)
	extractIfPresent(partial.Header, gtmSenderID, &ev.SenderID)
	extractIfPresent(partial.Header, gtmEmailAddress, &ev.Email)
	extractIfPresent(partial.Details, gtmRole, &ev.Role)
	extractIfPresent(partial.Details, gtmSoftwareFamily, &ev.SoftwareFamily)
	extractIfPresent(partial.Details, gtmSoftwareVersion, &ev.SoftwareVersion)
	extractIfPresent(partial.Details, gtmUserType, &ev.UserType)
	partial.populateRequestType(ev)
}

func (partial *partialGovTalkMessage) populateRequestType(ev *RATEAuditDetail) {
	var function string
	var qualifier string
	extractIfPresent(partial.Header, gtmPathFunction, &function)
	extractIfPresent(partial.Header, gtmPathQualifier, &qualifier)
	ev.RequestType = translateRequestType(function, qualifier)
}

func (partial *partialGovTalkMessage) populateIdentifiers(ev *RATEAuditEvent) {

	ev.Identifiers = types.DataMap{}

	if node := partial.Details.FindElementPath(gtmCredentialID); node != nil {
		ev.Identifiers["credID"] = node.Text()
	}

	if node := partial.Details.FindElementPath(gtmAgentGroupCode); node != nil {
		ev.Identifiers["agentGroupCode"] = node.Text()
	}

	if ids := partial.Details.FindElementsPath(gtmIdentifiers); len(ids) > 0 {
		for _, el := range ids {
			ev.Identifiers[el.SelectAttr("Type").Value] = el.Text()
		}
	}

}

func (partial *partialGovTalkMessage) populateEnrolments(ev *RATEAuditEvent) {
	if nodes := partial.Details.FindElementsPath(gtmAgentEnrolments); len(nodes) > 0 {
		ev.Enrolments = types.DataMap{}
		for _, service := range nodes {
			enrolmentIds := types.DataMap{}
			for _, ids := range service.Parent().FindElementsPath(gtmAgentEnrolmentIds) {
				enrolmentIds[ids.SelectAttr("IdentifierType").Value] = ids.Text()
			}
			ev.Enrolments[service.Text()] = enrolmentIds
		}
	}
}

func (partial *partialGovTalkMessage) populateMessageSpecificInfo(ev *RATEAuditEvent) {
	partial.populateSelfAssessmentData(ev)
}

func (partial *partialGovTalkMessage) populateSelfAssessmentData(ev *RATEAuditEvent) {
	if strings.HasPrefix(ev.AuditType, "HMRC-SA-") && partial.Body != nil {
		payload := xmlutils.XMLToDataMap(partial.Body.ChildElements(), []string{"AttachedFiles", "Attachment"})
		body := payload.GetDataMap("Body")
		if ev.RequestPayload != nil {
			ev.RequestPayload = types.DataMap{}
		}
		ev.RequestPayload["details"] = (body)
		if strings.HasPrefix(ev.AuditType, "HMRC-SA-SA100") {
			if el := partial.Body.FindElementPath(gtmSa110Repayment); el != nil {
				if amount, err := strconv.ParseFloat(el.Text(), 64); err == nil {
					ev.Detail.IsRepayment = strconv.FormatBool(amount < 0.00)
				}

			}
		}
		if strings.HasPrefix(ev.AuditType, "HMRC-SA-SA900") {
			claim := partial.Body.FindElementPath(gtmSa900Claim)
			amount := partial.Body.FindElementPath(gtmSa900Repayment)

			if claim != nil && amount != nil {
				if amount, err := strconv.ParseFloat(amount.Text(), 64); err == nil {
					ev.Detail.IsRepayment = strconv.FormatBool(strings.ToLower(strings.TrimSpace(claim.Text())) == "yes" && amount > 0.00)
				}

			}
		}
	}
}

// ChRISEnvelope Processing
type partialChrisEnvelope struct {
	Header     *etree.Document
	IrEnvelope *etree.Document
}

func ceGetMessageParts(decoder *xml.Decoder) (*partialChrisEnvelope, error) {
	var header xmlutils.XMLFragment
	var irEnv xmlutils.XMLFragment
	haveHeader := false
	haveIrEnv := false

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
			} else if se.Name.Local == "IRenvelope" {
				if err := decoder.DecodeElement(&irEnv, &se); err == nil {
					haveIrEnv = true
				}
			}
		}

		// Stop parsing when required data is obtained
		if haveHeader && haveIrEnv {
			partial := partialChrisEnvelope{}
			headerDoc := etree.NewDocument()
			if err := headerDoc.ReadFromBytes(header.InnerXML); err != nil {
				return nil, err
			}
			envDoc := etree.NewDocument()
			if err := envDoc.ReadFromBytes(irEnv.InnerXML); err != nil {
				return nil, err
			}
			partial.Header = headerDoc
			partial.IrEnvelope = envDoc
			return &partial, nil
		}
	}

	return nil, errors.New("Unexpected message structure. Headers/IREnvelope not present")
}

// Header Paths
var ceClass = etree.MustCompilePath("./MessageClass")
var cePathFunction = etree.MustCompilePath("./Function")
var cePathQualifier = etree.MustCompilePath("./Qualifier")

var ceCorrelationID = etree.MustCompilePath("./Sender/CorrelatingID")
var ceSenderID = etree.MustCompilePath("./Sender/Additions/EDI/TradingPartnerID")

// IRenvelope Paths
var ceIdentifiers = etree.MustCompilePath("./IRheader/Keys/Key")

func (partial *partialChrisEnvelope) populateAuditEvent(ev *AuditEvent) {
	extractIfPresent(partial.Header, ceClass, &ev.AuditType)
}

func (partial *partialChrisEnvelope) populateDetail(ev *RATEAuditDetail) {
	extractIfPresent(partial.Header, ceCorrelationID, &ev.CorrelationID)
	extractIfPresent(partial.Header, ceSenderID, &ev.SenderID)
	partial.populateRequestType(ev)
}

func (partial *partialChrisEnvelope) populateRequestType(ev *RATEAuditDetail) {
	var function string
	var qualifier string
	extractIfPresent(partial.Header, cePathFunction, &function)
	extractIfPresent(partial.Header, cePathQualifier, &qualifier)
	ev.RequestType = translateRequestType(function, qualifier)
}

func (partial *partialChrisEnvelope) populateIdentifiers(ev *RATEAuditEvent) {
	ev.Identifiers = types.DataMap{}
	if ids := partial.IrEnvelope.FindElementsPath(ceIdentifiers); len(ids) > 0 {
		for _, el := range ids {
			ev.Identifiers[el.SelectAttr("Type").Value] = el.Text()
		}
	}
}

func (partial *partialChrisEnvelope) populateEnrolments(ev *RATEAuditEvent) {
	ev.Enrolments = types.DataMap{}
}

func (partial *partialChrisEnvelope) populateMessageSpecificInfo(ev *RATEAuditEvent) {
	// Noop
}

func extractIfPresent(doc *etree.Document, path etree.Path, dest *string) {
	if node := doc.FindElementPath(path); node != nil {
		*dest = node.Text()
	}
}

func translateRequestType(function string, qualifier string) string {
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
