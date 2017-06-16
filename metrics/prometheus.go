package metrics

import (
	"github.com/containous/traefik/types"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	metricNamePrefix = "traefik_"

	// server meta information
	configReloadsTotalName        = metricNamePrefix + "config_reloads_total"
	configReloadFailuresTotalName = metricNamePrefix + "config_reload_failures_total"
	configLastReloadSuccessName   = metricNamePrefix + "config_last_reload_success"

	// entrypoint
	entrypointReqsTotalName   = metricNamePrefix + "entrypoint_requests_total"
	entrypointReqDurationName = metricNamePrefix + "entrypoint_request_duration_seconds"

	// backend level
	backendReqsTotalName    = metricNamePrefix + "backend_requests_total"
	backendReqDurationName  = metricNamePrefix + "backend_request_duration_seconds"
	backendRetriesTotalName = metricNamePrefix + "backend_retries_total"
)

// RegisterPrometheusMetrics registers all Prometheus metrics.
// It must be called only once and failing to register the metrics will lead to a panic.
func RegisterPrometheusMetrics(config *types.Prometheus) Registry {
	buckets := []float64{0.1, 0.3, 1.2, 5.0}
	if config.Buckets != nil {
		buckets = config.Buckets
	}

	configReloads := newPrometheusCounter(stdprometheus.CounterOpts{
		Name: configReloadsTotalName,
		Help: "Config reloads",
	})
	configReloadFailures := newPrometheusCounter(stdprometheus.CounterOpts{
		Name: configReloadFailuresTotalName,
		Help: "Config reload failures",
	})
	lastConfigReloadSuccess := newPrometheusGauge(stdprometheus.GaugeOpts{
		Name: configLastReloadSuccessName,
		Help: "Last config reload success",
	})

	entrypointReqs := prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Name: entrypointReqsTotalName,
		Help: "How many HTTP requests processed on an entrypoint, partitioned by status code and method.",
	}, []string{"code", "method", "entrypoint"})
	entrypointReqDurations := prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Name:    entrypointReqDurationName,
		Help:    "How long it took to process the request on an entrypoint, partitioned by status code.",
		Buckets: buckets,
	}, []string{"code", "method", "entrypoint"})

	backendReqs := prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Name: backendReqsTotalName,
		Help: "How many HTTP requests processed on a backend, partitioned by status code and method.",
	}, []string{"code", "method", "backend"})
	backendReqDurations := prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Name:    backendReqDurationName,
		Help:    "How long it took to process the request on a backend, partitioned by status code.",
		Buckets: buckets,
	}, []string{"code", "method", "backend"})
	backendRetries := prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Name: backendRetriesTotalName,
		Help: "How many request retries happened on a backend.",
	}, []string{"backend"})

	return &prometheusRegistry{
		configReloadsCounter:           configReloads,
		configReloadFailuresCounter:    configReloadFailures,
		lastConfigReloadSuccessGauge:   lastConfigReloadSuccess,
		entrypointReqsCounter:          entrypointReqs,
		entrypointReqDurationHistogram: entrypointReqDurations,
		backendReqsCounter:             backendReqs,
		backendReqDurationHistogram:    backendReqDurations,
		backendRetriesCounter:          backendRetries,
	}
}

type prometheusRegistry struct {
	configReloadsCounter           metrics.Counter
	configReloadFailuresCounter    metrics.Counter
	lastConfigReloadSuccessGauge   metrics.Gauge
	entrypointReqsCounter          metrics.Counter
	entrypointReqDurationHistogram metrics.Histogram
	backendReqsCounter             metrics.Counter
	backendReqDurationHistogram    metrics.Histogram
	backendRetriesCounter          metrics.Counter
}

func (r *prometheusRegistry) ConfigReloadsCounter() metrics.Counter {
	return r.configReloadsCounter
}

func (r *prometheusRegistry) ConfigReloadFailuresCounter() metrics.Counter {
	return r.configReloadFailuresCounter
}

func (r *prometheusRegistry) LastConfigReloadSuccessGauge() metrics.Gauge {
	return r.lastConfigReloadSuccessGauge
}

func (r *prometheusRegistry) EntrypointReqsCounter() metrics.Counter {
	return r.entrypointReqsCounter
}

func (r *prometheusRegistry) EntrypointReqDurationHistogram() metrics.Histogram {
	return r.entrypointReqDurationHistogram
}

func (r *prometheusRegistry) BackendReqsCounter() metrics.Counter {
	return r.backendReqsCounter
}

func (r *prometheusRegistry) BackendReqDurationHistogram() metrics.Histogram {
	return r.backendReqDurationHistogram
}

func (r *prometheusRegistry) BackendRetriesCounter() metrics.Counter {
	return r.backendRetriesCounter
}

func newPrometheusCounter(opts stdprometheus.CounterOpts) metrics.Counter {
	counter := stdprometheus.NewCounter(opts)
	stdprometheus.MustRegister(counter)
	return &prometheusCounter{promCounter: counter}
}

// prometheusCounter creates a metrics.Counter implementation that holds only a Prometheus Counter and not a CounterVec.
// It should be used when a Counter with no labels is required.
type prometheusCounter struct {
	promCounter stdprometheus.Counter
}

func (c *prometheusCounter) With(labelValues ...string) metrics.Counter {
	return c
}

func (c *prometheusCounter) Add(delta float64) {
	c.promCounter.Add(delta)
}

func newPrometheusGauge(opts stdprometheus.GaugeOpts) metrics.Gauge {
	gauge := stdprometheus.NewGauge(opts)
	stdprometheus.MustRegister(gauge)
	return &prometheusGauge{promGauge: gauge}
}

// prometheusGauge creates a metrics.Gauge implementation that holds only a Prometheus Gauge and not a GaugeVec.
// It should be used when a Gauge with no labels is required.
type prometheusGauge struct {
	promGauge stdprometheus.Gauge
}

func (g *prometheusGauge) With(labelValues ...string) metrics.Gauge {
	return g
}

func (g *prometheusGauge) Set(value float64) {
	g.promGauge.Set(value)
}
