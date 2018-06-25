package audittypes

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/stretchr/testify/assert"
)

func TestMdtpAuditEvent(t *testing.T) {

	types.TheClock = T0

	ev := &MdtpAuditEvent{}
	reqBody := "say=Hi&to=Dave"
	req := httptest.NewRequest("POST", "http://my-mdtp-app.public.mdtp/some/resource?p1=v1", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "auth456")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "My-Web-Client 0.01")
	req.Header.Set("Referer", "someotherwebsite")
	req.Header.Set("token", "my-to-ken")
	req.Header.Set("True-Client-IP", "55.44.33.22")
	req.Header.Set("True-Client-Port", "11223")
	req.Header.Set("Akamai-Reputation", "rep-54321")
	req.Header.Set("X-Forwarded-For", "77.88.77.99")
	req.Header.Set("X-Request-ID", "mdtp-req-123")
	req.Header.Set("X-Session-ID", "mdtp-session-99")
	req.AddCookie(&http.Cookie{Name: "mdtpdi", Value: "myMdtpDevice"})
	req.AddCookie(&http.Cookie{Name: "mdtpdf", Value: base64.StdEncoding.EncodeToString([]byte("myDeviceHasFingers"))})

	respBody := "Some response message"
	respHdrs := http.Header{}
	respHdrs.Set("Content-Type", "text/plain")
	respHdrs.Set("Location", "nowherespecific")
	respInfo := types.ResponseInfo{200, 101, []byte(respBody), 2048}

	spec := &AuditSpecification{}
	ev.AppendRequest(req, spec)
	ev.AppendResponse(respHdrs, respInfo, spec)

	assert.NotEmpty(t, ev.EventID)
	assert.Equal(t, "my-mdtp-app", ev.AuditSource)
	assert.Equal(t, "RequestReceived", ev.AuditType)
	assert.NotEmpty(t, ev.GeneratedAt)

	assert.Equal(t, "POST", ev.Detail.GetString("method"))
	assert.Equal(t, "/some/resource", ev.Tags.GetString("path"))
	assert.Equal(t, "p1=v1", ev.Detail.GetString("queryString"))
	assert.Equal(t, "auth456", ev.Detail.GetString("Authorization"))
	assert.Equal(t, "Request to /some/resource", ev.Detail.GetString("input"))
	assert.Equal(t, "someotherwebsite", ev.Detail.GetString("referrer"))
	assert.Equal(t, "77.88.77.99", ev.Detail.GetString("ipAddress"))
	assert.Equal(t, "myMdtpDevice", ev.Detail.GetString("deviceID"))
	assert.Equal(t, "myDeviceHasFingers", ev.Detail.GetString("deviceFingerprint"))
	assert.Equal(t, "My-Web-Client 0.01", ev.Detail.GetString("userAgentString"))
	assert.Equal(t, "application/x-www-form-urlencoded", ev.Detail.GetString(requestContentType))
	assert.EqualValues(t, len(reqBody), ev.Detail.Get(requestBodyLen))
	assert.Equal(t, string(reqBody), ev.Detail.GetString(requestBody))
	assert.Equal(t, "text/plain", ev.Detail.GetString(responseContentType))
	assert.EqualValues(t, len(respBody), ev.Detail.Get(responseBodyLen))
	assert.Equal(t, string(respBody), ev.Detail.GetString(responseBody))
	assert.Equal(t, "200", ev.Detail.GetString("statusCode"))
	assert.Equal(t, "nowherespecific", ev.Detail.GetString("Location"))
	assert.Equal(t, "my-to-ken", ev.Detail.GetString("token"))

	assert.Equal(t, "mdtp-req-123", ev.Tags["X-Request-ID"])
	assert.Equal(t, "mdtp-session-99", ev.Tags["X-Session-ID"])
	assert.Equal(t, "/some/resource", ev.Tags["path"])
	assert.Equal(t, "/some/resource?p1=v1", ev.Tags["transactionName"])
	assert.Equal(t, "55.44.33.22", ev.Tags["clientIP"])
	assert.Equal(t, "11223", ev.Tags["clientPort"])
	assert.Equal(t, "rep-54321", ev.Tags["Akamai-Reputation"])
}

func TestMdtpMappedAuditFieldsAppliedJustAtRequest(t *testing.T) {

	ev := &MdtpAuditEvent{}
	reqBody := "say=Hi&to=Dave"
	req := httptest.NewRequest("POST", "http://my-mdtp-app.public.mdtp/some/resource?p1=v1", strings.NewReader(reqBody))
	req.Header.Set("Gateway-Token", "SomeGatewayToken")
	req.Header.Set("Partner-Token", "SomePartnerToken")

	respHdrs := http.Header{}
	respHdrs.Set("Gateway-Token", "IShouldNotBeUsed")
	respHdrs.Set("Partner-Token", "IShouldNotBeUsed")
	respInfo := types.ResponseInfo{200, 101, nil, 2048}

	mappings := HeaderMappings{
		"detail": HeaderMapping{"partnerToken": "Partner-Token"},
		"tags":   HeaderMapping{"GatewayToken": "Gateway-Token"},
	}
	spec := &AuditSpecification{HeaderMappings: mappings}
	ev.AppendRequest(req, spec)
	ev.AppendResponse(respHdrs, respInfo, spec)

	assert.Equal(t, "SomeGatewayToken", ev.Tags["GatewayToken"])
	assert.Equal(t, "SomePartnerToken", ev.Detail["partnerToken"])
}

