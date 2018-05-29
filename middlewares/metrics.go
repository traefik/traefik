package middlewares

import (
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/metrics"
	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/urfave/negroni"
)

const (
	protoHTTP      = "http"
	protoSSE       = "sse"
	protoWebsocket = "websocket"
)

// NewEntryPointMetricsMiddleware creates a new metrics middleware for an Entrypoint.
func NewEntryPointMetricsMiddleware(registry metrics.Registry, entryPointName string) negroni.Handler {
	return &metricsMiddleware{
		reqsCounter:          registry.EntrypointReqsCounter(),
		reqDurationHistogram: registry.EntrypointReqDurationHistogram(),
		openConnsGauge:       registry.EntrypointOpenConnsGauge(),
		baseLabels:           []string{"entrypoint", entryPointName},
	}
}

// NewBackendMetricsMiddleware creates a new metrics middleware for a Backend.
func NewBackendMetricsMiddleware(registry metrics.Registry, backendName string) negroni.Handler {
	return &metricsMiddleware{
		reqsCounter:          registry.BackendReqsCounter(),
		reqDurationHistogram: registry.BackendReqDurationHistogram(),
		openConnsGauge:       registry.BackendOpenConnsGauge(),
		baseLabels:           []string{"backend", backendName},
	}
}

type metricsMiddleware struct {
	// Important: Since this int64 field is using sync/atomic, it has to be at the top of the struct due to a bug on 32-bit platform
	// See: https://golang.org/pkg/sync/atomic/ for more information
	openConns            int64
	reqsCounter          gokitmetrics.Counter
	reqDurationHistogram gokitmetrics.Histogram
	openConnsGauge       gokitmetrics.Gauge
	baseLabels           []string
}

func (m *metricsMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	labels := []string{"method", getMethod(r), "protocol", getRequestProtocol(r)}
	labels = append(labels, m.baseLabels...)

	openConns := atomic.AddInt64(&m.openConns, 1)
	m.openConnsGauge.With(labels...).Set(float64(openConns))
	defer func(labelValues []string) {
		openConns := atomic.AddInt64(&m.openConns, -1)
		m.openConnsGauge.With(labelValues...).Set(float64(openConns))
	}(labels)

	start := time.Now()
	recorder := &responseRecorder{rw, http.StatusOK}
	next(recorder, r)

	labels = append(labels, "code", strconv.Itoa(recorder.statusCode))
	m.reqsCounter.With(labels...).Add(1)
	m.reqDurationHistogram.With(labels...).Observe(time.Since(start).Seconds())
}

func getRequestProtocol(req *http.Request) string {
	switch {
	case isWebsocketRequest(req):
		return protoWebsocket
	case isSSERequest(req):
		return protoSSE
	default:
		return protoHTTP
	}
}

// isWebsocketRequest determines if the specified HTTP request is a websocket handshake request.
func isWebsocketRequest(req *http.Request) bool {
	return containsHeader(req, "Connection", "upgrade") && containsHeader(req, "Upgrade", "websocket")
}

// isSSERequest determines if the specified HTTP request is a request for an event subscription.
func isSSERequest(req *http.Request) bool {
	return containsHeader(req, "Accept", "text/event-stream")
}

func containsHeader(req *http.Request, name, value string) bool {
	items := strings.Split(req.Header.Get(name), ",")
	for _, item := range items {
		if value == strings.ToLower(strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

func getMethod(r *http.Request) string {
	if !utf8.ValidString(r.Method) {
		log.Warnf("Invalid HTTP method encoding: %s", r.Method)
		return "NON_UTF8_HTTP_METHOD"
	}
	return r.Method
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
func (m *MetricsRetryListener) Retried(req *http.Request, attempt int) {
	m.retryMetrics.BackendRetriesCounter().With("backend", m.backendName).Add(1)
}
