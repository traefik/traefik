package middlewares

import (
	"testing"

	"github.com/go-kit/kit/metrics"
)

func TestMetricsRetryListener(t *testing.T) {
	// nil implementation, nothing should fail
	retryListener := NewMetricsRetryListener(nil)
	retryListener.Retried(1)

	retryMetrics := newCollectingMetrics()
	retryListener = NewMetricsRetryListener(retryMetrics)
	retryListener.Retried(1)
	retryListener.Retried(2)

	wantCounterValue := float64(2)
	if retryMetrics.retryCounter.counterValue != wantCounterValue {
		t.Errorf("got counter value of %d, want %d", retryMetrics.retryCounter.counterValue, wantCounterValue)
	}
}

// collectingRetryMetrics is an implementation of the RetryMetrics interface that can be used inside tests to collect the times Add() was called.
type collectingRetryMetrics struct {
	retryCounter *collectingCounter
}

func newCollectingMetrics() collectingRetryMetrics {
	return collectingRetryMetrics{retryCounter: &collectingCounter{}}
}

func (metrics collectingRetryMetrics) getRetryCounter() metrics.Counter {
	return metrics.retryCounter
}

type collectingCounter struct {
	counterValue float64
}

func (c *collectingCounter) With(labelValues ...string) metrics.Counter {
	panic("collectingCounter.With not implemented!")
}

func (c *collectingCounter) Add(delta float64) {
	c.counterValue += delta
}