func TestRequestPayloadObfuscatedForFormContentType(t *testing.T) {

	types.TheClock = T0

	ev := &MdtpAuditEvent{}
	reqBody := "say=Hi&password=ishouldbesecret&authKey=notforyoureyes&to=Dave"
	req := httptest.NewRequest("POST", "http://my-mdtp-app.public.mdtp/some/resource", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respBody := "Some response message"
	respHdrs := http.Header{}
	respInfo := types.ResponseInfo{200, 101, []byte(respBody), 2048}

	obfuscate := AuditObfuscation{MaskFields: []string{"password", "authKey"}, MaskValue: "@@@"}
	spec := &AuditSpecification{AuditObfuscation: obfuscate}
	expectedBody := "say=Hi&password=@@@&authKey=@@@&to=Dave"
	ev.AppendRequest(req, spec)
	ev.AppendResponse(respHdrs, respInfo, spec)

	assert.EqualValues(t, len(reqBody), ev.Detail.Get(requestBodyLen))
	assert.Equal(t, expectedBody, ev.Detail.GetString(requestBody))

	assert.EqualValues(t, len(respBody), ev.Detail.Get(responseBodyLen))
	assert.Equal(t, string(respBody), ev.Detail.GetString(responseBody))
}

func TestPayloadsNotObfuscated(t *testing.T) {

	types.TheClock = T0

	ev := &MdtpAuditEvent{}
	reqBody := "say=Hi&password=ishouldbesecret&authKey=notforyoureyes&to=Dave"
	req := httptest.NewRequest("POST", "http://my-mdtp-app.public.mdtp/some/resource", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "text/plain")

	respBody := "Some response message"
	respHdrs := http.Header{}
	respHdrs.Set("Content-Type", "text/plain")
	respInfo := types.ResponseInfo{200, 101, []byte(respBody), 2048}

	obfuscate := AuditObfuscation{MaskFields: []string{"password", "authKey"}, MaskValue: "@@@"}
	spec := &AuditSpecification{AuditObfuscation: obfuscate}
	ev.AppendRequest(req, spec)
	ev.AppendResponse(respHdrs, respInfo, spec)

	assert.EqualValues(t, len(reqBody), ev.Detail.Get(requestBodyLen))
	assert.Equal(t, reqBody, ev.Detail.GetString(requestBody))
}

func TestHtmlResponseFiltered(t *testing.T) {

	types.TheClock = T0

	ev := &MdtpAuditEvent{}
	req := httptest.NewRequest("GET", "http://my-mdtp-app.public.mdtp/some/resource", nil)

	respBody := "NotAtAllRealHTML"
	respHdrs := http.Header{}
	respHdrs.Set("Content-Type", "text/html")
	respInfo := types.ResponseInfo{200, 101, []byte(respBody), 2048}

	spec := &AuditSpecification{}
	ev.AppendRequest(req, spec)
	ev.AppendResponse(respHdrs, respInfo, spec)

	assert.EqualValues(t, len(respBody), ev.Detail.Get(responseBodyLen))
	assert.Equal(t, "<HTML>...</HTML>", ev.Detail.GetString(responseBody))
}

func TestMdtpAuditEnforceConstraintsRequestTooLarge(t *testing.T) {

	max := 20
	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(max)}

	ev := &MdtpAuditEvent{Detail: types.DataMap{}}
	ev.Detail[requestBodyLen] = max + 1
	ev.Detail[requestBody] = "IWillBreachTheAllowedLimit"
	ev.EnforceConstraints(constraints)

	assert.Nil(t, ev.Detail.Get(requestBody))
}

func TestMdtpAuditEnforceConstraintsResponseTooLarge(t *testing.T) {

	max := 20
	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(max)}

	ev := &MdtpAuditEvent{Detail: types.DataMap{}}
	ev.Detail[responseBodyLen] = max + 1
	ev.Detail[responseBody] = "IWillBreachTheAllowedLimit"

	ev.EnforceConstraints(constraints)

	assert.Nil(t, ev.Detail.Get(requestBody))
	assert.Nil(t, ev.Detail.Get(responseBody))
}

func TestMdtpAuditEnforceConstraintsCombinedTooLarge(t *testing.T) {

	max := 20
	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(max)}

	ev := &MdtpAuditEvent{Detail: types.DataMap{}}
	ev.Detail[requestBodyLen] = max - 1
	ev.Detail[requestBody] = "NotMaxBytes"
	ev.Detail[responseBodyLen] = 2
	ev.Detail[responseBody] = "IWillBreachTheAllowedLimit"

	ev.EnforceConstraints(constraints)

	assert.NotEmpty(t, ev.Detail.Get(requestBody))
	assert.Nil(t, ev.Detail.Get(responseBody))
}

func TestMdtpAuditEnforceConstraintsRemovesEmptyDetailAndTagsFields(t *testing.T) {

	constraints := AuditConstraints{MaxAuditLength: 1000, MaxPayloadContentsLength: int64(100)}

	ev := &MdtpAuditEvent{Detail: types.DataMap{}, Tags: types.DataMap{}}
	ev.Detail["field1"] = "somedata"
	ev.Detail["aNilField"] = nil
	ev.Detail["anEmptyField"] = ""

	ev.Tags["field2"] = "moredata"
	ev.Tags["aNilField"] = nil
	ev.Tags["anEmptyField"] = ""

	ev.EnforceConstraints(constraints)

	assert.Equal(t, types.DataMap{"field1": "somedata"}, ev.Detail)
	assert.Equal(t, types.DataMap{"field2": "moredata"}, ev.Tags)

}

func TestAuditSourceDerivation(t *testing.T) {
	assert.Equal(t, "my-app", deriveAuditSource("my-app.service"))
	assert.Equal(t, "my-app", deriveAuditSource("  my-app.service  "))
	assert.Equal(t, "my-app", deriveAuditSource("my-app.protected.mdtp"))
	assert.Equal(t, "my-app", deriveAuditSource("my-app.public.mdtp"))
}
