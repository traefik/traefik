package retry

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/observability/tracing"
	"github.com/traefik/traefik/v3/pkg/server/service/loadbalancer/mirror"
	"github.com/traefik/traefik/v3/pkg/types"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
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

// Retried exists to implement the Listener interface. It calls Retried on each of its slice entries.
func (l Listeners) Retried(req *http.Request, attempt int) {
	for _, listener := range l {
		listener.Retried(req, attempt)
	}
}

type wroteRequestContextKey struct{}

// WroteRequest is a function to signal if the request has been written to the backend.
type WroteRequest func()

// ContextWroteRequest returns the WroteRequest function if it has been set by the Retry middleware in the chain.
func ContextWroteRequest(ctx context.Context) WroteRequest {
	f, _ := ctx.Value(wroteRequestContextKey{}).(WroteRequest)
	return f
}

type proxyReachedContextKey struct{}

// ProxyReached is a function signaling that the request reached the backend proxy.
type ProxyReached func()

// ContextProxyReached returns the ProxyReached function if it has been set by the Retry middleware in the chain.
func ContextProxyReached(ctx context.Context) ProxyReached {
	f, _ := ctx.Value(proxyReachedContextKey{}).(ProxyReached)
	return f
}

// WrapHandler wraps a given http.Handler to inject the httptrace.ClientTrace in the request context when it is needed
// by the retry middleware.
func WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Reaching this handler means the request reached the proxy layer,
		// which arms the network-error retry (a dial failure still reaches here and writes
		// a 502 without sending bytes). Responses produced earlier in the chain
		// (e.g. an auth middleware returning 401) never reach here and must not be retried.
		if proxyReached := ContextProxyReached(req.Context()); proxyReached != nil {
			proxyReached()
		}

		if wroteRequest := ContextWroteRequest(req.Context()); wroteRequest != nil {
			trace := &httptrace.ClientTrace{
				WroteHeaders: func() {
					wroteRequest()
				},
				WroteRequest: func(httptrace.WroteRequestInfo) {
					wroteRequest()
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
	attempts                   int
	statusCode                 types.HTTPCodeRanges
	maxRequestBodyBytes        int64
	disableRetryOnNetworkError bool
	initialInterval            time.Duration
	timeout                    time.Duration
	retryNonIdempotentMethod   bool

	next     http.Handler
	listener Listener
	name     string
}

// New returns a new retry middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Retry, listener Listener, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if len(config.Status) == 0 && config.DisableRetryOnNetworkError {
		return nil, errors.New("retry middleware requires at least HTTP status codes or retry on TCP")
	}

	if config.Attempts <= 0 {
		return nil, fmt.Errorf("incorrect (or empty) value for attempt (%d)", config.Attempts)
	}

	maxRequestBodyBytes := dynamic.RetryDefaultMaxRequestBodyBytes
	if config.MaxRequestBodyBytes != nil {
		maxRequestBodyBytes = *config.MaxRequestBodyBytes
	}

	retryCfg := &retry{
		attempts:                   config.Attempts,
		maxRequestBodyBytes:        maxRequestBodyBytes,
		disableRetryOnNetworkError: config.DisableRetryOnNetworkError,
		retryNonIdempotentMethod:   config.RetryNonIdempotentMethod,
		initialInterval:            time.Duration(config.InitialInterval),
		timeout:                    time.Duration(config.Timeout),
		name:                       name,
		listener:                   listener,
		next:                       next,
	}

	if len(config.Status) > 0 {
		httpCodeRanges, err := types.NewHTTPCodeRanges(config.Status)
		if err != nil {
			return nil, fmt.Errorf("creating HTTP code ranges: %w", err)
		}
		retryCfg.statusCode = httpCodeRanges
	}

	return retryCfg, nil
}

