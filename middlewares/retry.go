package middlewares

import (
	"bufio"
	"context"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/containous/traefik/log"
)

// Compile time validation that the response writer implements http interfaces correctly.
var _ Stateful = &retryResponseWriterWithCloseNotify{}

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
		netErrorOccurred := false
		// We pass in a pointer to netErrorOccurred so that we can set it to true on network errors
		// when proxying the HTTP requests to the backends. This happens in the custom RecordingErrorHandler.
		newCtx := context.WithValue(r.Context(), defaultNetErrCtxKey, &netErrorOccurred)
		retryResponseWriter := newRetryResponseWriter(rw, attempts >= retry.attempts, &netErrorOccurred)

		retry.next.ServeHTTP(retryResponseWriter, r.WithContext(newCtx))
		if !retryResponseWriter.ShouldRetry() {
			break
		}

		attempts++
		log.Debugf("New attempt %d for request: %v", attempts, r.URL)
		retry.listener.Retried(r, attempts)
	}
}

// netErrorCtxKey is a custom type that is used as key for the context.
type netErrorCtxKey string

// defaultNetErrCtxKey is the actual key which value is used to record network errors.
var defaultNetErrCtxKey netErrorCtxKey = "NetErrCtxKey"

// NetErrorRecorder is an interface to record net errors.
type NetErrorRecorder interface {
	// Record can be used to signal the retry middleware that an network error happened
	// and therefore the request should be retried.
	Record(ctx context.Context)
}

// DefaultNetErrorRecorder is the default NetErrorRecorder implementation.
type DefaultNetErrorRecorder struct{}

// Record is recording network errors by setting the context value for the defaultNetErrCtxKey to true.
func (DefaultNetErrorRecorder) Record(ctx context.Context) {
	val := ctx.Value(defaultNetErrCtxKey)

	if netErrorOccurred, isBoolPointer := val.(*bool); isBoolPointer {
		*netErrorOccurred = true
	}
}

// RetryListener is used to inform about retry attempts.
type RetryListener interface {
	// Retried will be called when a retry happens, with the request attempt passed to it.
	// For the first retry this will be attempt 2.
	Retried(req *http.Request, attempt int)
}

// RetryListeners is a convenience type to construct a list of RetryListener and notify
// each of them about a retry attempt.
type RetryListeners []RetryListener

// Retried exists to implement the RetryListener interface. It calls Retried on each of its slice entries.
func (l RetryListeners) Retried(req *http.Request, attempt int) {
	for _, retryListener := range l {
		retryListener.Retried(req, attempt)
	}
}

type retryResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	ShouldRetry() bool
}

func newRetryResponseWriter(rw http.ResponseWriter, attemptsExhausted bool, netErrorOccured *bool) retryResponseWriter {
	responseWriter := &retryResponseWriterWithoutCloseNotify{
		responseWriter:    rw,
		attemptsExhausted: attemptsExhausted,
		netErrorOccured:   netErrorOccured,
	}
	if _, ok := rw.(http.CloseNotifier); ok {
		return &retryResponseWriterWithCloseNotify{responseWriter}
	}
	return responseWriter
}

type retryResponseWriterWithoutCloseNotify struct {
	responseWriter    http.ResponseWriter
	attemptsExhausted bool
	netErrorOccured   *bool
}

func (rr *retryResponseWriterWithoutCloseNotify) ShouldRetry() bool {
	return *rr.netErrorOccured && !rr.attemptsExhausted
}

func (rr *retryResponseWriterWithoutCloseNotify) Header() http.Header {
	if rr.ShouldRetry() {
		return make(http.Header)
	}
	return rr.responseWriter.Header()
}

func (rr *retryResponseWriterWithoutCloseNotify) Write(buf []byte) (int, error) {
	if rr.ShouldRetry() {
		return 0, nil
	}
	return rr.responseWriter.Write(buf)
}

func (rr *retryResponseWriterWithoutCloseNotify) WriteHeader(code int) {
	if rr.ShouldRetry() {
		return
	}
	rr.responseWriter.WriteHeader(code)
}

func (rr *retryResponseWriterWithoutCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rr.responseWriter.(http.Hijacker).Hijack()
}

func (rr *retryResponseWriterWithoutCloseNotify) Flush() {
	if flusher, ok := rr.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type retryResponseWriterWithCloseNotify struct {
	*retryResponseWriterWithoutCloseNotify
}

func (rr *retryResponseWriterWithCloseNotify) CloseNotify() <-chan bool {
	return rr.responseWriter.(http.CloseNotifier).CloseNotify()
}
