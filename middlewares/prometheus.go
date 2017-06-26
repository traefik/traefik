package middlewares

import (
	"fmt"

	"github.com/containous/traefik/types"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	reqsTotalName    = "traefik_requests_total"
	reqDurationName  = "traefik_request_duration_seconds"
	retriesTotalName = "traefik_backend_retries_total"
)

// Prometheus is an Implementation for Metrics that exposes the following Prometheus metrics:
// - number of requests partitioned by status code and method
// - request durations partitioned by status code
// - amount of retries happened
type Prometheus struct {
	reqsCounter          metrics.Counter
	reqDurationHistogram metrics.Histogram
	retryCounter         metrics.Counter
}

func (p *Prometheus) getReqsCounter() metrics.Counter {
	return p.reqsCounter
}

func (p *Prometheus) getReqDurationHistogram() metrics.Histogram {
	return p.reqDurationHistogram
}

func (p *Prometheus) getRetryCounter() metrics.Counter {
	return p.retryCounter
}

// NewPrometheus returns a new Prometheus Metrics implementation.
// With the returned collectors you have the possibility to clean up the internal Prometheus state by unsubscribing the collectors.
// This is for example useful while testing the Prometheus implementation.
// If any of the Prometheus Metrics can not be registered an error will be returned and the returned Metrics implementation will be nil.
func NewPrometheus(name string, config *types.Prometheus) (*Prometheus, []stdprometheus.Collector, error) {
	var prom Prometheus
	var collectors []stdprometheus.Collector

	cv := stdprometheus.NewCounterVec(
		stdprometheus.CounterOpts{
			Name:        reqsTotalName,
			Help:        "How many HTTP requests processed, partitioned by status code and method.",
			ConstLabels: stdprometheus.Labels{"service": name},
		},
		[]string{"code", "method"},
	)
	cv, err := registerCounterVec(cv)
	if err != nil {
		return nil, collectors, err
	}
	prom.reqsCounter = prometheus.NewCounter(cv)
	collectors = append(collectors, cv)

	var buckets []float64
	if config.Buckets != nil {
		buckets = config.Buckets
	} else {
		buckets = []float64{0.1, 0.3, 1.2, 5}
	}
	hv := stdprometheus.NewHistogramVec(
		stdprometheus.HistogramOpts{
			Name:        reqDurationName,
			Help:        "How long it took to process the request.",
			ConstLabels: stdprometheus.Labels{"service": name},
			Buckets:     buckets,
		},
		[]string{"code"},
	)
	hv, err = registerHistogramVec(hv)
	if err != nil {
		return nil, collectors, err
	}
	prom.reqDurationHistogram = prometheus.NewHistogram(hv)
	collectors = append(collectors, hv)

	cv = stdprometheus.NewCounterVec(
		stdprometheus.CounterOpts{
			Name:        retriesTotalName,
			Help:        "How many request retries happened in total.",
			ConstLabels: stdprometheus.Labels{"service": name},
		},
		[]string{},
	)
	cv, err = registerCounterVec(cv)
	if err != nil {
		return nil, collectors, err
	}
	prom.retryCounter = prometheus.NewCounter(cv)
	collectors = append(collectors, cv)

	return &prom, collectors, nil
}

func registerCounterVec(cv *stdprometheus.CounterVec) (*stdprometheus.CounterVec, error) {
	err := stdprometheus.Register(cv)

	if err != nil {
		e, ok := err.(stdprometheus.AlreadyRegisteredError)
		if !ok {
			return nil, fmt.Errorf("error registering CounterVec: %s", e)
		}
		cv = e.ExistingCollector.(*stdprometheus.CounterVec)
	}

	return cv, nil
}

func registerHistogramVec(hv *stdprometheus.HistogramVec) (*stdprometheus.HistogramVec, error) {
	err := stdprometheus.Register(hv)

	if err != nil {
		e, ok := err.(stdprometheus.AlreadyRegisteredError)
		if !ok {
			return nil, fmt.Errorf("error registering HistogramVec: %s", e)
		}
		hv = e.ExistingCollector.(*stdprometheus.HistogramVec)
	}

	return hv, nil
}