func (r *retry) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if r.attempts == 1 {
		r.next.ServeHTTP(rw, req)
		return
	}

	logger := middlewares.GetLogger(req.Context(), r.name, typeName)

	// statusCodes contains the retriable status codes,
	// it stands as a condition for HTTP retry.
	var statusCodes types.HTTPCodeRanges
	isIdempotent := req.Method != http.MethodPost && req.Method != http.MethodPatch && req.Method != "LOCK"
	if r.retryNonIdempotentMethod || isIdempotent {
		// statusCode controls whether the request is retried.
		// A nil value bypass the retry.
		statusCodes = r.statusCode
	}

	var reusableReq *mirror.ReusableRequest
	if len(statusCodes) > 0 {
		var err error
		var readBytes []byte
		reusableReq, readBytes, err = mirror.NewReusableRequest(req, r.maxRequestBodyBytes)
		if err != nil && !errors.Is(err, mirror.ErrBodyTooLarge) {
			logger.Debug().Err(err).Msg("Error while creating reusable request for retry middleware")
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// If the request body has failed to be read,
		// disable the HTTP retry.
		if errors.Is(err, mirror.ErrBodyTooLarge) {
			statusCodes = nil

			req.Body = &peekedBody{
				ReadCloser: req.Body,
				peeked:     readBytes,
			}
		}
	}

	if reusableReq == nil {
		closableBody := req.Body
		defer closableBody.Close()

		// if we might make multiple attempts, swap the body for an io.NopCloser
		// cf https://github.com/traefik/traefik/issues/1008
		req.Body = io.NopCloser(closableBody)
	}

	start := time.Now()

	attempts := 1

	initialCtx := req.Context()
	tracer := tracing.TracerFromContext(initialCtx)

	var currentSpan trace.Span
	operation := func() error {
		if tracer != nil && observability.DetailedTracingEnabled(req.Context()) {
			if currentSpan != nil {
				currentSpan.End()
			}
			// Because multiple tracing spans may need to be created,
			// the Retry middleware does not implement trace.Traceable,
			// and directly creates a new span for each retry operation.
			var tracingCtx context.Context
			tracingCtx, currentSpan = tracer.Start(initialCtx, typeName, trace.WithSpanKind(trace.SpanKindInternal))

			currentSpan.SetAttributes(attribute.String("traefik.middleware.name", r.name))
			// Only add the attribute "http.resend_count" defined by semantic conventions starting from second attempt.
			if attempts > 1 {
				currentSpan.SetAttributes(semconv.HTTPRequestResendCount(attempts - 1))
			}

			req = req.WithContext(tracingCtx)
		}

		if reusableReq != nil {
			req = reusableReq.Clone(req.Context())
		}

		if attempts == r.attempts {
			r.next.ServeHTTP(rw, req)
			return nil
		}

		retryResponseWriter := newResponseWriter(rw, statusCodes, start, r.timeout, r.disableRetryOnNetworkError)

		retryReq := req
		if !r.disableRetryOnNetworkError {
			var wroteRequest WroteRequest = func() {
				retryResponseWriter.wroteRequest.Store(true)
			}
			ctx := context.WithValue(req.Context(), wroteRequestContextKey{}, wroteRequest)

			var proxyReached ProxyReached = func() {
				retryResponseWriter.proxyReached.Store(true)
			}
			ctx = context.WithValue(ctx, proxyReachedContextKey{}, proxyReached)

			retryReq = req.Clone(ctx)
		}

		r.next.ServeHTTP(retryResponseWriter, retryReq)

		// End the operation as soon as the client received something.
		if retryResponseWriter.written {
			return nil
		}

		attempts++

		return fmt.Errorf("attempt %d failed", attempts-1)
	}

	backOff := backoff.WithContext(r.newBackOff(), req.Context())

	notify := func(err error, d time.Duration) {
		logger.Debug().Msgf("New attempt %d for request: %v", attempts, req.URL)

		r.listener.Retried(req, attempts)
	}

	err := backoff.RetryNotify(operation, backOff, notify)
	if err != nil {
		logger.Debug().Err(err).Msg("Final retry attempt failed")
	}

	if currentSpan != nil {
		currentSpan.End()
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

func newResponseWriter(rw http.ResponseWriter, statusCodeRanges types.HTTPCodeRanges, start time.Time, timeout time.Duration, disableRetryOnNetworkError bool) *responseWriter {
	return &responseWriter{
		responseWriter:             rw,
		headers:                    make(http.Header),
		statusCodeRange:            statusCodeRanges,
		start:                      start,
		timeout:                    timeout,
		disableRetryOnNetworkError: disableRetryOnNetworkError,
	}
}

type responseWriter struct {
	responseWriter http.ResponseWriter
	headers        http.Header

	disableRetryOnNetworkError bool
	statusCodeRange            types.HTTPCodeRanges
	start                      time.Time
	timeout                    time.Duration

	proxyReached   atomic.Bool
	wroteRequest   atomic.Bool
	written        bool
	shouldNotWrite bool
}

func (r *responseWriter) Header() http.Header {
	if r.shouldNotWrite || !r.written {
		return r.headers
	}

	return r.responseWriter.Header()
}

func (r *responseWriter) Write(buf []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}

	if r.shouldNotWrite {
		return len(buf), nil
	}

	return r.responseWriter.Write(buf)
}

func (r *responseWriter) WriteHeader(code int) {
	if r.shouldNotWrite || r.written {
		return
	}

	timedOut := r.timeout > 0 && time.Since(r.start) >= r.timeout
	// Retry on a network error only when the request reached the backend proxy
	// but no bytes were sent to the backend yet. Responses produced before reaching the
	// proxy (e.g. an auth middleware returning 401) must flow through to the client.
	if !timedOut && !r.disableRetryOnNetworkError && r.proxyReached.Load() && !r.wroteRequest.Load() {
		r.shouldNotWrite = true
		return
	}

	// Handle 1xx informational responses, except 101 Switching Protocols which
	// is a final response (e.g. WebSocket upgrade) and should flow through to
	// the status-code retry logic below.
	// Accumulated headers are cleared to avoid leaking headers from informational
	// responses into the final response.
	if code >= 100 && code <= 199 && code != http.StatusSwitchingProtocols {
		clear(r.headers)

		return
	}

	if r.statusCodeRange != nil {
		r.shouldNotWrite = !timedOut && r.statusCodeRange.Contains(code)
	}

	if r.shouldNotWrite {
		return
	}

	// In that case retry case is set to false which means we at least managed
	// to write headers to the backend : we are not going to perform any further retry.
	// So it is now safe to alter current response headers with headers collected during
	// the latest try before writing headers to client.
	maps.Copy(r.responseWriter.Header(), r.headers)

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

type peekedBody struct {
	io.ReadCloser

	peeked []byte
}

// Read drains accumulated peeked bytes first, then reads the body.
func (p *peekedBody) Read(b []byte) (int, error) {
	if len(p.peeked) > 0 {
		n := copy(b, p.peeked)
		p.peeked = p.peeked[n:]
		if len(p.peeked) == 0 {
			p.peeked = nil
		}
		return n, nil
	}

	return p.ReadCloser.Read(b)
}
