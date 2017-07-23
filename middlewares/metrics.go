package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/multi"
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

// MultiMetrics is a struct that provides a wrapper container for multiple Metrics, if they are configured
type MultiMetrics struct {
	wrappedMetrics       *[]Metrics
	reqsCounter          metrics.Counter
	reqDurationHistogram metrics.Histogram
	retryCounter         metrics.Counter
}

// NewMultiMetrics creates a new instance of MultiMetrics
func NewMultiMetrics(manyMetrics []Metrics) *MultiMetrics {
	counters := []metrics.Counter{}
	histograms := []metrics.Histogram{}
	retryCounters := []metrics.Counter{}

	for _, m := range manyMetrics {
		counters = append(counters, m.getReqsCounter())
		histograms = append(histograms, m.getReqDurationHistogram())
		retryCounters = append(retryCounters, m.getRetryCounter())
	}

	var mm MultiMetrics

	mm.wrappedMetrics = &manyMetrics
	mm.reqsCounter = multi.NewCounter(counters...)
	mm.reqDurationHistogram = multi.NewHistogram(histograms...)
	mm.retryCounter = multi.NewCounter(retryCounters...)

	return &mm
}

func (mm *MultiMetrics) getReqsCounter() metrics.Counter {
	return mm.reqsCounter
}

func (mm *MultiMetrics) getReqDurationHistogram() metrics.Histogram {
	return mm.reqDurationHistogram
}

func (mm *MultiMetrics) getRetryCounter() metrics.Counter {
	return mm.retryCounter
}

func (mm *MultiMetrics) getWrappedMetrics() *[]Metrics {
	return mm.wrappedMetrics
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

	reqLabels := []string{"code", strconv.Itoa(prw.statusCode), "method", r.Method}
	m.Impl.getReqsCounter().With(reqLabels...).Add(1)

	reqDurationLabels := []string{"code", strconv.Itoa(prw.statusCode)}
	m.Impl.getReqDurationHistogram().With(reqDurationLabels...).Observe(float64(time.Since(start).Seconds()))
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
