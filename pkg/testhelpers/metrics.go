package testhelpers

import (
	"slices"
	"sync"

	"github.com/go-kit/kit/metrics"
)

// CollectingCounter is a metrics.Counter implementation that enables access to the counter value and last label values.
// It is safe for concurrent use, as the instrumented code usually runs in its own goroutine.
type CollectingCounter struct {
	mu              sync.RWMutex
	counterValue    float64
	lastLabelValues []string
}

// With is there to satisfy the metrics.Counter interface.
func (c *CollectingCounter) With(labelValues ...string) metrics.Counter {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastLabelValues = labelValues

	return c
}

// Add is there to satisfy the metrics.Counter interface.
func (c *CollectingCounter) Add(delta float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counterValue += delta
}

// CounterValue returns the accumulated counter value.
func (c *CollectingCounter) CounterValue() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.counterValue
}

// LastLabelValues returns a copy of the label values of the last With call.
func (c *CollectingCounter) LastLabelValues() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return slices.Clone(c.lastLabelValues)
}

// CollectingGauge is a metrics.Gauge implementation that enables access to the gauge value and last label values.
// It is safe for concurrent use, as the instrumented code usually runs in its own goroutine.
type CollectingGauge struct {
	mu              sync.RWMutex
	gaugeValue      float64
	lastLabelValues []string
}

// With is there to satisfy the metrics.Gauge interface.
func (g *CollectingGauge) With(labelValues ...string) metrics.Gauge {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.lastLabelValues = labelValues

	return g
}

// Set is there to satisfy the metrics.Gauge interface.
func (g *CollectingGauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.gaugeValue = value
}

// Add is there to satisfy the metrics.Gauge interface.
func (g *CollectingGauge) Add(delta float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.gaugeValue = delta
}

// GaugeValue returns the current gauge value.
func (g *CollectingGauge) GaugeValue() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.gaugeValue
}

// LastLabelValues returns a copy of the label values of the last With call.
func (g *CollectingGauge) LastLabelValues() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return slices.Clone(g.lastLabelValues)
}

// CollectingHealthCheckMetrics can be used for testing the Metrics instrumentation of the HealthCheck package.
type CollectingHealthCheckMetrics struct {
	Gauge *CollectingGauge
}

// NewCollectingHealthCheckMetrics creates a new CollectingHealthCheckMetrics instance.
func NewCollectingHealthCheckMetrics() *CollectingHealthCheckMetrics {
	return &CollectingHealthCheckMetrics{&CollectingGauge{}}
}

// BackendServerUpGauge is there to satisfy the healthcheck.metricsRegistry interface.
func (m *CollectingHealthCheckMetrics) BackendServerUpGauge() metrics.Gauge {
	return m.Gauge
}
