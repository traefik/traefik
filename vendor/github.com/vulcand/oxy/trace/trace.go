// Package trace implement structured logging of requests
package trace

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/vulcand/oxy/utils"
)

// Option is a functional option setter for Tracer
type Option func(*Tracer) error

// ErrorHandler is a functional argument that sets error handler of the server
func ErrorHandler(h utils.ErrorHandler) Option {
	return func(t *Tracer) error {
		t.errHandler = h
		return nil
	}
}

// RequestHeaders adds request headers to capture
func RequestHeaders(headers ...string) Option {
	return func(t *Tracer) error {
		t.reqHeaders = append(t.reqHeaders, headers...)
		return nil
	}
}

// ResponseHeaders adds response headers to capture
func ResponseHeaders(headers ...string) Option {
	return func(t *Tracer) error {
		t.respHeaders = append(t.respHeaders, headers...)
		return nil
	}
}

// Logger sets optional logger for trace used to report errors
func Logger(l utils.Logger) Option {
	return func(t *Tracer) error {
		t.log = l
		return nil
	}
}

// Tracer records request and response emitting JSON structured data to the output
type Tracer struct {
	errHandler  utils.ErrorHandler
	next        http.Handler
	reqHeaders  []string
	respHeaders []string
	writer      io.Writer
	log         utils.Logger
}

// New creates a new Tracer middleware that emits all the request/response information in structured format
// to writer and passes the request to the next handler. It can optionally capture request and response headers,
// see RequestHeaders and ResponseHeaders options for details.
func New(next http.Handler, writer io.Writer, opts ...Option) (*Tracer, error) {
	t := &Tracer{
		writer: writer,
		next:   next,
	}
	for _, o := range opts {
		if err := o(t); err != nil {
			return nil, err
		}
	}
	if t.errHandler == nil {
		t.errHandler = utils.DefaultHandler
	}
	if t.log == nil {
		t.log = utils.NullLogger
	}
	return t, nil
}

func (t *Tracer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	pw := &utils.ProxyWriter{W: w}
	t.next.ServeHTTP(pw, req)

	l := t.newRecord(req, pw, time.Since(start))
	if err := json.NewEncoder(t.writer).Encode(l); err != nil {
		t.log.Errorf("Failed to marshal request: %v", err)
	}
}

func (t *Tracer) newRecord(req *http.Request, pw *utils.ProxyWriter, diff time.Duration) *Record {
	return &Record{
		Request: Request{
			Method:    req.Method,
			URL:       req.URL.String(),
			TLS:       newTLS(req),
			BodyBytes: bodyBytes(req.Header),
			Headers:   captureHeaders(req.Header, t.reqHeaders),
		},
		Response: Response{
			Code:      pw.StatusCode(),
			BodyBytes: bodyBytes(pw.Header()),
			Roundtrip: float64(diff) / float64(time.Millisecond),
			Headers:   captureHeaders(pw.Header(), t.respHeaders),
		},
	}
}

func newTLS(req *http.Request) *TLS {
	if req.TLS == nil {
		return nil
	}
	return &TLS{
		Version:     versionToString(req.TLS.Version),
		Resume:      req.TLS.DidResume,
		CipherSuite: csToString(req.TLS.CipherSuite),
		Server:      req.TLS.ServerName,
	}
}

func captureHeaders(in http.Header, headers []string) http.Header {
	if len(headers) == 0 || in == nil {
		return nil
	}
	out := make(http.Header, len(headers))
	for _, h := range headers {
		vals, ok := in[h]
		if !ok || len(out[h]) != 0 {
			continue
		}
		for i := range vals {
			out.Add(h, vals[i])
		}
	}
	return out
}

// Record represents a structured request and response record
type Record struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// Req contains information about an HTTP request
type Request struct {
	Method    string      `json:"method"`            // Method - request method
	BodyBytes int64       `json:"body_bytes"`        // BodyBytes - size of request body in bytes
	URL       string      `json:"url"`               // URL - Request URL
	Headers   http.Header `json:"headers,omitempty"` // Headers - optional request headers, will be recorded if configured
	TLS       *TLS        `json:"tls,omitempty"`     // TLS - optional TLS record, will be recorded if it's a TLS connection
}

// Resp contains information about HTTP response
type Response struct {
	Code      int         `json:"code"`              // Code - response status code
	Roundtrip float64     `json:"roundtrip"`         // Roundtrip - round trip time in milliseconds
	Headers   http.Header `json:"headers,omitempty"` // Headers - optional headers, will be recorded if configured
	BodyBytes int64       `json:"body_bytes"`        // BodyBytes - size of response body in bytes
}

// TLS contains information about this TLS connection
type TLS struct {
	Version     string `json:"version"`      // Version - TLS version
	Resume      bool   `json:"resume"`       // Resume tells if the session has been re-used (session tickets)
	CipherSuite string `json:"cipher_suite"` // CipherSuite contains cipher suite used for this connection
	Server      string `json:"server"`       // Server contains server name used in SNI
}

func versionToString(v uint16) string {
	switch v {
	case tls.VersionSSL30:
		return "SSL30"
	case tls.VersionTLS10:
		return "TLS10"
	case tls.VersionTLS11:
		return "TLS11"
	case tls.VersionTLS12:
		return "TLS12"
	}
	return fmt.Sprintf("unknown: %x", v)
}

func csToString(cs uint16) string {
	switch cs {
	case tls.TLS_RSA_WITH_RC4_128_SHA:
		return "TLS_RSA_WITH_RC4_128_SHA"
	case tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:
		return "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA:
		return "TLS_RSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_256_CBC_SHA:
		return "TLS_RSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:
		return "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:
		return "TLS_ECDHE_RSA_WITH_RC4_128_SHA"
	case tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:
		return "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:
		return "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:
		return "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	}
	return fmt.Sprintf("unknown: %x", cs)
}

func bodyBytes(h http.Header) int64 {
	len := h.Get("Content-Length")
	if len == "" {
		return 0
	}
	bytes, err := strconv.ParseInt(len, 10, 0)
	if err == nil {
		return bytes
	}
	return 0
}
