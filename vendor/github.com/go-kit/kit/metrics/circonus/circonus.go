// Package circonus provides a Circonus backend for metrics.
package circonus

import (
	"github.com/circonus-labs/circonus-gometrics"

	"github.com/go-kit/kit/metrics"
)

// Circonus wraps a CirconusMetrics object and provides constructors for each of
// the Go kit metrics. The CirconusMetrics object manages aggregation of
// observations and emission to the Circonus server.
type Circonus struct {
	m *circonusgometrics.CirconusMetrics
}

// New creates a new Circonus object wrapping the passed CirconusMetrics, which
// the caller should create and set in motion. The Circonus object can be used
// to construct individual Go kit metrics.
func New(m *circonusgometrics.CirconusMetrics) *Circonus {
	return &Circonus{
		m: m,
	}
}

// NewCounter returns a counter metric with the given name.
func (c *Circonus) NewCounter(name string) *Counter {
	return &Counter{
		name: name,
		m:    c.m,
	}
}

// NewGauge returns a gauge metric with the given name.
func (c *Circonus) NewGauge(name string) *Gauge {
	return &Gauge{
		name: name,
		m:    c.m,
	}
}

// NewHistogram returns a histogram metric with the given name.
func (c *Circonus) NewHistogram(name string) *Histogram {
	return &Histogram{
		h: c.m.NewHistogram(name),
	}
}

// Counter is a Circonus implementation of a counter metric.
type Counter struct {
	name string
	m    *circonusgometrics.CirconusMetrics
}

// With implements Counter, but is a no-op, because Circonus metrics have no
// concept of per-observation label values.
func (c *Counter) With(labelValues ...string) metrics.Counter { return c }

// Add implements Counter. Delta is converted to uint64; precision will be lost.
func (c *Counter) Add(delta float64) { c.m.Add(c.name, uint64(delta)) }

// Gauge is a Circonus implementation of a gauge metric.
type Gauge struct {
	name string
	m    *circonusgometrics.CirconusMetrics
}

// With implements Gauge, but is a no-op, because Circonus metrics have no
// concept of per-observation label values.
func (g *Gauge) With(labelValues ...string) metrics.Gauge { return g }

// Set implements Gauge.
func (g *Gauge) Set(value float64) { g.m.SetGauge(g.name, value) }

// Histogram is a Circonus implementation of a histogram metric.
type Histogram struct {
	h *circonusgometrics.Histogram
}

// With implements Histogram, but is a no-op, because Circonus metrics have no
// concept of per-observation label values.
func (h *Histogram) With(labelValues ...string) metrics.Histogram { return h }

// Observe implements Histogram. No precision is lost.
func (h *Histogram) Observe(value float64) { h.h.RecordValue(value) }
