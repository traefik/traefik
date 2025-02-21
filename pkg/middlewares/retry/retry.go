package retry

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

// Compile time validation that the response writer implements http interfaces correctly.
var _ middlewares.Stateful = &responseWriterWithCloseNotify{}

const (
	typeName = "Retry"
)

// Listener is used to inform about retry attempts.
type Listener interface {
	// Retried will be called when a retry happens, with the request attempt passed to it.
	// For the first retry this will be attempt 2.
	Retried(req *http.Request, attempt int)
}

// Listeners is a convenience type to construct a list of Listener and notify
// each of them about a retry attempt.
type Listeners []Listener

// Retried exists to implement the Listener interface. It calls Retried on each of its slice entries.
func (l Listeners) Retried(req *http.Request, attempt int) {
	for _, listener := range l {
		listener.Retried(req, attempt)
	}
}

type shouldRetryContextKey struct{}

// ShouldRetry is a function allowing to enable/disable the retry middleware mechanism.
type ShouldRetry func(shouldRetry bool)

// ContextShouldRetry returns the ShouldRetry function if it has been set by the Retry middleware in the chain.
func ContextShouldRetry(ctx context.Context) ShouldRetry {
	f, _ := ctx.Value(shouldRetryContextKey{}).(ShouldRetry)
	return f
}

// WrapHandler wraps a given http.Handler to inject the httptrace.ClientTrace in the request context when it is needed
// by the retry middleware.
func WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if shouldRetry := ContextShouldRetry(req.Context()); shouldRetry != nil {
			shouldRetry(true)

			trace := &httptrace.ClientTrace{
				WroteHeaders: func() {
					shouldRetry(false)
				},
				WroteRequest: func(httptrace.WroteRequestInfo) {
					shouldRetry(false)
				},
			}
			newCtx := httptrace.WithClientTrace(req.Context(), trace)
			next.ServeHTTP(rw, req.WithContext(newCtx))
			return
		}

		next.ServeHTTP(rw, req)
	})
}

// retry is a middleware that retries requests.
type retry struct {
	attempts        int
	initialInterval time.Duration
	next            http.Handler
	listener        Listener
	name            string
}

// New returns a new retry middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Retry, listener Listener, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	if config.Attempts <= 0 {
		return nil, fmt.Errorf("incorrect (or empty) value for attempt (%d)", config.Attempts)
	}

	return &retry{
		attempts:        config.Attempts,
		initialInterval: time.Duration(config.InitialInterval),
		next:            next,
		listener:        listener,
		name:            name,
	}, nil
}

func (r *retry) GetTracingInformation() (string, ext.SpanKindEnum) {
	return r.name, tracing.SpanKindNoneEnum
}

func (r *retry) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if r.attempts == 1 {
		r.next.ServeHTTP(rw, req)
		return
	}

	closableBody := req.Body
	defer closableBody.Close()

	// if we might make multiple attempts, swap the body for an io.NopCloser
	// cf https://github.com/traefik/traefik/issues/1008
	req.Body = io.NopCloser(closableBody)

	attempts := 1

	operation := func() error {
		remainAttempts := attempts < r.attempts
		retryResponseWriter := newResponseWriter(rw)

		var shouldRetry ShouldRetry = func(shouldRetry bool) {
			retryResponseWriter.SetShouldRetry(remainAttempts && shouldRetry)
		}
		newCtx := context.WithValue(req.Context(), shouldRetryContextKey{}, shouldRetry)

		r.next.ServeHTTP(retryResponseWriter, req.Clone(newCtx))

		if !retryResponseWriter.ShouldRetry() {
			return nil
		}

		attempts++

		return fmt.Errorf("attempt %d failed", attempts-1)
	}

	backOff := backoff.WithContext(r.newBackOff(), req.Context())

	notify := func(err error, d time.Duration) {
		log.FromContext(middlewares.GetLoggerCtx(req.Context(), r.name, typeName)).
			Debugf("New attempt %d for request: %v", attempts, req.URL)

		r.listener.Retried(req, attempts)
	}

	err := backoff.RetryNotify(operation, backOff, notify)
	if err != nil {
		log.FromContext(middlewares.GetLoggerCtx(req.Context(), r.name, typeName)).
			Debugf("Final retry attempt failed: %v", err.Error())
	}
}

func (r *retry) newBackOff() backoff.BackOff {
	if r.attempts < 2 || r.initialInterval <= 0 {
		return &backoff.ZeroBackOff{}
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = r.initialInterval

	// calculate the multiplier for the given number of attempts
	// so that applying the multiplier for the given number of attempts will not exceed 2 times the initial interval
	// it allows to control the progression along the attempts
	b.Multiplier = math.Pow(2, 1/float64(r.attempts-1))

	// according to docs, b.Reset() must be called before using
	b.Reset()
	return b
}

type responseWriter interface {
	http.ResponseWriter
	http.Flusher
	ShouldRetry() bool
	SetShouldRetry(shouldRetry bool)
}

func newResponseWriter(rw http.ResponseWriter) responseWriter {
	responseWriter := &responseWriterWithoutCloseNotify{
		responseWriter: rw,
		headers:        make(http.Header),
	}
	if _, ok := rw.(http.CloseNotifier); ok {
		return &responseWriterWithCloseNotify{
			responseWriterWithoutCloseNotify: responseWriter,
		}
	}
	return responseWriter
}

type responseWriterWithoutCloseNotify struct {
	responseWriter http.ResponseWriter
	headers        http.Header
	shouldRetry    bool
	written        bool
}

func (r *responseWriterWithoutCloseNotify) ShouldRetry() bool {
	return r.shouldRetry
}

func (r *responseWriterWithoutCloseNotify) SetShouldRetry(shouldRetry bool) {
	r.shouldRetry = shouldRetry
}

func (r *responseWriterWithoutCloseNotify) Header() http.Header {
	if r.written {
		return r.responseWriter.Header()
	}
	return r.headers
}

func (r *responseWriterWithoutCloseNotify) Write(buf []byte) (int, error) {
	if r.shouldRetry {
		return len(buf), nil
	}
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	return r.responseWriter.Write(buf)
}

func (r *responseWriterWithoutCloseNotify) WriteHeader(code int) {
	if r.shouldRetry || r.written {
		return
	}

	// In that case retry case is set to false which means we at least managed
	// to write headers to the backend : we are not going to perform any further retry.
	// So it is now safe to alter current response headers with headers collected during
	// the latest try before writing headers to client.
	headers := r.responseWriter.Header()
	for header, value := range r.headers {
		headers[header] = value
	}

	r.responseWriter.WriteHeader(code)

	// Handling informational headers.
	// This allows to keep writing to r.headers map until a final status code is written.
	if code >= 100 && code <= 199 {
		return
	}

	r.written = true
}

func (r *responseWriterWithoutCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.responseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.responseWriter)
	}
	return hijacker.Hijack()
}

func (r *responseWriterWithoutCloseNotify) Flush() {
	if flusher, ok := r.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type responseWriterWithCloseNotify struct {
	*responseWriterWithoutCloseNotify
}

func (r *responseWriterWithCloseNotify) CloseNotify() <-chan bool {
	return r.responseWriter.(http.CloseNotifier).CloseNotify()
}
