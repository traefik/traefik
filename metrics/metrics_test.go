package metrics

import (
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewMultiRegistry(t *testing.T) {
	registries := []Registry{newCollectingRetryMetrics(), newCollectingRetryMetrics()}
	registry := NewMultiRegistry(registries)

	registry.BackendReqsCounter().With("key", "requests").Add(1)
	registry.BackendReqDurationHistogram().With("key", "durations").Observe(2)
	registry.BackendRetriesCounter().With("key", "retries").Add(3)

	for _, collectingRegistry := range registries {
		cReqsCounter := collectingRegistry.BackendReqsCounter().(*counterMock)
		cReqDurationHistogram := collectingRegistry.BackendReqDurationHistogram().(*histogramMock)
		cRetriesCounter := collectingRegistry.BackendRetriesCounter().(*counterMock)

		wantCounterValue := float64(1)
		if cReqsCounter.counterValue != wantCounterValue {
			t.Errorf("Got value %f for ReqsCounter, want %f", cReqsCounter.counterValue, wantCounterValue)
		}
		wantHistogramValue := float64(2)
		if cReqDurationHistogram.lastHistogramValue != wantHistogramValue {
			t.Errorf("Got last observation %f for ReqDurationHistogram, want %f", cReqDurationHistogram.lastHistogramValue, wantHistogramValue)
		}
		wantCounterValue = float64(3)
		if cRetriesCounter.counterValue != wantCounterValue {
			t.Errorf("Got value %f for RetriesCounter, want %f", cRetriesCounter.counterValue, wantCounterValue)
		}

		assert.Equal(t, []string{"key", "requests"}, cReqsCounter.lastLabelValues)
		assert.Equal(t, []string{"key", "durations"}, cReqDurationHistogram.lastLabelValues)
		assert.Equal(t, []string{"key", "retries"}, cRetriesCounter.lastLabelValues)
	}
}

func newCollectingRetryMetrics() Registry {
	return &standardRegistry{
		backendReqsCounter:          &counterMock{},
		backendReqDurationHistogram: &histogramMock{},
		backendRetriesCounter:       &counterMock{},
	}
}

type counterMock struct {
	counterValue    float64
	lastLabelValues []string
}

func (c *counterMock) With(labelValues ...string) metrics.Counter {
	c.lastLabelValues = labelValues
	return c
}

func (c *counterMock) Add(delta float64) {
	c.counterValue += delta
}

type histogramMock struct {
	lastHistogramValue float64
	lastLabelValues    []string
}

func (c *histogramMock) With(labelValues ...string) metrics.Histogram {
	c.lastLabelValues = labelValues
	return c
}

func (c *histogramMock) Observe(value float64) {
	c.lastHistogramValue = value
}
