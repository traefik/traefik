package middlewares

import (
	"bufio"
	"bytes"
	"github.com/containous/traefik/log"
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
			utils.CopyHeaders(rw.Header(), recorder.Header())
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

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *ResponseRecorder) Write(buf []byte) (int, error) {
	if rw.Body != nil {
		return rw.Body.Write(buf)
	}
	return 0, nil
}

// WriteHeader sets rw.Code.
func (rw *ResponseRecorder) WriteHeader(code int) {
	rw.Code = code
}

// Hijack hijacks the connection
func (rw *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.responseWriter.(http.Hijacker).Hijack()
}
