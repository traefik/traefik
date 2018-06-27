package audittypes

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/containous/traefik/log"
	ahttp "github.com/containous/traefik/middlewares/audittap/http"
	"github.com/containous/traefik/middlewares/audittap/types"
	uuid "github.com/satori/go.uuid"
)

const (
	requestBody         = "requestBody"
	requestBodyLen      = "requestBodyLen"
	requestContentType  = "requestContentType"
	responseBodyLen     = "responseBodyLen"
	responseBody        = "responseMessage"
	responseContentType = "responseContentType"
)

// MdtpAuditEvent is the audit event created for API calls
type MdtpAuditEvent struct {
	AuditSource     string        `json:"auditSource,omitempty"`
	AuditType       string        `json:"auditType,omitempty"`
	EventID         string        `json:"eventId,omitempty"`
	GeneratedAt     string        `json:"generatedAt,omitempty"`
	Version         string        `json:"version,omitempty"`
	ClientHeaders   types.DataMap `json:"clientHeaders,omitempty"`
	RequestHeaders  types.DataMap `json:"requestHeaders,omitempty"`
	ResponseHeaders types.DataMap `json:"responseHeaders,omitempty"`
	Detail          types.DataMap `json:"detail,omitempty"`
	Tags            types.DataMap `json:"tags,omitempty"`
}

// AppendRequest appends information about the request to the audit event
func (ev *MdtpAuditEvent) AppendRequest(req *http.Request, auditSpec *AuditSpecification) {

	ev.AuditSource = deriveAuditSource(req.Host)
	ev.AuditType = "RequestReceived"
	ev.EventID = uuid.NewV4().String()
	ev.GeneratedAt = types.TheClock.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
	ev.Version = "1"

	if ev.Detail == nil {
		ev.Detail = types.DataMap{}
	}
	if ev.Tags == nil {
		ev.Tags = types.DataMap{}
	}

	hdr := ahttp.NewHeaders(req.Header).SimplifyCookies()
	flatHdr := hdr.Flatten()

	clientHeaders, requestHeaders := hdr.ClientAndRequestHeaders()
	ev.ClientHeaders = clientHeaders
	ev.RequestHeaders = requestHeaders

	url, _ := url.ParseRequestURI(req.RequestURI)
	ev.Detail.AddAll(detailFromRequest(req, url, flatHdr, auditSpec))
	ev.Tags.AddAll(tagsFromRequest(req, url, flatHdr, auditSpec))
}

// AppendResponse appends information about the response to the audit event
func (ev *MdtpAuditEvent) AppendResponse(responseHeaders http.Header, respInfo types.ResponseInfo, auditSpec *AuditSpecification) {
	headers := ahttp.NewHeaders(responseHeaders).SimplifyCookies()
	flatHeaders := headers.Flatten()
	ev.ResponseHeaders = headers.ResponseHeaders()
	ev.Detail.AddAll(detailFromResponse(respInfo, flatHeaders))
	ev.Tags.AddAll(tagsFromResponse(respInfo, flatHeaders))
}

// EnforceConstraints ensures the audit event satisfies constraints
func (ev *MdtpAuditEvent) EnforceConstraints(constraints AuditConstraints) bool {
	ev.Detail.PurgeEmptyValues()
	ev.Tags.PurgeEmptyValues()
	reqLen, _ := ev.Detail[requestBodyLen].(int) // Zero if not int or missing
	lenRequest := int64(reqLen)
	requestTooBig := lenRequest > constraints.MaxPayloadContentsLength
	if lenRequest == 0 || requestTooBig {
		delete(ev.Detail, requestBody)
		lenRequest = 0
	}

	respLen, _ := ev.Detail[responseBodyLen].(int)
	lenResponse := int64(respLen)
	responseTooBig := lenResponse > constraints.MaxPayloadContentsLength
	combinedTooBig := lenRequest+lenResponse > constraints.MaxPayloadContentsLength
	if lenResponse == 0 || responseTooBig || combinedTooBig {
		delete(ev.Detail, responseBody)
	}
	return true
}

// ToEncoded transforms the event into an Encoded
func (ev *MdtpAuditEvent) ToEncoded() types.Encoded {
	return types.ToEncoded(ev)
}

// NewMdtpAuditEvent creates a new MDTP AuditEvent with the provided auditSource and auditType
func NewMdtpAuditEvent() Auditer {
	ev := MdtpAuditEvent{Detail: types.DataMap{}, Tags: types.DataMap{}}
	return &ev
}

