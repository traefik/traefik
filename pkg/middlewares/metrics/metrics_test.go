package metrics

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-kit/kit/metrics"
)

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

func TestMetricsRetryListener(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	retryMetrics := newCollectingRetryMetrics()
	retryListener := NewRetryListener(retryMetrics, "serviceName")
	retryListener.Retried(req, 1)
	retryListener.Retried(req, 2)

	wantCounterValue := float64(2)
	if retryMetrics.retriesCounter.CounterValue != wantCounterValue {
		t.Errorf("got counter value of %f, want %f", retryMetrics.retriesCounter.CounterValue, wantCounterValue)
	}

	wantLabelValues := []string{"service", "serviceName"}
	if !reflect.DeepEqual(retryMetrics.retriesCounter.LastLabelValues, wantLabelValues) {
		t.Errorf("wrong label values %v used, want %v", retryMetrics.retriesCounter.LastLabelValues, wantLabelValues)
	}
}

// collectingRetryMetrics is an implementation of the retryMetrics interface that can be used inside tests to collect the times Add() was called.
type collectingRetryMetrics struct {
	retriesCounter *CollectingCounter
}

func newCollectingRetryMetrics() *collectingRetryMetrics {
	return &collectingRetryMetrics{retriesCounter: &CollectingCounter{}}
}

func (m *collectingRetryMetrics) ServiceRetriesCounter() metrics.Counter {
	return m.retriesCounter
}
