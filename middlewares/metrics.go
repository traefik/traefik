package middlewares

import (
	"github.com/go-kit/kit/metrics"
	"net/http"
	"strconv"
	"time"
)

// Metrics is an Interface that must be satisfied by any system that
// wants to expose and monitor metrics
type Metrics interface {
	getReqsCounter() metrics.Counter
	getLatencyHistogram() metrics.Histogram
	handler() http.Handler
}

// MetricsWrapper is a Negroni compatible Handler which relies on a
// given Metrics implementation to expose and monitor Traefik metrics
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
	labels := []string{"code", strconv.Itoa(prw.StatusCode()), "method", r.Method}
	m.Impl.getReqsCounter().With(labels...).Add(1)
	m.Impl.getLatencyHistogram().With(labels...).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
}

func (rw *responseRecorder) StatusCode() int {
	return rw.statusCode
}

// Handler is the chance for the Metrics implementation
// to expose its metrics on a server endpoint
func (m *MetricsWrapper) Handler() http.Handler {
	return m.Impl.handler()
}
