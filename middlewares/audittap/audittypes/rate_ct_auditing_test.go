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

func TestRateCTInfoIgnoresNonSubmission(t *testing.T) {

	types.TheClock = T0

	ctDecl, err := ioutil.ReadFile("testdata/HMRC-CT-CT600.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/notasubmission?qq=zz", bytes.NewReader([]byte(ctDecl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	spec := &AuditSpecification{}
	event.AppendRequest(NewRequestContext(req), spec)
	event.AppendResponse(respHdrs, respInfo, spec)

	assert.Equal(t, "HMRC-CT-CT600", event.AuditType)
	assert.Equal(t, types.DataMap{}, event.RequestPayload.GetDataMap(keyPayloadContents))
}

func TestRateCTAuditEvent(t *testing.T) {

	types.TheClock = T0

	ctDecl, err := ioutil.ReadFile("testdata/HMRC-CT-CT600.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewReader([]byte(ctDecl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{Entity: ([]byte("<SomeCtRespPayload />"))}

	event := &RATEAuditEvent{}
	spec := &AuditSpecification{}
	event.AppendRequest(NewRequestContext(req), spec)
	event.AppendResponse(respHdrs, respInfo, spec)

	assert.Equal(t, "HMRC-CT-CT600", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	ctData := event.RequestPayload.GetString(keyPayloadContents)
	assert.True(t, strings.HasPrefix(ctData, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><GovTalkMessage"))
	assert.Contains(t, ctData, "<Key Type=\"UTR\">8596148860</Key>")
	assert.Contains(t, ctData, "<TaxPayable>20302.69</TaxPayable>")
	assert.Equal(t, "<SomeCtRespPayload />", event.ResponsePayload.GetString(keyPayloadContents))
}

func TestRateCTRemovesAttachmentContent(t *testing.T) {

	types.TheClock = T0
	x := `
	<Body>
	<IRenvelope>
		<AttachedFiles>
			<Attachment att="1" size="999">wdokawdoakwdokw</Attachment>
			<Attachment att="2" size="123">wdefafiejfiajefd</Attachment>
		</AttachedFiles>
		<AttachedFiles>
		<Attachments>
			<Attachment att="3" size="888">wdwadaevaefaefaewf</Attachment>
			<Attachment att="4" size="101001">4trgrgsefsedawwadawd</Attachment>
		</Attachments>
	</IRenvelope>
	</Body>
	`
	gtm, err := makePartialGtmWithBody(x)
	if err != nil {
		t.Fatal(err)
	}
	event := &RATEAuditEvent{}
	event.AuditType = "HMRC-CT-CT600"
	gtm.populateDetails(event)
	contents := event.RequestPayload.GetString(keyPayloadContents)
	assert.Contains(t, contents, "AttachedFiles")
	assert.Contains(t, contents, "<Attachment att=\"1\" size=\"999\"></Attachment>")
	assert.Contains(t, contents, "<Attachment att=\"2\" size=\"123\"></Attachment>")
	assert.NotContains(t, contents, "wdokawdoakwdokw")
	assert.NotContains(t, contents, "4trgrgsefsedawwadawd")
}
