package audittypes

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestRateVATInfoIgnoresNonSubmission(t *testing.T) {

	types.TheClock = T0

	vatDecl, err := ioutil.ReadFile("testdata/HMRC-VAT-DEC-TMSG.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/notasubmission?qq=zz", bytes.NewReader([]byte(vatDecl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	obfuscate := AuditObfuscation{}
	event.AppendRequest(req, obfuscate)
	event.AppendResponse(respHdrs, respInfo, obfuscate)

	assert.Equal(t, "HMRC-VAT-DEC-TMSG", event.AuditType)
	assert.Equal(t, types.DataMap{}, event.RequestPayload.GetDataMap("contents"))
}

func TestRateVATAuditEvent(t *testing.T) {

	types.TheClock = T0

	vatDecl, err := ioutil.ReadFile("testdata/HMRC-VAT-DEC-TMSG.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewReader([]byte(vatDecl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	obfuscate := AuditObfuscation{}
	event.AppendRequest(req, obfuscate)
	event.AppendResponse(respHdrs, respInfo, obfuscate)

	assert.Equal(t, "HMRC-VAT-DEC-TMSG", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	vatData := event.RequestPayload.GetString("contents")
	assert.True(t, strings.HasPrefix(vatData, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><GovTalkMessage"))
	assert.Contains(t, vatData, "<VATDueOnECAcquisitions>13.12</VATDueOnECAcquisitions>")
	assert.Contains(t, vatData, "<Key Type=\"VATRegNo\">999947314</Key>")
	assert.Equal(t, "true", event.Detail.IsRepayment)
}

func TestRateVATAuditEventIsNotRepayment(t *testing.T) {

	types.TheClock = T0
	x := `	
	<GovTalkMessage>
	<Body>
	<IRenvelope>
		<VATDeclarationRequest>
			<TotalVAT>0.02</TotalVAT>
			<VATReclaimedOnInputs>0.01</VATReclaimedOnInputs>
		</VATDeclarationRequest>
	</IRenvelope>
	</Body>	
	</GovTalkMessage>
	`
	gtm, err := makePartialGtmWithBody(x)
	if err != nil {
		t.Fatal(err)
	}
	event := &RATEAuditEvent{}
	event.AuditType = "HMRC-VAT-TMSG"
	gtm.populateSelfAssessmentData(event)

	assert.Equal(t, "false", event.Detail.IsRepayment)
}

func TestRateVATAuditEventIsRepaymentWhenEmpty(t *testing.T) {

	types.TheClock = T0
	x := `	
	<GovTalkMessage>
	<Body>
	<IRenvelope>
		<VATDeclarationRequest>
			<NetVAT />
		</VATDeclarationRequest>
	</IRenvelope>
	</Body>	
	</GovTalkMessage>
	`
	gtm, err := makePartialGtmWithBody(x)
	if err != nil {
		t.Fatal(err)
	}
	event := &RATEAuditEvent{}
	event.AuditType = "HMRC-VAT-TMSG"
	gtm.populateSelfAssessmentData(event)

	assert.Equal(t, "false", event.Detail.IsRepayment)
}

func TestRateVATAuditEventIsRepaymentOmitted(t *testing.T) {

	types.TheClock = T0
	x := `	
	<GovTalkMessage>
	<Body>
	<IRenvelope>
		<VATDeclarationRequest>
			<NoVatInfoHere>foo</NoVatInfoHere>
		</VATDeclarationRequest>
	</IRenvelope>
	</Body>	
	</GovTalkMessage>
	`
	gtm, err := makePartialGtmWithBody(x)
	if err != nil {
		t.Fatal(err)
	}
	event := &RATEAuditEvent{}
	event.AuditType = "HMRC-VAT-TMSG"
	gtm.populateSelfAssessmentData(event)

	assert.Equal(t, "false", event.Detail.IsRepayment)
}
