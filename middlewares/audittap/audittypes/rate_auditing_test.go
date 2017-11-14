package audittypes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/xml"
	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestRateAuditEvent(t *testing.T) {

	types.TheClock = T0

	vatDecl := `
<?xml version="1.0" encoding="UTF-8"?>
<GovTalkMessage xmlns="http://www.govtalk.gov.uk/CM/envelope">
	<EnvelopeVersion>2.0</EnvelopeVersion>
	<Header>
		<MessageDetails>
			<Class>HMRC-VAT-DEC-TIL</Class>
			<Qualifier>request</Qualifier>
			<Function>submit</Function>
			<TransactionID>1A002180711</TransactionID>
			<AuditID>A1002180711</AuditID>
			<CorrelationID>998C7D7DF2134835A332FAB2EB1914F3</CorrelationID>
			<ResponseEndPoint>HMCE-SERVICE-ENDPOINT-EXTRA-1</ResponseEndPoint>
			<GatewayTest>0</GatewayTest>
			<GatewayTimestamp>2015-11-18T08:53:23.970</GatewayTimestamp>
		</MessageDetails>
		<SenderDetails>
			<IDAuthentication>
				<SenderID>0000001663017753</SenderID>
				<Authentication>
					<Method>clear</Method>
					<Role>principal</Role>
					<Value>**********</Value>
				</Authentication>
			</IDAuthentication>
			<X509Certificate/>
			<EmailAddress>aa@aa.com</EmailAddress>
		</SenderDetails>
	</Header>
	<GovTalkDetails>
		<Keys>
			<Key Type='SA'>SA554433</Key>
			<Key Type='VATRegNo'>VAT443322</Key>
		</Keys>
		<TargetDetails>
			<Organisation>DecisionSoft Ltd.</Organisation>
		</TargetDetails>
		<GatewayValidation>
			<Processed>yes</Processed>
			<Result>pass</Result>
		</GatewayValidation>
		<ChannelRouting>
			<Channel>
				<URI>http://www.decisionsoft.com/9876</URI>
				<Product>X-Meta</Product>
				<Version>2.02</Version>
			</Channel>
			<ID Type='main'>Channel 1</ID>
			<Timestamp>2009-03-17T02:58:41</Timestamp>
		</ChannelRouting>
		<GatewayAdditions>
			<Submitter xmlns="http://www.govtalk.gov.uk/gateway/submitterdetails">
				<AgentDetails>
					<AuthenticationType>1</AuthenticationType>
					<AgentGroupCode>AGT334455</AgentGroupCode>
					<AgentEnrolments>
						<Enrolment>
							<ServiceName>SERV_AGT1</ServiceName>
							<Identifiers>
								<Identifier IdentifierType="AGT1_ID1">XXYY1111</Identifier>
								<Identifier IdentifierType="AGT1_ID2">XXYY2222</Identifier>
							</Identifiers>
						</Enrolment>
						<Enrolment>
							<ServiceName>SERV_AGT2</ServiceName>
							<Identifiers>
								<Identifier IdentifierType="AGT2_ID1">TTYY1111</Identifier>
								<Identifier IdentifierType="AGT2_ID2">TTYY2222</Identifier>
							</Identifiers>
						</Enrolment>
					</AgentEnrolments>
				</AgentDetails>
				<SubmitterDetails>
					<RegistrationCategory>Individual</RegistrationCategory>
					<UserType>Principal</UserType>
					<CredentialRole>User</CredentialRole>
					<CredentialIdentifier>0000001663017753</CredentialIdentifier>
				</SubmitterDetails>
			</Submitter>
		</GatewayAdditions>
	</GovTalkDetails>
	<Body>
		<IRenvelope xmlns="http://www.govtalk.gov.uk/taxation/vat/vatdeclaration/2">
			<IRheader>
				<Keys>
					<Key Type='VATRegNo'>999947314</Key>
				</Keys>
				<PeriodID>2009-12</PeriodID>
				<PeriodStart>2009-12-01</PeriodStart>
				<PeriodEnd>2009-12-31</PeriodEnd>
				<Principal>
					<Contact>
						<Name>
							<Ttl>Dr</Ttl>
							<Fore>James</Fore>
							<Sur>Bacon</Sur>
						</Name>
						<Email Preferred='yes' Type='work'>sample@EmailStructure.com
						</Email>
						<Telephone Mobile='yes' Preferred='yes' Type='work'>
							<Number>01865 203192</Number>
							<Extension>6969</Extension>
						</Telephone>
						<Fax Mobile='yes' Preferred='yes' Type='work'>
							<Number>01865 203192</Number>
							<Extension>6969</Extension>
						</Fax>
					</Contact>
				</Principal>
				<Agent>
					<AgentID>SM1</AgentID>
					<Company>DecisionSoft Ltd.</Company>
					<Address>
						<Line>Rectory Stable Cottage</Line>
						<PostCode>AA1 1AA</PostCode>
						<Country>UK</Country>
					</Address>
					<Contact>
						<Name>
							<Ttl>Dr</Ttl>
							<Fore>James</Fore>
							<Sur>Bacon</Sur>
						</Name>
						<Email Preferred='yes' Type='work'>sample@EmailStructure.com
						</Email>
						<Telephone Mobile='yes' Preferred='yes' Type='work'>
							<Number>01865 203192</Number>
							<Extension>6969</Extension>
						</Telephone>
						<Fax Mobile='yes' Preferred='yes' Type='work'>
							<Number>01865 203192</Number>
							<Extension>6969</Extension>
						</Fax>
					</Contact>
				</Agent>
				<DefaultCurrency>GBP</DefaultCurrency>
				<Manifest>
					<Contains>
						<Reference>
							<Namespace>http://www.govtalk.gov.uk/taxation/vat/vatdeclaration/2
							</Namespace>
							<SchemaVersion>2009-v1.00</SchemaVersion>
							<TopElementName>A</TopElementName>
						</Reference>
					</Contains>
				</Manifest>
				<IRmark Type='generic'>ju9VSFvIRsMYd7RAYWJf4jRTMiY=</IRmark>
				<Sender>Individual</Sender>
			</IRheader>
			<VATDeclarationRequest>
				<VATDueOnOutputs>13.12</VATDueOnOutputs>
				<VATDueOnECAcquisitions>13.12</VATDueOnECAcquisitions>
				<TotalVAT>26.24</TotalVAT>
				<VATReclaimedOnInputs>10.01</VATReclaimedOnInputs>
				<NetVAT>16.23</NetVAT>
				<NetSalesAndOutputs>13</NetSalesAndOutputs>
				<NetPurchasesAndInputs>13</NetPurchasesAndInputs>
				<NetECSupplies>0</NetECSupplies>
				<NetECAcquisitions>13</NetECAcquisitions>
				<AASBalancingPayment>0.00</AASBalancingPayment>
			</VATDeclarationRequest>
		</IRenvelope>
	</Body>
</GovTalkMessage>
`
	req := httptest.NewRequest("POST", "/some/rate/url?qq=zz", bytes.NewReader([]byte(vatDecl)))
	req.Header.Set("X-Request-Id", "req321")
	req.Header.Set("True-Client-IP", "101.1.101.1")
	req.Header.Set("True-Client-Port", "5005")
	req.Header.Set("X-Source", "202.2.202.2")
	req.Header.Set("Request-ID", "R123")
	req.Header.Set("Session-ID", "S123")
	req.Header.Set("Akamai-Test-Hdr", "Ak999")

	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "HMRC-VAT-DEC-TIL", event.AuditType)
	assert.Equal(t, "1", event.Version)
	assert.Equal(t, "POST", event.Method)
	assert.Equal(t, "/some/rate/url", event.Path)
	assert.Equal(t, "qq=zz", event.QueryString)
	assert.Equal(t, "req321", event.RequestID)
	assert.Equal(t, "101.1.101.1", event.ClientIP)
	assert.Equal(t, "5005", event.ClientPort)
	assert.Equal(t, "202.2.202.2", event.ReceivingIP)
	assert.Equal(t, "2001-09-09T01:46:40.000Z", event.GeneratedAt)
	assert.Equal(t, types.DataMap{"session-id": "S123", "request-id": "R123"}, event.ClientHeaders)
	assert.Equal(t, types.DataMap{"akamai-test-hdr": "Ak999"}, event.RequestHeaders)

	assert.Equal(t, "998C7D7DF2134835A332FAB2EB1914F3", event.Detail.CorrelationID)
	assert.Equal(t, "1A002180711", event.Detail.TransactionID)
	assert.Equal(t, "0000001663017753", event.Detail.SenderID)
	assert.Equal(t, "aa@aa.com", event.Detail.Email)
	assert.Equal(t, "X-Meta", event.Detail.SoftwareFamily)
	assert.Equal(t, "2.02", event.Detail.SoftwareVersion)
	assert.Equal(t, "SUBMISSION_REQUEST", event.Detail.RequestType)
	assert.Equal(t, "User", event.Detail.Role)
	assert.Equal(t, "Individual", event.Detail.UserType)

	assert.Equal(t, "0000001663017753", event.Identifiers.Get("credID"))
	assert.Equal(t, "AGT334455", event.Identifiers.Get("agentGroupCode"))
	assert.Equal(t, "SA554433", event.Identifiers.Get("SA"))
	assert.Equal(t, "VAT443322", event.Identifiers.Get("VATRegNo"))

	assert.Equal(t, types.DataMap{"AGT1_ID1": "XXYY1111", "AGT1_ID2": "XXYY2222"}, event.Enrolments.Get("SERV_AGT1"))
	assert.Equal(t, types.DataMap{"AGT2_ID1": "TTYY1111", "AGT2_ID2": "TTYY2222"}, event.Enrolments.Get("SERV_AGT2"))

}

