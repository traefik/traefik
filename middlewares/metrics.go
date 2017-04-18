package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/metrics"
)

// Metrics is an Interface that must be satisfied by any system that
// wants to expose and monitor Metrics.
type Metrics interface {
	getReqsCounter() metrics.Counter
	getReqDurationHistogram() metrics.Histogram
	RetryMetrics
}

// RetryMetrics must be satisfied by any system that wants to collect and
// expose retry specific Metrics.
type RetryMetrics interface {
	getRetryCounter() metrics.Counter
}

// MetricsWrapper is a Negroni compatible Handler which relies on a
// given Metrics implementation to expose and monitor Traefik Metrics.
type MetricsWrapper struct {
	Impl Metrics
}

// NewMetricsWrapper return a MetricsWrapper struct with
// a given Metrics implementation e.g Prometheuss
func NewMetricsWrapper(impl Metrics) *MetricsWrapper {
	var metricsWrapper = MetricsWrapper{
		Impl: impl,
	}

	return &metricsWrapper
}

func (m *MetricsWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	prw := &responseRecorder{rw, http.StatusOK}
	next(prw, r)
	labels := []string{"code", strconv.Itoa(prw.statusCode), "method", r.Method}
	m.Impl.getReqsCounter().With(labels...).Add(1)
	m.Impl.getReqDurationHistogram().Observe(float64(time.Since(start).Seconds()))
}

// MetricsRetryListener is an implementation of the RetryListener interface to
// record Metrics about retry attempts.
type MetricsRetryListener struct {
	retryMetrics RetryMetrics
}

// Retried tracks the retry in the Metrics implementation.
func (m *MetricsRetryListener) Retried(attempt int) {
	if m.retryMetrics != nil {
		m.retryMetrics.getRetryCounter().Add(1)
	}
}

// NewMetricsRetryListener instantiates a MetricsRetryListener with the given RetryMetrics.
func NewMetricsRetryListener(retryMetrics RetryMetrics) RetryListener {
	return &MetricsRetryListener{retryMetrics: retryMetrics}
}
