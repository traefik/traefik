package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
)

// SerializableHttpRequest serializable HTTP request
type SerializableHttpRequest struct {
	Method           string
	URL              *url.URL
	Proto            string // "HTTP/1.0"
	ProtoMajor       int    // 1
	ProtoMinor       int    // 0
	Header           http.Header
	ContentLength    int64
	TransferEncoding []string
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	TLS              *tls.ConnectionState
}

// Clone clone a request
func Clone(r *http.Request) *SerializableHttpRequest {
	if r == nil {
		return nil
	}

	rc := new(SerializableHttpRequest)
	rc.Method = r.Method
	rc.URL = r.URL
	rc.Proto = r.Proto
	rc.ProtoMajor = r.ProtoMajor
	rc.ProtoMinor = r.ProtoMinor
	rc.Header = r.Header
	rc.ContentLength = r.ContentLength
	rc.Host = r.Host
	rc.RemoteAddr = r.RemoteAddr
	rc.RequestURI = r.RequestURI
	return rc
}

// ToJson serializes to JSON
func (s *SerializableHttpRequest) ToJson() string {
	jsonVal, err := json.Marshal(s)
	if err != nil || jsonVal == nil {
		return fmt.Sprintf("Error marshalling SerializableHttpRequest to json: %s", err)
	}
	return string(jsonVal)
}

// DumpHttpRequest dump a HTTP request to JSON
func DumpHttpRequest(req *http.Request) string {
	return Clone(req).ToJson()
}