func TestChrisRateAuditEvent(t *testing.T) {

	types.TheClock = T0

	chrisEnvelope := `
<?xml version="1.0" encoding="UTF-8"?>
<ChRISEnvelope xmlns="http://www.hmrc.gov.uk/ChRIS/Envelope/2" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	<EnvelopeVersion>2.0</EnvelopeVersion>
	<Header>
		<MessageClass>HMRC-PAYE-RTI-EPS</MessageClass>
		<Qualifier>request</Qualifier>
		<Function>submit</Function>
		<Sender>
			<System>EDI-TP</System>
			<CorrelatingID>AAAAZZZZCORRID</CorrelatingID>
			<ReceiptDate>2016-03-11T08:51:22</ReceiptDate>
			<Additions>
				<EDI xmlns="http://www.hmrc.gov.uk/ChRIS/Envelope/Additions/EDI/1" xmlns:EDIAdditions="http://www.hmrc.gov.uk/ChRIS/Envelope/Additions/ED/1">
					<Ref>EPS17|Worksoft 98|999|AZ12345678|123PQ7654321X|15935</Ref>
					<TradingPartnerID>bzc0qd</TradingPartnerID>
				</EDI>
			</Additions>
		</Sender>
	</Header>
	<Body>
		<IRenvelope xmlns="http://www.govtalk.gov.uk/taxation/PAYE/RTI/EmployerPaymentSummary/16-17/1">
			<IRheader>
				<Keys>
					<Key Type="TaxOfficeNumber">999</Key>
					<Key Type="TaxOfficeReference">AZ12345678</Key>
					<Key Type="AORef">123PQ7654321X</Key>
				</Keys>
				<PeriodEnd>2017-04-05</PeriodEnd>
				<DefaultCurrency>GBP</DefaultCurrency>
				<IRmark Type="generic"></IRmark>
				<Sender>Company</Sender>
			</IRheader>
			<EmployerPaymentSummary>
				<EmpRefs>
					<OfficeNo>999</OfficeNo>
					<PayeRef>AZ12345678</PayeRef>
					<AORef>123PQ7654321X</AORef>
				</EmpRefs>
				<EmpAllceInd>yes</EmpAllceInd>
				<RecoverableAmountsYTD>
					<TaxMonth>1</TaxMonth>
					<ShPPRecovered>132.40</ShPPRecovered>
					<NICCompensationOnShPP>52.40</NICCompensationOnShPP>
				</RecoverableAmountsYTD>
				<RelatedTaxYear>16-17</RelatedTaxYear>
			</EmployerPaymentSummary>
		</IRenvelope>
	</Body>
</ChRISEnvelope>
	`

	req := httptest.NewRequest("POST", "/some/rate/chris/url?qq=zz", bytes.NewReader([]byte(chrisEnvelope)))
	req.Header.Set("X-Request-Id", "req321")
	req.Header.Set("True-Client-IP", "101.1.101.1")
	req.Header.Set("True-Client-Port", "5005")
	req.Header.Set("X-Source", "202.2.202.2")
	req.Header.Set("Request-ID", "R123")
	req.Header.Set("Session-ID", "S123")
	req.Header.Set("Akamai-Test-Hdr", "Ak999")

	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "HMRC-PAYE-RTI-EPS", event.AuditType)
	assert.Equal(t, "1", event.Version)
	assert.Equal(t, "POST", event.Method)
	assert.Equal(t, "/some/rate/chris/url", event.Path)
	assert.Equal(t, "qq=zz", event.QueryString)
	assert.Equal(t, "req321", event.RequestID)
	assert.Equal(t, "101.1.101.1", event.ClientIP)
	assert.Equal(t, "5005", event.ClientPort)
	assert.Equal(t, "202.2.202.2", event.ReceivingIP)
	assert.Equal(t, "2001-09-09T01:46:40.000Z", event.GeneratedAt)
	assert.Equal(t, types.DataMap{"session-id": "S123", "request-id": "R123"}, event.ClientHeaders)
	assert.Equal(t, types.DataMap{"akamai-test-hdr": "Ak999"}, event.RequestHeaders)

	assert.Equal(t, "AAAAZZZZCORRID", event.Detail.CorrelationID)
	assert.Equal(t, "bzc0qd", event.Detail.SenderID)
	assert.Equal(t, "SUBMISSION_REQUEST", event.Detail.RequestType)
	assert.Equal(t, "", event.Detail.TransactionID)
	assert.Equal(t, "", event.Detail.Email)
	assert.Equal(t, "", event.Detail.SoftwareFamily)
	assert.Equal(t, "", event.Detail.SoftwareVersion)
	assert.Equal(t, "", event.Detail.Role)
	assert.Equal(t, "", event.Detail.UserType)

	assert.Equal(t, "999", event.Identifiers.Get("TaxOfficeNumber"))
	assert.Equal(t, "AZ12345678", event.Identifiers.Get("TaxOfficeReference"))
	assert.Equal(t, "123PQ7654321X", event.Identifiers.Get("AORef"))

	assert.Equal(t, types.DataMap{}, event.Enrolments)
}

