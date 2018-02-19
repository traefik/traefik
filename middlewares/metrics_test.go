package middlewares

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/go-kit/kit/metrics"
)

func TestMetricsRetryListener(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	retryMetrics := newCollectingRetryMetrics()
	retryListener := NewMetricsRetryListener(retryMetrics, "backendName")
	retryListener.Retried(req, 1)
	retryListener.Retried(req, 2)

	wantCounterValue := float64(2)
	if retryMetrics.retriesCounter.CounterValue != wantCounterValue {
		t.Errorf("got counter value of %f, want %f", retryMetrics.retriesCounter.CounterValue, wantCounterValue)
	}

	wantLabelValues := []string{"backend", "backendName"}
	if !reflect.DeepEqual(retryMetrics.retriesCounter.LastLabelValues, wantLabelValues) {
		t.Errorf("wrong label values %v used, want %v", retryMetrics.retriesCounter.LastLabelValues, wantLabelValues)
	}
}

// collectingRetryMetrics is an implementation of the retryMetrics interface that can be used inside tests to collect the times Add() was called.
type collectingRetryMetrics struct {
	retriesCounter *testhelpers.CollectingCounter
}

func newCollectingRetryMetrics() *collectingRetryMetrics {
	return &collectingRetryMetrics{retriesCounter: &testhelpers.CollectingCounter{}}
}

func (metrics *collectingRetryMetrics) BackendRetriesCounter() metrics.Counter {
	return metrics.retriesCounter
}
