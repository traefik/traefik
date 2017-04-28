package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/containous/traefik/middlewares/common"
	"github.com/go-kit/kit/metrics"
)

// Metrics is an interface that must be satisfied by any system that
// wants to expose and monitor metrics.
type Metrics interface {
	getReqsCounter() metrics.Counter
	getLatencyHistogram() metrics.Histogram
	handler() http.Handler
}

// MetricsWrapper is a http.Handler which relies on a
// given Metrics implementation to expose and monitor Traefik metrics
type MetricsWrapper struct {
	common.BasicMiddleware
	Impl Metrics
}

var _ common.Middleware = &MetricsWrapper{}

// NewMetricsWrapper return a MetricsWrapper struct with
// a given Metrics implementation e.g Prometheuss
func NewMetricsWrapper(impl Metrics, next http.Handler) common.Middleware {
	return &MetricsWrapper{common.NewMiddleware(next), impl}
}

func (m *MetricsWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now().UTC()
	prw := &responseRecorder{rw, http.StatusOK}

	m.Next().ServeHTTP(prw, r)

	labels := []string{"code", strconv.Itoa(prw.StatusCode()), "method", r.Method}
	m.Impl.getReqsCounter().With(labels...).Add(1)
	m.Impl.getLatencyHistogram().Observe(float64(time.Now().UTC().Sub(start).Seconds()))
}

func (rw *responseRecorder) StatusCode() int {
	return rw.statusCode
}

// Handler is the chance for the Metrics implementation
// to expose its metrics on a server endpoint
func (m *MetricsWrapper) Handler() http.Handler {
	return m.Impl.handler()
}