func TestWillHandleUnknownXml(t *testing.T) {

	types.TheClock = T0

	x := `
		<Root>
			<Header>
				<Element1 />
			</Header>
			<SomeDetails>
				<Element2 />
			</SomeDetails>
		</Root>
	`

	req := httptest.NewRequest("POST", "/some/rate/url?qq=zz", bytes.NewReader([]byte(x)))
	req.Header.Set("X-Request-Id", "req321")
	req.Header.Set("True-Client-IP", "101.1.101.1")
	req.Header.Set("True-Client-Port", "5005")
	req.Header.Set("X-Source", "202.2.202.2")
	req.Header.Set("Request-ID", "R123")
	req.Header.Set("Session-ID", "S123")
	req.Header.Set("Akamai-Test-Hdr", "Ak999")

	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "UnclassifiedRequest", event.AuditType)
	assert.Equal(t, "1", event.Version)
	assert.Equal(t, "POST", event.Method)
	assert.Equal(t, "/some/rate/url", event.Path)
	assert.Equal(t, "qq=zz", event.QueryString)
	assert.Equal(t, "req321", event.RequestID)
	assert.Equal(t, "101.1.101.1", event.ClientIP)
	assert.Equal(t, "5005", event.ClientPort)
	assert.Equal(t, "202.2.202.2", event.ReceivingIP)
	assert.Equal(t, "2001-09-09T01:46:40.000Z", event.GeneratedAt)
	assert.Equal(t, types.DataMap{"session-id": "S123", "request-id": "R123"}, event.ClientHeaders)
	assert.Equal(t, types.DataMap{"akamai-test-hdr": "Ak999"}, event.RequestHeaders)

	assert.Equal(t, RATEAuditDetail{}, event.Detail)
	assert.Equal(t, types.DataMap(nil), event.Identifiers)
	assert.Equal(t, types.DataMap(nil), event.Enrolments)

}

