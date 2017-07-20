package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/containous/traefik/metrics"
	gokitmetrics "github.com/go-kit/kit/metrics"
)

// MetricsWrapper is a Negroni compatible Handler which relies on a
// given Metrics implementation to expose and monitor Traefik Metrics.
type MetricsWrapper struct {
	registry    metrics.Registry
	serviceName string
}

// NewMetricsWrapper return a MetricsWrapper struct with
// a given Metrics implementation e.g Prometheuss
func NewMetricsWrapper(registry metrics.Registry, service string) *MetricsWrapper {
	var metricsWrapper = MetricsWrapper{
		registry:    registry,
		serviceName: service,
	}

	return &metricsWrapper
}

func (m *MetricsWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	prw := &responseRecorder{rw, http.StatusOK}
	next(prw, r)

	reqLabels := []string{"service", m.serviceName, "code", strconv.Itoa(prw.statusCode), "method", r.Method}
	m.registry.ReqsCounter().With(reqLabels...).Add(1)

	reqDurationLabels := []string{"service", m.serviceName, "code", strconv.Itoa(prw.statusCode)}
	m.registry.ReqDurationHistogram().With(reqDurationLabels...).Observe(float64(time.Since(start).Seconds()))
}

type retryMetrics interface {
	RetriesCounter() gokitmetrics.Counter
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
	m.retryMetrics.RetriesCounter().With("backend", m.backendName).Add(1)
}
