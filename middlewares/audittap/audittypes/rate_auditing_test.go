package audittypes

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/xml"

	"github.com/beevik/etree"
	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestRateAuditEvent(t *testing.T) {

	types.TheClock = T0
	vatDecl, err := ioutil.ReadFile("testdata/HMRC-VAT-DEC-TMSG.xml")
	if err != nil {
		t.Fatal(err)
	}
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

	assert.Equal(t, "HMRC-VAT-DEC-TMSG", event.AuditType)
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

	payloadLen, _ := event.RequestPayload.Get("length").(int64)
	assert.EqualValues(t, len(vatDecl), payloadLen)

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

	shouldAudit := event.EnforceConstraints(AuditConstraints{MaxAuditLength: 100000, MaxPayloadContentsLength: 100000})
	assert.True(t, shouldAudit)
}

func TestChrisRateAuditEvent(t *testing.T) {

	types.TheClock = T0

	chrisPayeDecl, err := ioutil.ReadFile("testdata/HMRC-PAYE-RTI-EPS.xml")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/some/rate/chris/url?qq=zz", bytes.NewReader([]byte(chrisPayeDecl)))
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

	payloadLen, _ := event.RequestPayload.Get("length").(int64)
	assert.EqualValues(t, len(chrisPayeDecl), payloadLen)

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

	shouldAudit := event.EnforceConstraints(AuditConstraints{MaxAuditLength: 100000, MaxPayloadContentsLength: 100000})
	assert.True(t, shouldAudit)
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
	xbytes := []byte(x)

	req := httptest.NewRequest("POST", "/some/rate/url?qq=zz", bytes.NewReader(xbytes))
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
	payloadLen, _ := event.RequestPayload.Get("length").(int64)
	assert.EqualValues(t, len(xbytes), payloadLen)

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
	parts, _ := gtmGetMessageParts(decoder, "", bytes.NewBuffer([]byte{}))
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
	_, err := gtmGetMessageParts(decoder, "", bytes.NewBuffer([]byte{}))
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
	_, err := gtmGetMessageParts(decoder, "", bytes.NewBuffer([]byte{}))
	assert.Error(t, err)
}

func TestRemovesAuthenticationCredentials(t *testing.T) {
	x := `
	<GovTalkMessage>
		<EnvelopeVersion>2.0</EnvelopeVersion>
		<Header>
			<MessageDetails>
				<Class>HMRC-SA-SA900</Class>
				<Qualifier>request</Qualifier>
				<Function>submit</Function>
				<TransactionID/>
				<CorrelationID>2A8E3E81EEF7490D8506C2EEBB82C882</CorrelationID>
				<ResponseEndPoint>IR-SERVICE-ENDPOINT-EXTRA-4</ResponseEndPoint>
				<Transformation>XML</Transformation>
				<GatewayTest>0</GatewayTest>
				<GatewayTimestamp>2016-04-01T08:04:10.600</GatewayTimestamp>
			</MessageDetails>
			<SenderDetails>
				<IDAuthentication>
					<SenderID>0000002120421621</SenderID>
					<Authentication>
						<Method>clear</Method>
						<Role>Authenticate/Validate</Role>
						<Value>DoNotIncludeMe</Value>
					</Authentication>
				</IDAuthentication>
				<X509Certificate/>
				<EmailAddress>placeholder@gateway.com</EmailAddress>
			</SenderDetails>
		</Header>
		<GovTalkDetails>
			<Keys>
				<Key Type='UTR'>7122173812</Key>
			</Keys>
		</GovTalkDetails>	
	</GovTalkMessage>
	`

	decoder := xml.NewDecoder(bytes.NewReader([]byte(x)))
	parts, _ := gtmGetMessageParts(decoder, "/submission", bytes.NewReader([]byte(x)))
	assert.NotEmpty(t, parts.Message)
	cred := parts.Message.FindElement("./GovTalkMessage/Header/SenderDetails/IDAuthentication/Authentication/Value")
	assert.NotNil(t, cred)
	assert.Equal(t, "***", cred.Text())
}

func TestProcessingSkippedForTestInLive(t *testing.T) {
	types.TheClock = T0

	sa100Decl, err := ioutil.ReadFile("testdata/HMRC-SA-SA100-TIL.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewReader([]byte(sa100Decl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "HMRC-SA-SA100-TIL", event.AuditType)
	assert.Equal(t, types.DataMap{}, event.RequestPayload.GetDataMap("contents"))
	assert.Nil(t, event.Identifiers)
	assert.Nil(t, event.Enrolments)
	assert.Equal(t, "", event.Detail.IsRepayment)
	assert.False(t, event.EnforceConstraints(AuditConstraints{MaxAuditLength: 1000000, MaxPayloadContentsLength: 100000}))

}

func TestNewRateAudit(t *testing.T) {
	auditer := NewRATEAuditEvent()
	if rate, ok := auditer.(*RATEAuditEvent); ok {
		rate.AuditSource = "transaction-engine-frontend"
	} else {
		assert.Fail(t, "Was not a RATEAuditEvent")
	}
}

// debugEvent debug utility function to output event JSON structure
func debugEvent(t *testing.T, ev *RATEAuditEvent) {
	s := string(ev.ToEncoded().Bytes)
	t.Log(s)
	t.Fatal("Stop the test")
}

func makePartialGtmWithBody(s string) (*partialGovTalkMessage, error) {
	doc := etree.NewDocument()
	err := doc.ReadFromString(s)
	if err != nil {
		return nil, err
	}

	gtm := &partialGovTalkMessage{}
	gtm.Message = doc
	return gtm, nil
}
