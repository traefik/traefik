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
	reqsName    = "traefik_requests_total"
	latencyName = "traefik_request_duration_seconds"
)

// Prometheus is an Implementation for Metrics that exposes prometheus metrics for the latency
// and the number of requests partitioned by status code and method.
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

	cv := stdprometheus.NewCounterVec(
		stdprometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code and method.",
			ConstLabels: stdprometheus.Labels{"service": name},
		},
		[]string{"code", "method"},
	)

	err := stdprometheus.Register(cv)
	if err != nil {
		e, ok := err.(stdprometheus.AlreadyRegisteredError)
		if !ok {
			panic(err)
		}
		m.reqsCounter = prometheus.NewCounter(e.ExistingCollector.(*stdprometheus.CounterVec))
	} else {
		m.reqsCounter = prometheus.NewCounter(cv)
	}

	var buckets []float64
	if config.Buckets != nil {
		buckets = config.Buckets
	} else {
		buckets = []float64{0.1, 0.3, 1.2, 5}
	}

	hv := stdprometheus.NewHistogramVec(
		stdprometheus.HistogramOpts{
			Name:        latencyName,
			Help:        "How long it took to process the request.",
			ConstLabels: stdprometheus.Labels{"service": name},
			Buckets:     buckets,
		},
		[]string{},
	)

	err = stdprometheus.Register(hv)
	if err != nil {
		e, ok := err.(stdprometheus.AlreadyRegisteredError)
		if !ok {
			panic(err)
		}
		m.latencyHistogram = prometheus.NewHistogram(e.ExistingCollector.(*stdprometheus.HistogramVec))
	} else {
		m.latencyHistogram = prometheus.NewHistogram(hv)
	}

	return &m
}

func (p *Prometheus) handler() http.Handler {
	return promhttp.Handler()
}
