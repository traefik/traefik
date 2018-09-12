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
	spec := &AuditSpecification{}
	event.AppendRequest(NewRequestContext(req), spec)
	event.AppendResponse(respHdrs, respInfo, spec)

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
	spec := &AuditSpecification{}
	event.AppendRequest(NewRequestContext(req), spec)
	event.AppendResponse(respHdrs, respInfo, spec)

	assert.Equal(t, "HMRC-SA-SA100-ATT", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	saData := event.RequestPayload.GetString("contents")
	assert.True(t, strings.HasPrefix(saData, "<?xml version=\"1.0\"?><GovTalkMessage"))
	assert.Contains(t, saData, "<NationalInsuranceNumber>GY001093A")
	assert.Contains(t, saData, "AttachedFiles")
	assert.Contains(t, saData, "<Attachment FileFormat=\"pdf\" Filename=\"tubemap.pdf\" Description=\"TubeMap\" Size=\"315001\"></Attachment>")
}

func TestRateSA800AuditEvent(t *testing.T) {

	types.TheClock = T0

	sa800Decl, err := ioutil.ReadFile("testdata/HMRC-SA-SA800-ATT-TMSG.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewBuffer([]byte(sa800Decl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	spec := &AuditSpecification{}
	event.AppendRequest(NewRequestContext(req), spec)
	event.AppendResponse(respHdrs, respInfo, spec)

	assert.Equal(t, "HMRC-SA-SA800-ATT-TMSG", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	saData := event.RequestPayload.GetString("contents")
	assert.True(t, strings.HasPrefix(saData, "<GovTalkMessage"))
	assert.Contains(t, saData, "PartnershipName>ABCDEFGHIJKLMNOPQRSTUVWXYZ123456")
	assert.Contains(t, saData, "<Attachment FileFormat=\"pdf\" Filename=\"POSATT035small1.pdf\" Size=\"12345\" Description=\"small attachment 1\"></Attachment>")
	assert.Contains(t, saData, "<Attachment FileFormat=\"pdf\" Filename=\"POSATT035small2.pdf\" Size=\"100\" Description=\"small attachment 2\"></Attachment>")
}

func TestRateSA900AuditEvent(t *testing.T) {

	types.TheClock = T0

	sa900Decl, err := ioutil.ReadFile("testdata/HMRC-SA-SA900-ATT-TMSG.xml")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/submission?qq=zz", bytes.NewBuffer([]byte(sa900Decl)))
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{}

	event := &RATEAuditEvent{}
	spec := &AuditSpecification{}
	event.AppendRequest(NewRequestContext(req), spec)
	event.AppendResponse(respHdrs, respInfo, spec)

	assert.Equal(t, "HMRC-SA-SA900-ATT-TMSG", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	saData := event.RequestPayload.GetString("contents")
	assert.True(t, strings.HasPrefix(saData, "<GovTalkMessage"))
	assert.Contains(t, saData, "TrustName>Cap Trust")
	assert.Contains(t, saData, "Attachment")
}

func TestRateSARemovesAttachmentContent(t *testing.T) {

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
	event.AuditType = "HMRC-SA-SA900"
	gtm.populateDetails(event)
	contents := event.RequestPayload.GetString("contents")
	assert.Contains(t, contents, "AttachedFiles")
	assert.Contains(t, contents, "<Attachment att=\"1\" size=\"999\"></Attachment>")
	assert.Contains(t, contents, "<Attachment att=\"2\" size=\"123\"></Attachment>")
	assert.NotContains(t, contents, "wdokawdoakwdokw")
	assert.NotContains(t, contents, "4trgrgsefsedawwadawd")
}
