package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/containous/traefik/metrics"
	gokitmetrics "github.com/go-kit/kit/metrics"
)

// EntryPointMetricsMiddleware is a Negroni compatible Handler that collects
// request metrics on an entry point.
type EntryPointMetricsMiddleware struct {
	registry       metrics.Registry
	entryPointName string
}

// NewEntryPointMetricsMiddleware creates a new EntryPointMetricsMiddleware.
func NewEntryPointMetricsMiddleware(registry metrics.Registry, entryPointName string) *EntryPointMetricsMiddleware {
	return &EntryPointMetricsMiddleware{
		registry:       registry,
		entryPointName: entryPointName,
	}
}

func (m *EntryPointMetricsMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	recorder := serveHTTP(rw, r, next)

	labels := []string{"code", strconv.Itoa(recorder.statusCode), "method", r.Method, "entrypoint", m.entryPointName}
	m.registry.EntrypointReqsCounter().With(labels...).Add(1)
	m.registry.EntrypointReqDurationHistogram().With(labels...).Observe(float64(time.Since(start).Seconds()))
}

// BackendMetricsMiddleware is a Negroni compatible Handler that collects
// request metrics on an entry point.
type BackendMetricsMiddleware struct {
	registry    metrics.Registry
	backendName string
}

// NewBackendMetricsMiddleware creates a new BackendMetricsMiddleware.
func NewBackendMetricsMiddleware(registry metrics.Registry, backendName string) *BackendMetricsMiddleware {
	return &BackendMetricsMiddleware{
		registry:    registry,
		backendName: backendName,
	}
}

func (m *BackendMetricsMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	recorder := serveHTTP(rw, r, next)

	labels := []string{"code", strconv.Itoa(recorder.statusCode), "method", r.Method, "backend", m.backendName}
	m.registry.BackendReqsCounter().With(labels...).Add(1)
	m.registry.BackendReqDurationHistogram().With(labels...).Observe(float64(time.Since(start).Seconds()))
}

func serveHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) *responseRecorder {
	prw := &responseRecorder{rw, http.StatusOK}
	next(prw, r)
	return prw
}

type retryMetrics interface {
	BackendRetriesCounter() gokitmetrics.Counter
}

// NewMetricsRetryListener instantiates a MetricsRetryListener with the given retryMetrics.
func NewMetricsRetryListener(retryMetrics retryMetrics, backendName string) RetryListener {
	return &MetricsRetryListener{retryMetrics: retryMetrics, backendName: backendName}
}

// MetricsRetryListener is an implementation of the RetryListener interface to
// record RequestMetrics about retry attempts.
type MetricsRetryListener struct {
	retryMetrics retryMetrics
	backendName  string
}

// Retried tracks the retry in the RequestMetrics implementation.
func (m *MetricsRetryListener) Retried(attempt int) {
	m.retryMetrics.BackendRetriesCounter().With("backend", m.backendName).Add(1)
}
