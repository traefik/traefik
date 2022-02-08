package metrics

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
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

type rwWithCloseNotify struct {
	*httptest.ResponseRecorder
}

func (r *rwWithCloseNotify) CloseNotify() <-chan bool {
	panic("implement me")
}

func TestCloseNotifier(t *testing.T) {
	testCases := []struct {
		rw                      http.ResponseWriter
		desc                    string
		implementsCloseNotifier bool
	}{
		{
			rw:                      httptest.NewRecorder(),
			desc:                    "does not implement CloseNotifier",
			implementsCloseNotifier: false,
		},
		{
			rw:                      &rwWithCloseNotify{httptest.NewRecorder()},
			desc:                    "implements CloseNotifier",
			implementsCloseNotifier: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, ok := test.rw.(http.CloseNotifier)
			assert.Equal(t, test.implementsCloseNotifier, ok)

			rw := newResponseRecorder(test.rw)
			_, impl := rw.(http.CloseNotifier)
			assert.Equal(t, test.implementsCloseNotifier, impl)
		})
	}
}

func Test_getMethod(t *testing.T) {
	testCases := []struct {
		method   string
		expected string
	}{
		{
			method:   http.MethodGet,
			expected: http.MethodGet,
		},
		{
			method:   strings.ToLower(http.MethodGet),
			expected: "EXTENSION_METHOD",
		},
		{
			method:   "THIS_IS_NOT_A_VALID_METHOD",
			expected: "EXTENSION_METHOD",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.method, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(test.method, "http://example.com", nil)
			assert.Equal(t, test.expected, getMethod(request))
		})
	}
}
