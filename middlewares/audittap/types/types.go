package types

import (
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

// UnclassifiedRequest is used as the auditType when no explicit value is configured or can be derived
const UnclassifiedRequest = "UnclassifiedRequest"

// DataMap holds headers in which the values are all either string or []string.
type DataMap map[string]interface{}

// ResponseInfo is a summary of an HTTP response
type ResponseInfo struct {
	Status          int
	Size            int
	Entity          []byte
	MaxEntityLength int
}

// AuditResponseWriter is an extended ResponseWriter that also provides a ev of the request and response.
type AuditResponseWriter interface {
	http.ResponseWriter
	GetResponseInfo() ResponseInfo
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

// GetDataMap gets a element whose value is expected to be a DataMap
// returning either the value or empty if it's not present
func (m DataMap) GetDataMap(key string) DataMap {
	var i = m.Get(key)
	if sm, ok := i.(DataMap); ok {
		return sm
	}
	return DataMap{}
}

// PurgeEmptyValues removes nil and empty values from this map
func (m DataMap) PurgeEmptyValues() {
	for k, v := range m {
		if v == nil || v == "" {
			delete(m, k)
		}
	}
}