func deriveAuditSource(s string) string {
	re := regexp.MustCompile("^(.*)\\.(service|public\\.mdtp|protected\\.mdtp)$")
	matches := re.FindStringSubmatch(strings.TrimSpace(s))
	auditSource := s
	if len(matches) == 3 {
		auditSource = matches[1]
	}
	log.Debugf("Derived auditSource is %s", auditSource)
	return auditSource
}

func detailFromRequest(req *http.Request, url *url.URL, headers types.DataMap, spec *AuditSpecification) types.DataMap {
	m := types.DataMap{}

	m["Authorization"] = headers.GetString("authorization")
	m["deviceID"] = extractDeviceID(req, headers)
	m["host"] = req.Host
	m["input"] = fmt.Sprintf("Request to %s", url.Path)
	m["ipAddress"] = headers.GetString("x-forwarded-for")
	m["method"] = req.Method
	m["port"] = ""
	m["queryString"] = url.RawQuery
	m["referrer"] = req.Referer()
	m["surrogate"] = headers.GetString("surrogate")
	m["token"] = headers.GetString("token")
	m["userAgentString"] = headers.GetString("user-agent")

	if df, err := extractDeviceFingerprint(req); err == nil {
		m["deviceFingerprint"] = df
	}

	ct := headers.GetString("content-type")
	m[requestContentType] = ct
	if body, _, err := copyRequestBody(req); err == nil {
		m[requestBodyLen] = len(body)
		// Obfuscation only applies to form requests
		if ct == "application/x-www-form-urlencoded" && &spec.AuditObfuscation != nil {
			if sanitised, err := spec.AuditObfuscation.ObfuscateURLEncoded(body); err == nil {
				m[requestBody] = strings.TrimSpace(string(sanitised))
			}
		} else {
			m[requestBody] = strings.TrimSpace(string(body))
		}
	}

	m.AddAll(extractAdditionalHeaders(headers, "detail", spec.HeaderMappings))

	return m
}

func tagsFromRequest(req *http.Request, url *url.URL, headers types.DataMap, spec *AuditSpecification) types.DataMap {
	m := types.DataMap{}
	m["X-Request-ID"] = headers.GetString("x-request-id")
	m["X-Session-ID"] = headers.GetString("x-session-id")
	m["clientIP"] = headers.GetString("true-client-ip")
	m["clientPort"] = headers.GetString("true-client-port")
	m["path"] = url.Path
	m["transactionName"] = url.RequestURI()
	m["Akamai-Reputation"] = headers.GetString("akamai-reputation")
	m.AddAll(extractAdditionalHeaders(headers, "tags", spec.HeaderMappings))
	return m
}

func detailFromResponse(respInfo types.ResponseInfo, headers types.DataMap) types.DataMap {
	m := types.DataMap{}
	m["statusCode"] = strconv.Itoa(respInfo.Status)
	m["Location"] = headers.GetString("location")

	ct := headers.GetString("content-type")
	m[responseContentType] = ct
	m[responseBodyLen] = len(respInfo.Entity)

	if ct == "text/html" {
		m[responseBody] = "<HTML>...</HTML>"
	} else {
		m[responseBody] = strings.TrimSpace(string(respInfo.Entity))
	}

	return m
}

func tagsFromResponse(respInfo types.ResponseInfo, headers types.DataMap) types.DataMap {
	m := types.DataMap{}
	return m
}

func extractDeviceFingerprint(req *http.Request) (string, error) {
	df, err := req.Cookie("mdtpdf")
	if err != nil {
		return "", err
	}
	decoded, err := base64.StdEncoding.DecodeString(df.Value)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func extractDeviceID(req *http.Request, headers types.DataMap) string {
	deviceID := headers.GetString("deviceID")
	if deviceID == "" {
		if cookie, err := req.Cookie("mdtpdi"); err == nil {
			deviceID = cookie.Value
		}
	}
	return deviceID
}

func extractAdditionalHeaders(headers types.DataMap, section string, mappings HeaderMappings) types.DataMap {
	m := types.DataMap{}
	log.Debugf("Adding additional fields to %s. +%v +%v", section, mappings, headers)
	if mappings != nil {
		for k, v := range mappings[section] {
			m[k] = headers.GetString(strings.ToLower(v))
		}
	}
	return m
}
