package audittypes

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"strconv"

	ahttp "github.com/containous/traefik/middlewares/audittap/http"
	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/satori/go.uuid"
)

const (
	keyPayloadContents = "contents"
	keyPayloadLength   = "length"
)

// AuditEvent captures the content and metadata of an HTTP request and response.
type AuditEvent struct {
	AuditSource     string        `json:"auditSource,omitempty"`
	AuditType       string        `json:"auditType,omitempty"`
	EventID         string        `json:"eventID,omitempty"`
	GeneratedAt     string        `json:"generatedAt,omitempty"`
	Version         string        `json:"version,omitempty"`
	RequestID       string        `json:"requestID,omitempty"`
	Method          string        `json:"method,omitempty"`
	Path            string        `json:"path,omitempty"`
	QueryString     string        `json:"queryString,omitempty"`
	ClientIP        string        `json:"clientIP,omitempty"`
	ClientPort      string        `json:"clientPort,omitempty"`
	ReceivingIP     string        `json:"receivingIP,omitempty"`
	ResponseStatus  string        `json:"responseStatus,omitempty"`
	ResponsePayload types.DataMap `json:"responsePayload,omitempty"`
	ClientHeaders   types.DataMap `json:"clientHeaders,omitempty"`
	RequestHeaders  types.DataMap `json:"requestHeaders,omitempty"`
	RequestPayload  types.DataMap `json:"requestPayload,omitempty"`
	ResponseHeaders types.DataMap `json:"responseHeaders,omitempty"`
}

// AuditConstraints defines validation constraints an audit event must satisfy
type AuditConstraints struct {
	MaxAuditLength           int64
	MaxPayloadContentsLength int64
}

// AuditObfuscation defines obfuscation of sensitive data in audits
type AuditObfuscation struct {
	MaskFields []string
	MaskValue  string
}

// HeaderMapping defines a field whose value is sourced from a header
type HeaderMapping map[string]string

// HeaderMappings defines the dynamic mappings to be applied for a section of an audit event.
type HeaderMappings map[string]HeaderMapping

// Filter defines filter conditions
type Filter struct {
	Source     string
	Contains   []string
	EndsWith   []string
	StartsWith []string
	Matches    []*regexp.Regexp
}

// AuditSpecification groups together configuration used to define the structure of audit events.
type AuditSpecification struct {
	AuditConstraints
	AuditObfuscation
	HeaderMappings
	Inclusions []*Filter
	Exclusions []*Filter
}

// AuditStream describes a type to which audit events can be sent.
type AuditStream interface {
	io.Closer
	Audit(encoder types.Encoded) error
}

// RequestContext wraps a HTTP requests and exposes useful metadatfor the request
type RequestContext struct {
	Req *http.Request
	*url.URL
	FlatHeaders    types.DataMap
	ClientHeaders  types.DataMap
	RequestHeaders types.DataMap
}

// NewRequestContext creates a new requestContext
func NewRequestContext(r *http.Request) *RequestContext {

	hdr := ahttp.NewHeaders(r.Header).SimplifyCookies()

	// Need to create a URL from the RequestURI, because the URL in the request is overwritten
	// by oxy'tap RoundRobin and loses Path and RawQuery
	u, _ := url.ParseRequestURI(r.RequestURI)

	clientHeaders, requestHeaders := hdr.ClientAndRequestHeaders()

	return &RequestContext{
		Req:            r,
		URL:            u,
		FlatHeaders:    hdr.Flatten(),
		ClientHeaders:  clientHeaders,
		RequestHeaders: requestHeaders,
	}
}

// Auditer is a type that audits information from a HTTP request and response
type Auditer interface {
	AppendRequest(ctx *RequestContext, auditSpec *AuditSpecification)
	AppendResponse(responseHeaders http.Header, resp types.ResponseInfo, auditSpec *AuditSpecification)
	// EnforceConstraints ensures the audit event complies with rules for the audit type
	// returns true if audit event is valid for auditing
	EnforceConstraints(constraints AuditConstraints) bool
	types.Encodeable
}

func appendCommonRequestFields(ev *AuditEvent, ctx *RequestContext) {

	requestPayload := types.DataMap{}
	requestPayload["length"] = ctx.Req.ContentLength

	var requestContentType = ctx.FlatHeaders.GetString("content-type")

	if requestContentType != "" {
		requestPayload["type"] = requestContentType
	}

	ev.EventID = uuid.NewV4().String()
	ev.GeneratedAt = types.TheClock.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
	ev.Version = "1"
	ev.RequestID = ctx.FlatHeaders.GetString("x-request-id")
	ev.Method = ctx.Req.Method
	ev.Path = ctx.Path
	ev.QueryString = ctx.RawQuery
	ev.ClientIP = ctx.FlatHeaders.GetString("true-client-ip")
	ev.ClientPort = ctx.FlatHeaders.GetString("true-client-port")
	ev.ReceivingIP = ctx.FlatHeaders.GetString("x-source")
	ev.ClientHeaders = ctx.ClientHeaders
	ev.RequestHeaders = ctx.RequestHeaders
	ev.RequestPayload = requestPayload
	ev.ResponseHeaders = nil
	ev.ResponseStatus = ""
	ev.ResponsePayload = nil
}

