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
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

// Compile time validation that the response writer implements http interfaces correctly.
var _ middlewares.Stateful = &responseWriter{}

const typeName = "Retry"

// Listener is used to inform about retry attempts.
type Listener interface {
	// Retried will be called when a retry happens, with the request attempt passed to it.
	// For the first retry this will be attempt 2.
	Retried(req *http.Request, attempt int)
}

// Listeners is a convenience type to construct a list of Listener and notify
// each of them about a retry attempt.
type Listeners []Listener

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
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

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
		shouldRetry := attempts < r.attempts
		retryResponseWriter := newResponseWriter(rw, shouldRetry)

		// Disable retries when the backend already received request data
		trace := &httptrace.ClientTrace{
			WroteHeaders: func() {
				retryResponseWriter.DisableRetries()
			},
			WroteRequest: func(httptrace.WroteRequestInfo) {
				retryResponseWriter.DisableRetries()
			},
		}
		newCtx := httptrace.WithClientTrace(req.Context(), trace)

		r.next.ServeHTTP(retryResponseWriter, req.WithContext(newCtx))

		if !retryResponseWriter.ShouldRetry() {
			return nil
		}

		attempts++

		return fmt.Errorf("attempt %d failed", attempts-1)
	}

	logger := middlewares.GetLogger(req.Context(), r.name, typeName)

	backOff := backoff.WithContext(r.newBackOff(), req.Context())

	notify := func(err error, d time.Duration) {
		logger.Debug().Msgf("New attempt %d for request: %v", attempts, req.URL)

		r.listener.Retried(req, attempts)
	}

	err := backoff.RetryNotify(operation, backOff, notify)
	if err != nil {
		logger.Debug().Err(err).Msg("Final retry attempt failed")
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

// Retried exists to implement the Listener interface. It calls Retried on each of its slice entries.
func (l Listeners) Retried(req *http.Request, attempt int) {
	for _, listener := range l {
		listener.Retried(req, attempt)
	}
}

func newResponseWriter(rw http.ResponseWriter, shouldRetry bool) *responseWriter {
	return &responseWriter{
		responseWriter: rw,
		headers:        make(http.Header),
		shouldRetry:    shouldRetry,
	}
}

type responseWriter struct {
	responseWriter http.ResponseWriter
	headers        http.Header
	shouldRetry    bool
	written        bool
}

func (r *responseWriter) ShouldRetry() bool {
	return r.shouldRetry
}

func (r *responseWriter) DisableRetries() {
	r.shouldRetry = false
}

func (r *responseWriter) Header() http.Header {
	if r.written {
		return r.responseWriter.Header()
	}
	return r.headers
}

func (r *responseWriter) Write(buf []byte) (int, error) {
	if r.ShouldRetry() {
		return len(buf), nil
	}
	return r.responseWriter.Write(buf)
}

func (r *responseWriter) WriteHeader(code int) {
	if r.ShouldRetry() && code == http.StatusServiceUnavailable {
		// We get a 503 HTTP Status Code when there is no backend server in the pool
		// to which the request could be sent.  Also, note that r.ShouldRetry()
		// will never return true in case there was a connection established to
		// the backend server and so we can be sure that the 503 was produced
		// inside Traefik already and we don't have to retry in this cases.
		r.DisableRetries()
	}

	if r.ShouldRetry() {
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
	r.written = true
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.responseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.responseWriter)
	}
	return hijacker.Hijack()
}

func (r *responseWriter) Flush() {
	if flusher, ok := r.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
