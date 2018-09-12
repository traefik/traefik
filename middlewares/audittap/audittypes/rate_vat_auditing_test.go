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
	spec := &AuditSpecification{}
	event.AppendRequest(NewRequestContext(req), spec)
	event.AppendResponse(respHdrs, respInfo, spec)

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
	spec := &AuditSpecification{}
	event.AppendRequest(NewRequestContext(req), spec)
	event.AppendResponse(respHdrs, respInfo, spec)

	assert.Equal(t, "HMRC-VAT-DEC-TMSG", event.AuditType)
	assert.NotEmpty(t, event.RequestPayload)
	vatData := event.RequestPayload.GetString("contents")
	assert.True(t, strings.HasPrefix(vatData, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><GovTalkMessage"))
	assert.Contains(t, vatData, "<VATDueOnECAcquisitions>13.12</VATDueOnECAcquisitions>")
	assert.Contains(t, vatData, "<Key Type=\"VATRegNo\">999947314</Key>")
}
