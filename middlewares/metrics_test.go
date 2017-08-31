package middlewares

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-kit/kit/metrics"
)

func TestMetricsRetryListener(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	retryMetrics := newCollectingRetryMetrics()
	retryListener := NewMetricsRetryListener(retryMetrics, "backendName")
	retryListener.Retried(req, 1)
	retryListener.Retried(req, 2)

	wantCounterValue := float64(2)
	if retryMetrics.retryCounter.counterValue != wantCounterValue {
		t.Errorf("got counter value of %d, want %d", retryMetrics.retryCounter.counterValue, wantCounterValue)
	}

	wantLabelValues := []string{"backend", "backendName"}
	if !reflect.DeepEqual(retryMetrics.retryCounter.lastLabelValues, wantLabelValues) {
		t.Errorf("wrong label values %v used, want %v", retryMetrics.retryCounter.lastLabelValues, wantLabelValues)
	}
}

// collectingRetryMetrics is an implementation of the retryMetrics interface that can be used inside tests to collect the times Add() was called.
type collectingRetryMetrics struct {
	retryCounter *collectingCounter
}

func newCollectingRetryMetrics() collectingRetryMetrics {
	return collectingRetryMetrics{retryCounter: &collectingCounter{}}
}

func (metrics collectingRetryMetrics) RetriesCounter() metrics.Counter {
	return metrics.retryCounter
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
