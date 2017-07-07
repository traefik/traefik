package audittypes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Request data keys
const (
	Host       = "host"
	Method     = "method"
	Path       = "path"
	Query      = "query"
	RemoteAddr = "remoteAddr"
	BeganAt    = "beganAt"
)

// Response data keys
const (
	Status      = "status"
	Size        = "size"
	CompletedAt = "completedAt"
)

// Request and Response data keys
const (
	Entity = "entity"
)

// DataMap holds headers in which the values are all either string or []string.
type DataMap map[string]interface{}

// Summary captures the content and metadata of an HTTP request and response.
type Summary struct {
	AuditSource        string  `json:"auditSource,omitempty"`
	AuditType          string  `json:"auditType,omitempty"`
	EventID            string  `json:"eventID,omitempty"`
	GeneratedAt        string  `json:"generatedAt,omitempty"`
	Version            string  `json:"version,omitempty"`
	RequestID          string  `json:"requestID,omitempty"`
	Method             string  `json:"method,omitempty"`
	Path               string  `json:"path,omitempty"`
	QueryString        string  `json:"queryString,omitempty"`
	ClientIP           string  `json:"clientIP,omitempty"`
	ClientPort         string  `json:"clientPort,omitempty"`
	ReceivingIP        string  `json:"receivingIP,omitempty"`
	AuthorisationToken string  `json:"authorisationToken,omitempty"`
	ResponseStatus     string  `json:"responseStatus,omitempty"`
	ResponsePayload    DataMap `json:"responsePayload"`
	ClientHeaders      DataMap `json:"clientHeaders"`
	RequestHeaders     DataMap `json:"requestHeaders"`
	RequestPayload     DataMap `json:"requestPayload"`
	ResponseHeaders    DataMap `json:"responseHeaders"`
}

// AuditResponseWriter is an extended ResponseWriter that also provides a summary of the request and response.
type AuditResponseWriter interface {
	http.ResponseWriter
	SummariseResponse(summary *Summary)
}

// AuditStream describes a type to which audit summaries can be sent.
type AuditStream interface {
	io.Closer
	Audit(summary Summary) error
}

//-------------------------------------------------------------------------------------------------

// ToJSON converts the summary to JSON
func (summary Summary) ToJSON() Encoded {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(summary)
	return Encoded{buffer.Bytes(), err}
}

//-------------------------------------------------------------------------------------------------

// AddAll merges a datamap
func (m DataMap) AddAll(other DataMap) DataMap {
	for k, v := range other {
		m[k] = v
	}
	return m
}

// Get gets a key from the datamap
func (m DataMap) Get(key string) interface{} {
	return m[key]
}

// GetString gets a string key from the datamap, returning either
// the string or an empty string if it's not present
func (m DataMap) GetString(key string) string {
	var i = m.Get(key)

	s, ok := i.(string)

	if ok {
		return s
	}

	return ""
}