func appendCommonResponseFields(ev *AuditEvent, responseHeaders http.Header, info types.ResponseInfo) types.DataMap {

	headers := ahttp.NewHeaders(responseHeaders).SimplifyCookies()
	flatHeaders := headers.Flatten()

	ev.ResponseStatus = strconv.Itoa(info.Status)
	ev.ResponseHeaders = headers.ResponseHeaders()
	ev.ResponsePayload = types.DataMap{}
	if ct := flatHeaders.GetString("content-type"); ct != "" {
		ev.ResponsePayload["type"] = ct
	}

	return flatHeaders
}

func (ev *AuditEvent) addRequestPayloadContents(s string) {
	if ev.RequestPayload == nil {
		ev.RequestPayload = types.DataMap{}
	}
	ev.RequestPayload[keyPayloadContents] = s
	ev.RequestPayload[keyPayloadLength] = len(s)
}

func (ev *AuditEvent) addResponsePayloadContents(s string) {
	if ev.ResponsePayload == nil {
		ev.ResponsePayload = types.DataMap{}
	}
	ev.ResponsePayload[keyPayloadContents] = s
	ev.ResponsePayload[keyPayloadLength] = len(s)
}

// ObfuscateURLEncoded applies the obfuscation criteria against the supplied byte slice than contains
// URLEncoded key=value content
func (obs *AuditObfuscation) ObfuscateURLEncoded(b []byte) ([]byte, error) {
	src := b
	for _, field := range obs.MaskFields {
		expr := fmt.Sprintf("(%s=[^\\&]+)", field)
		re, err := regexp.Compile(expr)
		replacement := field + "=" + obs.MaskValue
		if err != nil {
			return nil, fmt.Errorf("Obfuscation error for required mask '%s'. %v", expr, err)
		}
		src = re.ReplaceAll(src, []byte(replacement))
	}
	return src, nil
}

func enforcePrecedentConstraints(ev *AuditEvent, constraints AuditConstraints) {
	reqLen, _ := ev.RequestPayload[keyPayloadLength].(int) // Zero if not int or missing
	lenRequest := int64(reqLen)
	requestTooBig := lenRequest > constraints.MaxPayloadContentsLength
	if lenRequest == 0 || requestTooBig {
		delete(ev.RequestPayload, keyPayloadContents)
		lenRequest = 0
	}

	respLen, _ := ev.ResponsePayload[keyPayloadLength].(int)
	lenResponse := int64(respLen)
	responseTooBig := lenResponse > constraints.MaxPayloadContentsLength
	combinedTooBig := lenRequest+lenResponse > constraints.MaxPayloadContentsLength
	if lenResponse == 0 || responseTooBig || combinedTooBig {
		delete(ev.ResponsePayload, keyPayloadContents)
	}
}

// Need to take a copy of the body contents so a fresh reader for body is available to subsequent handlers
func copyRequestBody(req *http.Request) ([]byte, int, error) {
	buf, err := ioutil.ReadAll(req.Body)
	if err == nil {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		return buf, len(buf), nil
	}
	return nil, 0, err
}

// ShouldAudit asserts if request metadata matches specified exclusions from config
func ShouldAudit(rc *RequestContext, spec *AuditSpecification) bool {

	include := false
	for _, inc := range spec.Inclusions {
		if rc.satisfiesFilter(inc) {
			include = true
			break
		}
	}

	if !include && len(spec.Inclusions) > 0 {
		return false
	}

	for _, exc := range spec.Exclusions {
		if rc.satisfiesFilter(exc) {
			return false
		}
	}

	return true
}

func (rc *RequestContext) satisfiesFilter(filter *Filter) bool {
	lcHdr := strings.ToLower(filter.Source)
	// Get host or path direct from request
	if (lcHdr == "host" || lcHdr == "requesthost") && filter.satisfiedBy(rc.Req.Host) {
		return true
	} else if lcHdr == "path" || lcHdr == "requestpath" {
		if filter.satisfiedBy(rc.URL.Path) {
			return true
		}
	} else if filter.satisfiedBy(rc.Req.Header.Get(filter.Source)) {
		return true
	}
	return false
}

// SatisfiedBy checks if this filter satisfies the supplied value
func (f *Filter) satisfiedBy(s string) bool {
	return matchesTerm(s, f.StartsWith, strings.HasPrefix) ||
		matchesTerm(s, f.EndsWith, strings.HasSuffix) ||
		matchesTerm(s, f.Contains, strings.Contains) ||
		matchesRegex(s, f.Matches)
}

func matchesTerm(v string, terms []string, fn func(string, string) bool) bool {
	if v != "" {
		for _, x := range terms {
			if fn(v, x) {
				return true
			}
		}

	}
	return false
}

func matchesRegex(v string, expressions []*regexp.Regexp) bool {
	if v != "" {
		for _, re := range expressions {
			if re.Match([]byte(v)) {
				return true
			}
		}

	}
	return false
}