func TestGetMessageParts(t *testing.T) {
	x := `
		<Root>
			<Header>
				<Element1 />
			</Header>
			<GovTalkDetails>
				<Element2 />
			</GovTalkDetails>
		</Root>
	`

	decoder := xml.NewDecoder(bytes.NewReader([]byte(x)))
	parts, _ := gtmGetMessageParts(decoder)
	assert.NotEmpty(t, parts.Header)
	assert.NotEmpty(t, parts.Details)
}

func TestXmlMissingHeader(t *testing.T) {
	x := `
		<Root>
			<GovTalkDetails>
				<Element1 />
			</GovTalkDetails>
		</Root>
	`

	decoder := xml.NewDecoder(bytes.NewReader([]byte(x)))
	_, err := gtmGetMessageParts(decoder)
	assert.Error(t, err)
}

func TestXmlMissingDetails(t *testing.T) {
	x := `
		<Root>
			<Header>
				<Element1 />
			</Header>
		</Root>
	`

	decoder := xml.NewDecoder(bytes.NewReader([]byte(x)))
	_, err := gtmGetMessageParts(decoder)
	assert.Error(t, err)
}

func TestNewRateAudit(t *testing.T) {
	auditer := NewRATEAuditEvent()
	if rate, ok := auditer.(*RATEAuditEvent); ok {
		rate.AuditSource = "transaction-engine-frontend"
	} else {
		assert.Fail(t, "Was not a RATEAuditEvent")
	}
}
