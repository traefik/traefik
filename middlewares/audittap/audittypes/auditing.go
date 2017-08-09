package audittypes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	ahttp "github.com/containous/traefik/middlewares/audittap/http"
	"github.com/containous/traefik/middlewares/audittap/types"
	"github.com/satori/go.uuid"
	"strconv"
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

// AuditStream describes a type to which audit events can be sent.
type AuditStream interface {
	io.Closer
	Audit(encoder types.Encodeable) error
}

// Auditer is a type that audits information from a HTTP request and response
type Auditer interface {
	AppendRequest(req *http.Request)
	AppendResponse(responseHeaders http.Header, resp types.ResponseInfo)
	types.Encodeable
}

func appendCommonRequestFields(ev *AuditEvent, req *http.Request) types.DataMap {
	hdr := ahttp.NewHeaders(req.Header).SimplifyCookies()

	// Need to create a URL from the RequestURI, because the URL in the request is overwritten
	// by oxy'tap RoundRobin and loses Path and RawQuery
	u, _ := url.ParseRequestURI(req.RequestURI)

	clientHeaders, requestHeaders := hdr.ClientAndRequestHeaders()
	flatHdr := hdr.Flatten()

	requestPayload := types.DataMap{}

	var requestContentType = flatHdr.GetString("content-type")

	if requestContentType != "" {
		requestPayload["type"] = requestContentType
	}

	ev.EventID = uuid.NewV4().String()
	ev.GeneratedAt = types.TheClock.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
	ev.Version = "1"
	ev.RequestID = flatHdr.GetString("x-request-id")
	ev.Method = req.Method
	ev.Path = u.Path
	ev.QueryString = u.RawQuery
	ev.ClientIP = flatHdr.GetString("true-client-ip")
	ev.ClientPort = flatHdr.GetString("true-client-port")
	ev.ReceivingIP = flatHdr.GetString("x-source")
	ev.ClientHeaders = clientHeaders
	ev.RequestHeaders = requestHeaders
	ev.RequestPayload = requestPayload
	ev.ResponseHeaders = nil
	ev.ResponseStatus = ""
	ev.ResponsePayload = nil

	return flatHdr
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

//-------------------------------------------------------------------------------------------------

// EncodeToJSON transforms event event to JSON and then to bytes
func EncodeToJSON(event interface{}) types.Encoded {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(event)
	return types.Encoded{buffer.Bytes(), err}
}
