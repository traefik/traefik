package middlewares

import (
	"bufio"
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/vulcand/oxy/utils"
	"net"
	"net/http"
)

// Retry is a middleware that retries requests
type Retry struct {
	attempts int
	next     http.Handler
}

// NewRetry returns a new Retry instance
func NewRetry(attempts int, next http.Handler) *Retry {
	return &Retry{
		attempts: attempts,
		next:     next,
	}
}

func (retry *Retry) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	attempts := 1
	for {
		recorder := NewRecorder()
		recorder.responseWriter = rw
		retry.next.ServeHTTP(recorder, r)
		if !isNetworkError(recorder.Code) || attempts >= retry.attempts {
			utils.CopyHeaders(recorder.Header(), rw.Header())
			rw.WriteHeader(recorder.Code)
			rw.Write(recorder.Body.Bytes())
			break
		}
		attempts++
		log.Debugf("New attempt %d for request: %v", attempts, r.URL)
	}
}

func isNetworkError(status int) bool {
	return status == http.StatusBadGateway || status == http.StatusGatewayTimeout
}

// ResponseRecorder is an implementation of http.ResponseWriter that
// records its mutations for later inspection in tests.
type ResponseRecorder struct {
	Code      int           // the HTTP response code from WriteHeader
	HeaderMap http.Header   // the HTTP response headers
	Body      *bytes.Buffer // if non-nil, the bytes.Buffer to append written data to
	Flushed   bool

	wroteHeader    bool
	responseWriter http.ResponseWriter
}

// NewRecorder returns an initialized ResponseRecorder.
func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
		Code:      200,
	}
}

// Header returns the response headers.
func (rw *ResponseRecorder) Header() http.Header {
	m := rw.HeaderMap
	if m == nil {
		m = make(http.Header)
		rw.HeaderMap = m
	}
	return m
}

// writeHeader writes a header if it was not written yet and
// detects Content-Type if needed.
//
// bytes or str are the beginning of the response body.
// We pass both to avoid unnecessarily generate garbage
// in rw.WriteString which was created for performance reasons.
// Non-nil bytes win.
func (rw *ResponseRecorder) writeHeader(b []byte, str string) {
	if rw.wroteHeader {
		return
	}
	if len(str) > 512 {
		str = str[:512]
	}

	_, hasType := rw.HeaderMap["Content-Type"]
	hasTE := rw.HeaderMap.Get("Transfer-Encoding") != ""
	if !hasType && !hasTE {
		if b == nil {
			b = []byte(str)
		}
		if rw.HeaderMap == nil {
			rw.HeaderMap = make(http.Header)
		}
		rw.HeaderMap.Set("Content-Type", http.DetectContentType(b))
	}

	rw.WriteHeader(200)
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *ResponseRecorder) Write(buf []byte) (int, error) {
	rw.writeHeader(buf, "")
	if rw.Body != nil {
		rw.Body.Write(buf)
	}
	return len(buf), nil
}

// WriteString always succeeds and writes to rw.Body, if not nil.
func (rw *ResponseRecorder) WriteString(str string) (int, error) {
	rw.writeHeader(nil, str)
	if rw.Body != nil {
		rw.Body.WriteString(str)
	}
	return len(str), nil
}

// WriteHeader sets rw.Code.
func (rw *ResponseRecorder) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.Code = code
		rw.wroteHeader = true
	}
}

// Flush sets rw.Flushed to true.
func (rw *ResponseRecorder) Flush() {
	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}
	rw.Flushed = true
}

// Hijack hijacks the connection
func (rw *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.responseWriter.(http.Hijacker).Hijack()
}
