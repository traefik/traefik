package metrics

import (
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewVoidRegistry(t *testing.T) {
	registry := NewVoidRegistry()

	registry.ReqsCounter().With("some", "value").Add(1)
	registry.ReqDurationHistogram().With("some", "value").Observe(1)
	registry.RetriesCounter().With("some", "value").Add(1)
}

func TestNewMutliRegistry(t *testing.T) {
	registries := []Registry{newCollectingRetryMetrics(), newCollectingRetryMetrics()}
	registry := NewMultiRegistry(registries)

	registry.ReqsCounter().With("key", "requests").Add(1)
	registry.ReqDurationHistogram().With("key", "durations").Observe(2)
	registry.RetriesCounter().With("key", "retries").Add(3)

	for _, collectingRegistry := range registries {
		cReqsCounter := collectingRegistry.ReqsCounter().(*collectingCounter)
		cReqDurationHistogram := collectingRegistry.ReqDurationHistogram().(*collectingHistogram)
		cRetriesCounter := collectingRegistry.RetriesCounter().(*collectingCounter)

		if cReqsCounter.counterValue != 1 {
			t.Errorf("Got value %v for ReqsCounter, want %v", cReqsCounter.counterValue, 1)
		}
		if cReqDurationHistogram.lastHistogramValue != 2 {
			t.Errorf("Got last observation %v for ReqDurationHistogram, want %v", cReqDurationHistogram.lastHistogramValue, 2)
		}
		if cRetriesCounter.counterValue != 3 {
			t.Errorf("Got value %v for RetriesCounter, want %v", cRetriesCounter.counterValue, 3)
		}

		assert.Equal(t, []string{"key", "requests"}, cReqsCounter.lastLabelValues)
		assert.Equal(t, []string{"key", "durations"}, cReqDurationHistogram.lastLabelValues)
		assert.Equal(t, []string{"key", "retries"}, cRetriesCounter.lastLabelValues)
	}
}

// collectingRegistry is an implementation of themetrics.Registry interface that can be used inside tests.
type collectingRegistry struct {
	reqsCounter          *collectingCounter
	reqDurationHistogram *collectingHistogram
	retriesCounter       *collectingCounter
}

func newCollectingRetryMetrics() *collectingRegistry {
	return &collectingRegistry{
		reqsCounter:          &collectingCounter{},
		reqDurationHistogram: &collectingHistogram{},
		retriesCounter:       &collectingCounter{},
	}
}

func (r *collectingRegistry) ReqsCounter() metrics.Counter {
	return r.reqsCounter
}

func (r *collectingRegistry) ReqDurationHistogram() metrics.Histogram {
	return r.reqDurationHistogram
}

func (r *collectingRegistry) RetriesCounter() metrics.Counter {
	return r.retriesCounter
}

type collectingCounter struct {
	counterValue    float64
	lastLabelValues []string
}

func (c *collectingCounter) With(labelValues ...string) metrics.Counter {
	c.lastLabelValues = labelValues
	return c
}

func (c *collectingCounter) Add(delta float64) {
	c.counterValue += delta
}

type collectingHistogram struct {
	lastHistogramValue float64
	lastLabelValues    []string
}

func (c *collectingHistogram) With(labelValues ...string) metrics.Histogram {
	c.lastLabelValues = labelValues
	return c
}

func (c *collectingHistogram) Observe(value float64) {
	c.lastHistogramValue = value
}
