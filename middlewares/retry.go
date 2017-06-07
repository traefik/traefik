package middlewares

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/vulcand/oxy/utils"
)

// Compile time validation responseRecorder implements http interfaces correctly.
var (
	_ Stateful = &retryResponseRecorder{}
)

// Retry is a middleware that retries requests
type Retry struct {
	attempts int
	next     http.Handler
	listener RetryListener
}

// NewRetry returns a new Retry instance
func NewRetry(attempts int, next http.Handler, listener RetryListener) *Retry {
	return &Retry{
		attempts: attempts,
		next:     next,
		listener: listener,
	}
}

func (retry *Retry) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// if we might make multiple attempts, swap the body for an ioutil.NopCloser
	// cf https://github.com/containous/traefik/issues/1008
	if retry.attempts > 1 {
		body := r.Body
		defer body.Close()
		r.Body = ioutil.NopCloser(body)
	}
	attempts := 1
	for {
		recorder := newRetryResponseRecorder()
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
		retry.listener.Retried(attempts)
	}
}

func isNetworkError(status int) bool {
	return status == http.StatusBadGateway || status == http.StatusGatewayTimeout
}

// RetryListener is used to inform about retry attempts.
type RetryListener interface {
	// Retried will be called when a retry happens, with the request attempt passed to it.
	// For the first retry this will be attempt 2.
	Retried(attempt int)
}

// retryResponseRecorder is an implementation of http.ResponseWriter that
// records its mutations for later inspection.
type retryResponseRecorder struct {
	Code      int           // the HTTP response code from WriteHeader
	HeaderMap http.Header   // the HTTP response headers
	Body      *bytes.Buffer // if non-nil, the bytes.Buffer to append written data to

	responseWriter http.ResponseWriter
	err            error
}

// newRetryResponseRecorder returns an initialized retryResponseRecorder.
func newRetryResponseRecorder() *retryResponseRecorder {
	return &retryResponseRecorder{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
		Code:      200,
	}
}

// Header returns the response headers.
func (rw *retryResponseRecorder) Header() http.Header {
	m := rw.HeaderMap
	if m == nil {
		m = make(http.Header)
		rw.HeaderMap = m
	}
	return m
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *retryResponseRecorder) Write(buf []byte) (int, error) {
	if rw.err != nil {
		return 0, rw.err
	}
	return rw.Body.Write(buf)
}

// WriteHeader sets rw.Code.
func (rw *retryResponseRecorder) WriteHeader(code int) {
	rw.Code = code
}

// Hijack hijacks the connection
func (rw *retryResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.responseWriter.(http.Hijacker).Hijack()
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone
// away.
func (rw *retryResponseRecorder) CloseNotify() <-chan bool {
	return rw.responseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush sends any buffered data to the client.
func (rw *retryResponseRecorder) Flush() {
	_, err := rw.responseWriter.Write(rw.Body.Bytes())
	if err != nil {
		log.Errorf("Error writing response in retryResponseRecorder: %s", err)
		rw.err = err
	}
	rw.Body.Reset()
	flusher, ok := rw.responseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}
