package middlewares

import (
	"reflect"
	"testing"

	"github.com/go-kit/kit/metrics"
)

func TestMetricsRetryListener(t *testing.T) {
	retryMetrics := newCollectingMetrics()
	retryListener := NewMetricsRetryListener(retryMetrics, "backendName")
	retryListener.Retried(1)
	retryListener.Retried(2)

	wantCounterValue := float64(2)
	if retryMetrics.retriesCounter.counterValue != wantCounterValue {
		t.Errorf("got counter value of %d, want %d", retryMetrics.retriesCounter.counterValue, wantCounterValue)
	}

	wantLabelValues := []string{"backend", "backendName"}
	if !reflect.DeepEqual(retryMetrics.retriesCounter.lastLabelValues, wantLabelValues) {
		t.Errorf("wrong label values %v used, want %v", retryMetrics.retriesCounter.lastLabelValues, wantLabelValues)
	}
}

// collectingRetryMetrics is an implementation of the retryMetrics interface that can be used inside tests to collect the times Add() was called.
type collectingRetryMetrics struct {
	retriesCounter *collectingCounter
}

func newCollectingMetrics() collectingRetryMetrics {
	return collectingRetryMetrics{retriesCounter: &collectingCounter{}}
}

func (metrics collectingRetryMetrics) BackendRetriesCounter() metrics.Counter {
	return metrics.retriesCounter
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
