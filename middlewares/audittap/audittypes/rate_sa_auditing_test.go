package audittypes

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beevik/etree"

	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestRateSaInfoIgnoresNonSubmission(t *testing.T) {

	types.TheClock = T0

	sa100Decl, err := ioutil.ReadFile("testdata/HMRC-SA-SA100-ATT.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/notasubmission?qq=zz", bytes.NewReader([]byte(sa100Decl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "HMRC-SA-SA100-ATT", event.AuditType)
	assert.Equal(t, types.DataMap{}, event.RequestPayload.GetDataMap("contents"))
}

func TestRateSA100AuditEvent(t *testing.T) {

	types.TheClock = T0

	sa100Decl, err := ioutil.ReadFile("testdata/HMRC-SA-SA100-ATT.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewReader([]byte(sa100Decl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "HMRC-SA-SA100-ATT", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	saData := event.RequestPayload.GetDataMap("contents").GetDataMap("IRenvelope").GetDataMap("MTR").GetDataMap("SA100")
	assert.NotEmpty(t, saData)
	assert.Equal(t, "GY001093A", saData.GetDataMap("YourPersonalDetails").GetString("NationalInsuranceNumber"))
	assert.Nil(t, saData.Get("AttachedFiles"))
	assert.Equal(t, "false", event.Detail.IsRepayment)
}

func TestRateSA100AuditEventIsRepayment(t *testing.T) {

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
	assert.Equal(t, "true", event.Detail.IsRepayment)
}

func TestRateSA100AuditEventIsRepaymentWhenEmpty(t *testing.T) {

	types.TheClock = T0
	x := `	
	<Body>
	<IRenvelope xmlns="http://www.govtalk.gov.uk/taxation/SA/SA100/15-16/1">
		<MTR>
			<SA110>
				<SelfAssessment>
					<TotalTaxEtcDue />
				</SelfAssessment>
				<UnderpaidTax>
					<UnderpaidTaxForEarlierYearsIncludedInCode>0.00</UnderpaidTaxForEarlierYearsIncludedInCode>
					<UnderpaidTaxForYearIncludedInFutureCode>0.00</UnderpaidTaxForYearIncludedInFutureCode>
				</UnderpaidTax>
			</SA110>
		</MTR>
	</IRenvelope>
	</Body>	
	`
	gtm, err := makePartialGtmWithBody(x)
	if err != nil {
		t.Fatal(err)
	}
	event := &RATEAuditEvent{}
	event.AuditType = "HMRC-SA-SA100"
	gtm.populateSelfAssessmentData(event)

	assert.Equal(t, "false", event.Detail.IsRepayment)
}

func TestRateSA100AuditEventIsRepaymentOmitted(t *testing.T) {

	types.TheClock = T0
	x := `	
	<Body>
	<IRenvelope xmlns="http://www.govtalk.gov.uk/taxation/SA/SA100/15-16/1">
		<MTR>
			<SA110>
				<SelfAssessment />
				<UnderpaidTax>
					<UnderpaidTaxForEarlierYearsIncludedInCode>0.00</UnderpaidTaxForEarlierYearsIncludedInCode>
					<UnderpaidTaxForYearIncludedInFutureCode>0.00</UnderpaidTaxForYearIncludedInFutureCode>
				</UnderpaidTax>
			</SA110>
		</MTR>
	</IRenvelope>
	</Body>	
	`
	gtm, err := makePartialGtmWithBody(x)
	if err != nil {
		t.Fatal(err)
	}
	event := &RATEAuditEvent{}
	event.AuditType = "HMRC-SA-SA100"
	gtm.populateSelfAssessmentData(event)

	assert.Equal(t, "", event.Detail.IsRepayment)
}

func TestRateSA800AuditEvent(t *testing.T) {

	types.TheClock = T0

	sa800Decl, err := ioutil.ReadFile("testdata/HMRC-SA-SA800-ATT-TIL.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewReader([]byte(sa800Decl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "HMRC-SA-SA800-ATT-TIL", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	saData := event.RequestPayload.GetDataMap("contents").GetDataMap("IRenvelope").GetDataMap("SApartnership")
	assert.NotEmpty(t, saData)
	assert.Equal(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456", saData.GetString("PartnershipName"))
	assert.Nil(t, saData.Get("AttachedFiles"))
	assert.Equal(t, "", event.Detail.IsRepayment)
}

func TestRateSA900AuditEvent(t *testing.T) {

	types.TheClock = T0

	sa900Decl, err := ioutil.ReadFile("testdata/HMRC-SA-SA900-ATT-TIL.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewReader([]byte(sa900Decl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "HMRC-SA-SA900-ATT-TIL", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	saData := event.RequestPayload.GetDataMap("contents").GetDataMap("IRenvelope").GetDataMap("SAtrust")
	assert.NotEmpty(t, saData)
	assert.Equal(t, "Cap Trust", saData.GetString("TrustName"))
	assert.Equal(t, "yes", saData.GetDataMap("TrustEstate").GetDataMap("NotLiableAtTrustRate").GetString("NotLiable"))
	assert.Nil(t, saData.Get("AttachedFiles"))
	assert.Equal(t, "false", event.Detail.IsRepayment)
}

func TestRateSA900AuditEventIsRepayment(t *testing.T) {

	types.TheClock = T0

	sa900Decl, err := ioutil.ReadFile("testdata/HMRC-SA-SA900.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewReader([]byte(sa900Decl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	event.AppendRequest(req)
	event.AppendResponse(respHdrs, respInfo)

	assert.Equal(t, "HMRC-SA-SA900", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	assert.Equal(t, "true", event.Detail.IsRepayment)

}

func TestRateSA900AuditEventIsRepaymentWhenEmpty(t *testing.T) {

	types.TheClock = T0
	x := `	
	<Body>
	<IRenvelope>
		<SAtrust>
			<TrustEstate>
				<TaxCalculation>
					<ClaimRepaymentForNextYear />
					<RepaymentForNextYear />
				</TaxCalculation>
			</TrustEstate>
		</SAtrust>
	</IRenvelope>
	</Body>	
	`
	gtm, err := makePartialGtmWithBody(x)
	if err != nil {
		t.Fatal(err)
	}
	event := &RATEAuditEvent{}
	event.AuditType = "HMRC-SA-SA900"
	gtm.populateSelfAssessmentData(event)

	assert.Equal(t, "false", event.Detail.IsRepayment)
}

func TestRateSA900AuditEventIsRepaymentOmitted(t *testing.T) {

	types.TheClock = T0
	x := `	
	<Body>
	<IRenvelope>
		<SAtrust>
			<TrustEstate>
				<TaxCalculation>
					<SomeOtherData>ABC</SomeOtherData>
				</TaxCalculation>
			</TrustEstate>
		</SAtrust>
	</IRenvelope>
	</Body>	
	`
	gtm, err := makePartialGtmWithBody(x)
	if err != nil {
		t.Fatal(err)
	}
	event := &RATEAuditEvent{}
	event.AuditType = "HMRC-SA-SA900"
	gtm.populateSelfAssessmentData(event)

	assert.Equal(t, "", event.Detail.IsRepayment)
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
	gtm.Body = doc
	return gtm, nil
}
