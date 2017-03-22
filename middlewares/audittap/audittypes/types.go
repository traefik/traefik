package audittypes

import (
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
	Source   string  `json:"auditSource,omitempty"`
	Request  DataMap `json:"request"`
	Response DataMap `json:"response"`
}

// AuditResponseWriter is an extended ResponseWriter that also provides a summary of the request and response.
type AuditResponseWriter interface {
	http.ResponseWriter
	SummariseResponse() DataMap
}

// AuditStream describes a type to which audit summaries can be sent.
type AuditStream interface {
	io.Closer
	Audit(summary Summary) error
}

//-------------------------------------------------------------------------------------------------

// ToJSON converts the summary to JSON
func (summary Summary) ToJSON() Encoded {
	b, err := json.Marshal(summary)
	return Encoded{b, err}
}

//-------------------------------------------------------------------------------------------------

// AddAll merges a datamap
func (m DataMap) AddAll(other DataMap) DataMap {
	for k, v := range other {
		m[k] = v
	}
	return m
}
