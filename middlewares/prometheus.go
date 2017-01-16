package middlewares

import (
	"github.com/containous/traefik/types"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const (
	reqsName    = "requests_total"
	latencyName = "request_duration_milliseconds"
)

// Prometheus is an Implementation for Metrics that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code and method.
type Prometheus struct {
	reqsCounter      metrics.Counter
	latencyHistogram metrics.Histogram
}

func (p *Prometheus) getReqsCounter() metrics.Counter {
	return p.reqsCounter
}

func (p *Prometheus) getLatencyHistogram() metrics.Histogram {
	return p.latencyHistogram
}

// NewPrometheus returns a new prometheus Metrics implementation.
func NewPrometheus(name string, config *types.Prometheus) *Prometheus {
	var m Prometheus
	m.reqsCounter = prometheus.NewCounterFrom(
		stdprometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code and method.",
			ConstLabels: stdprometheus.Labels{"service": name},
		},
		[]string{"code", "method"},
	)

	var buckets []float64
	if config.Buckets != nil {
		buckets = config.Buckets
	} else {
		buckets = []float64{100, 300, 1200, 5000}
	}

	m.latencyHistogram = prometheus.NewHistogramFrom(
		stdprometheus.HistogramOpts{
			Name:        latencyName,
			Help:        "How long it took to process the request, partitioned by status code and method.",
			ConstLabels: stdprometheus.Labels{"service": name},
			Buckets:     buckets,
		},
		[]string{"code", "method"},
	)
	return &m
}

func (p *Prometheus) handler() http.Handler {
	return promhttp.Handler()
}
