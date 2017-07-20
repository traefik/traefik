package metrics

import (
	"github.com/go-kit/kit/metrics"
)

// Registry has to implemented by any system that wants to monitor and expose metrics.
type Registry interface {
	// server metrics
	ConfigReloadsCounter() metrics.Counter
	ConfigReloadFailuresCounter() metrics.Counter
	LastConfigReloadSuccessGauge() metrics.Gauge

	// entry point metrics
	EntrypointReqsCounter() metrics.Counter
	EntrypointReqDurationHistogram() metrics.Histogram

	// backend metrics
	BackendReqsCounter() metrics.Counter
	BackendReqDurationHistogram() metrics.Histogram
	BackendRetriesCounter() metrics.Counter
}

// NewVoidRegistry is a noop implementation of metrics.Registry.
// It is used to avoid nil checking in components that do metric collections.
func NewVoidRegistry() Registry {
	return &voidRegistry{
		c: &voidCounter{},
		g: &voidGauge{},
		h: &voidHistogram{},
	}
}

type voidRegistry struct {
	c metrics.Counter
	g metrics.Gauge
	h metrics.Histogram
}

func (r *voidRegistry) ConfigReloadsCounter() metrics.Counter             { return r.c }
func (r *voidRegistry) ConfigReloadFailuresCounter() metrics.Counter      { return r.c }
func (r *voidRegistry) LastConfigReloadSuccessGauge() metrics.Gauge       { return r.g }
func (r *voidRegistry) EntrypointReqsCounter() metrics.Counter            { return r.c }
func (r *voidRegistry) EntrypointReqDurationHistogram() metrics.Histogram { return r.h }
func (r *voidRegistry) BackendReqsCounter() metrics.Counter               { return r.c }
func (r *voidRegistry) BackendReqDurationHistogram() metrics.Histogram    { return r.h }
func (r *voidRegistry) BackendRetriesCounter() metrics.Counter            { return r.c }

type voidCounter struct{}

func (v *voidCounter) With(labelValues ...string) metrics.Counter { return v }
func (v *voidCounter) Add(delta float64)                          {}

type voidGauge struct{}

func (g *voidGauge) With(labelValues ...string) metrics.Gauge { return g }
func (g *voidGauge) Set(value float64)                        {}

type voidHistogram struct{}

func (h *voidHistogram) With(labelValues ...string) metrics.Histogram { return h }
func (h *voidHistogram) Observe(value float64)                        {}
