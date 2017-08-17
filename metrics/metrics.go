package metrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/multi"
)

// Registry has to implemented by any system that wants to monitor and expose metrics.
type Registry interface {
	ReqsCounter() metrics.Counter
	ReqDurationHistogram() metrics.Histogram
	RetriesCounter() metrics.Counter
}

// NewMultiRegistry creates a new multiRegistry that wraps multiple Registries.
func NewMultiRegistry(registries []Registry) Registry {
	reqsCounters := []metrics.Counter{}
	reqDurationHistograms := []metrics.Histogram{}
	retriesCounters := []metrics.Counter{}

	for _, r := range registries {
		reqsCounters = append(reqsCounters, r.ReqsCounter())
		reqDurationHistograms = append(reqDurationHistograms, r.ReqDurationHistogram())
		retriesCounters = append(retriesCounters, r.RetriesCounter())
	}

	return &multiRegistry{
		reqsCounter:          multi.NewCounter(reqsCounters...),
		reqDurationHistogram: multi.NewHistogram(reqDurationHistograms...),
		retriesCounter:       multi.NewCounter(retriesCounters...),
	}
}

// multiRegistry is a Registry that wraps multiple Registries and calls each of them when a Metric is tracked.
type multiRegistry struct {
	reqsCounter          metrics.Counter
	reqDurationHistogram metrics.Histogram
	retriesCounter       metrics.Counter
}

func (r *multiRegistry) ReqsCounter() metrics.Counter {
	return r.reqsCounter
}

// ReqsCounter
func (r *multiRegistry) ReqDurationHistogram() metrics.Histogram {
	return r.reqDurationHistogram
}

func (r *multiRegistry) RetriesCounter() metrics.Counter {
	return r.retriesCounter
}

// NewVoidRegistry is a noop implementation of metrics.Registry.
// It is used to avoid nil checking in components that do metric collections.
func NewVoidRegistry() Registry {
	return &voidRegistry{
		c: &voidCounter{},
		h: &voidHistogram{},
	}
}

type voidRegistry struct {
	c metrics.Counter
	h metrics.Histogram
}

func (r *voidRegistry) ReqsCounter() metrics.Counter            { return r.c }
func (r *voidRegistry) ReqDurationHistogram() metrics.Histogram { return r.h }
func (r *voidRegistry) RetriesCounter() metrics.Counter         { return r.c }

type voidCounter struct{}

func (v *voidCounter) With(labelValues ...string) metrics.Counter { return v }
func (v *voidCounter) Add(delta float64)                          {}

type voidHistogram struct{}

func (h *voidHistogram) With(labelValues ...string) metrics.Histogram { return h }
func (h *voidHistogram) Observe(value float64)                        {}
