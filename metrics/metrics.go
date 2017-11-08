package metrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/multi"
)

// Registry has to implemented by any system that wants to monitor and expose metrics.
type Registry interface {
	// IsEnabled shows whether metrics instrumentation is enabled.
	IsEnabled() bool
	ReqsCounter() metrics.Counter
	ReqDurationHistogram() metrics.Histogram
	RetriesCounter() metrics.Counter
	PluginDurationHistogram() metrics.Histogram
}

// NewMultiRegistry creates a new standardRegistry that wraps multiple Registries.
func NewMultiRegistry(registries []Registry) Registry {
	reqsCounters := []metrics.Counter{}
	reqDurationHistograms := []metrics.Histogram{}
	retriesCounters := []metrics.Counter{}
	pluginDurationHistograms := []metrics.Histogram{}

	for _, r := range registries {
		reqsCounters = append(reqsCounters, r.ReqsCounter())
		reqDurationHistograms = append(reqDurationHistograms, r.ReqDurationHistogram())
		retriesCounters = append(retriesCounters, r.RetriesCounter())
		pluginDurationHistograms = append(pluginDurationHistograms, r.PluginDurationHistogram())
	}

	return &standardRegistry{
		enabled:                 true,
		reqsCounter:             multi.NewCounter(reqsCounters...),
		reqDurationHistogram:    multi.NewHistogram(reqDurationHistograms...),
		retriesCounter:          multi.NewCounter(retriesCounters...),
		pluginDurationHistogram: multi.NewHistogram(pluginDurationHistograms...),
	}
}

type standardRegistry struct {
	enabled                 bool
	reqsCounter             metrics.Counter
	reqDurationHistogram    metrics.Histogram
	retriesCounter          metrics.Counter
	pluginDurationHistogram metrics.Histogram
}

func (r *standardRegistry) IsEnabled() bool {
	return r.enabled
}

func (r *standardRegistry) ReqsCounter() metrics.Counter {
	return r.reqsCounter
}

func (r *standardRegistry) ReqDurationHistogram() metrics.Histogram {
	return r.reqDurationHistogram
}

func (r *standardRegistry) RetriesCounter() metrics.Counter {
	return r.retriesCounter
}

func (r *standardRegistry) PluginDurationHistogram() metrics.Histogram {
	return r.pluginDurationHistogram
}

// NewVoidRegistry is a noop implementation of metrics.Registry.
// It is used to avoid nil checking in components that do metric collections.
func NewVoidRegistry() Registry {
	return &standardRegistry{
		enabled:                 false,
		reqsCounter:             &voidCounter{},
		reqDurationHistogram:    &voidHistogram{},
		retriesCounter:          &voidCounter{},
		pluginDurationHistogram: &voidHistogram{},
	}
}

type voidCounter struct{}

func (v *voidCounter) With(labelValues ...string) metrics.Counter { return v }
func (v *voidCounter) Add(delta float64)                          {}

type voidHistogram struct{}

func (h *voidHistogram) With(labelValues ...string) metrics.Histogram { return h }
func (h *voidHistogram) Observe(value float64)                        {}
