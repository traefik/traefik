package metrics

import (
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "traefik"
)

// TODO: #156 - More detailed metrics
type entrypointMetrics struct {
}

// TODO: #156 - More detailed metrics
type providerMetrics struct {
}

// TODO: #156 - More detailed metrics
type frontendMetrics struct {
}

// TODO: #156 - More detailed metrics
type backendMetrics struct {
}

type globalMetrics struct {
	UpTime                    prometheus.Gauge
	RequestResponseTimeTotal  prometheus.Gauge
	RequestResponseTimeAvg    prometheus.Gauge
	RequestStatusCountCurrent map[string]prometheus.Gauge
	RequestStatusCountTotal   map[string]prometheus.Gauge

	Entrypoints entrypointMetrics
	Providers   providerMetrics
	Frontends   frontendMetrics
	Backends    backendMetrics
}

// Exporter holds exporter configuration and metric collectors
type Exporter struct {
	mutex   sync.RWMutex
	metrics globalMetrics
}

//func newCounter(metricName string, docString string, constLabels prometheus.Labels) prometheus.Counter {
//	return prometheus.NewCounter(
//		prometheus.CounterOpts{
//			Namespace:   namespace,
//			Name:        metricName,
//			Help:        docString,
//			ConstLabels: constLabels,
//		},
//	)
//}

func newGauge(metricName string, docString string, constLabels prometheus.Labels) prometheus.Gauge {
	return prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        metricName,
			Help:        docString,
			ConstLabels: constLabels,
		},
	)
}

//func newHistogram(metricName string, docString string, buckets []float64, constLabels prometheus.Labels) prometheus.Gauge {
//	return prometheus.NewHistogram(
//		prometheus.HistogramOpts{
//			Namespace:   namespace,
//			Name:        metricName,
//			Help:        docString,
//			ConstLabels: constLabels,
//			Buckets:     buckets,
//		},
//	)
//}

// Please use summary carefully. It can slow-down you code !
// The merge algorithm seems to be slow at write and fast at scrape time.
// https://github.com/prometheus/client_golang/blob/master/prometheus/summary.go#L142
// Do not create too many objectives and think about the margin of error (value of the map). Both move performances down.
//func newSummary(metricName string, docString string, Objectives map[float64]float64, constLabels prometheus.Labels) prometheus.Gauge {
//	return prometheus.NewSummary(
//		prometheus.SummaryOpts{
//			Namespace:   namespace,
//			Name:        metricName,
//			Help:        docString,
//			ConstLabels: constLabels,
//			Objectives:  map[float64]float64,
//		},
//	)
//}

// Describe describes all the metrics ever exported by the Traefik exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.metrics.UpTime.Desc()
	ch <- e.metrics.RequestResponseTimeTotal.Desc()
	ch <- e.metrics.RequestResponseTimeAvg.Desc()

	for _, metric := range e.metrics.RequestStatusCountCurrent {
		ch <- metric.Desc()
	}
	for _, metric := range e.metrics.RequestStatusCountTotal {
		ch <- metric.Desc()
	}
}

// Collect delivers the Traefik stats to Prometheus
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	e.refresh()

	ch <- e.metrics.UpTime
	ch <- e.metrics.RequestResponseTimeTotal
	ch <- e.metrics.RequestResponseTimeAvg

	for _, metric := range e.metrics.RequestStatusCountCurrent {
		ch <- metric
	}
	for _, metric := range e.metrics.RequestStatusCountTotal {
		ch <- metric
	}
}

func (e *Exporter) refresh() {
	var data = Metrics.Data()

	e.metrics.UpTime.Set(data.UpTimeSec)
	e.metrics.RequestResponseTimeTotal.Set(data.TotalResponseTimeSec)
	e.metrics.RequestResponseTimeAvg.Set(data.AverageResponseTimeSec)

	// Current request count, labeled by statusCode
	// Must be reset for missing status code metrics in data
	for _, metric := range e.metrics.RequestStatusCountCurrent {
		metric.Set(0)
	}
	for statusCode, nbr := range data.StatusCodeCount {
		if _, ok := e.metrics.RequestStatusCountCurrent[statusCode]; ok == false {
			e.metrics.RequestStatusCountCurrent[statusCode] = newGauge("request_count_current", "Number of request handled by Traefik", prometheus.Labels{"statusCode": statusCode})
		}
		e.metrics.RequestStatusCountCurrent[statusCode].Set(float64(nbr))
	}

	// Total request count, labeled by statusCode
	for statusCode, nbr := range data.TotalStatusCodeCount {
		if _, ok := e.metrics.RequestStatusCountTotal[statusCode]; ok == false {
			e.metrics.RequestStatusCountTotal[statusCode] = newGauge("request_count_total", "Number of request handled by Traefik", prometheus.Labels{"statusCode": statusCode})
		}
		e.metrics.RequestStatusCountTotal[statusCode].Set(float64(nbr))
	}
}

var exporter *Exporter

func init() {
	exporter = &Exporter{
		metrics: globalMetrics{
			UpTime: newGauge("uptime", "Current Traefik uptime", nil),
			RequestResponseTimeTotal:  newGauge("request_response_time_total", "Total response time of Traefik requests", nil),
			RequestResponseTimeAvg:    newGauge("request_response_time_avg", "Average response time of Traefik requests", nil),
			RequestStatusCountCurrent: map[string]prometheus.Gauge{},
			RequestStatusCountTotal:   map[string]prometheus.Gauge{},
		},
	}
}

// GetPrometheusHandler expose Metrics data in the Prometheus format on /metrics
func GetPrometheusHandler() http.Handler {
	prometheus.MustRegister(exporter)
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	return prometheus.UninstrumentedHandler()
}
