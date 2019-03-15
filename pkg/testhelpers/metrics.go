package testhelpers

import "github.com/go-kit/kit/metrics"

// CollectingCounter is a metrics.Counter implementation that enables access to the CounterValue and LastLabelValues.
type CollectingCounter struct {
	CounterValue    float64
	LastLabelValues []string
}

// With is there to satisfy the metrics.Counter interface.
func (c *CollectingCounter) With(labelValues ...string) metrics.Counter {
	c.LastLabelValues = labelValues
	return c
}

// Add is there to satisfy the metrics.Counter interface.
func (c *CollectingCounter) Add(delta float64) {
	c.CounterValue += delta
}

// CollectingGauge is a metrics.Gauge implementation that enables access to the GaugeValue and LastLabelValues.
type CollectingGauge struct {
	GaugeValue      float64
	LastLabelValues []string
}

// With is there to satisfy the metrics.Gauge interface.
func (g *CollectingGauge) With(labelValues ...string) metrics.Gauge {
	g.LastLabelValues = labelValues
	return g
}

// Set is there to satisfy the metrics.Gauge interface.
func (g *CollectingGauge) Set(value float64) {
	g.GaugeValue = value
}

// Add is there to satisfy the metrics.Gauge interface.
func (g *CollectingGauge) Add(delta float64) {
	g.GaugeValue = delta
}

// CollectingHealthCheckMetrics can be used for testing the Metrics instrumentation of the HealthCheck package.
type CollectingHealthCheckMetrics struct {
	Gauge *CollectingGauge
}

// BackendServerUpGauge is there to satisfy the healthcheck.metricsRegistry interface.
func (m *CollectingHealthCheckMetrics) BackendServerUpGauge() metrics.Gauge {
	return m.Gauge
}

// NewCollectingHealthCheckMetrics creates a new CollectingHealthCheckMetrics instance.
func NewCollectingHealthCheckMetrics() *CollectingHealthCheckMetrics {
	return &CollectingHealthCheckMetrics{&CollectingGauge{}}
}
