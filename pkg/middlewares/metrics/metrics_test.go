package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	metricsregistry "github.com/traefik/traefik/v3/pkg/observability/metrics"
	"google.golang.org/grpc/codes"
)

func TestMetricsMiddleware_ResponseAborted(t *testing.T) {
	for _, abort := range []bool{false, true} {
		name := "completed"
		if abort {
			name = "aborted"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			requests := &CollectingCounter{}
			requestBytes := &CollectingCounter{}
			responseBytes := &CollectingCounter{}
			duration := &collectingHistogram{}
			middleware := &metricsMiddleware{
				next: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					_, err := io.Copy(io.Discard, req.Body)
					require.NoError(t, err)
					rw.WriteHeader(http.StatusOK)
					_, err = io.WriteString(rw, "chunk")
					require.NoError(t, err)
					if abort {
						panic(http.ErrAbortHandler)
					}
				}),
				reqsCounter:          metricsregistry.NewCounterWithNoopHeaders(requests),
				reqDurationHistogram: duration,
				reqsBytesCounter:     requestBytes,
				respsBytesCounter:    responseBytes,
				baseLabels:           []string{"service", "test"},
			}
			handler, err := capture.Wrap(middleware)
			require.NoError(t, err)
			ctx := observability.WithObservability(t.Context(), observability.Observability{MetricsEnabled: true})
			req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader("request"))
			rw := httptest.NewRecorder()

			if abort {
				assert.PanicsWithValue(t, http.ErrAbortHandler, func() { handler.ServeHTTP(rw, req) })
			} else {
				assert.NotPanics(t, func() { handler.ServeHTTP(rw, req) })
			}

			labels := []string{"service", "test", "method", "POST", "protocol", "http", "code", "200"}
			assert.InDelta(t, 1, requests.CounterValue, 0.001)
			assert.Equal(t, labels, requests.LastLabelValues)
			assert.InDelta(t, 7, requestBytes.CounterValue, 0.001)
			assert.Equal(t, labels, requestBytes.LastLabelValues)
			assert.InDelta(t, 5, responseBytes.CounterValue, 0.001)
			assert.Equal(t, labels, responseBytes.LastLabelValues)
			assert.Equal(t, 1, duration.observations)
			assert.Equal(t, labels, duration.labels)
			assert.Equal(t, http.StatusOK, rw.Code)
			assert.Equal(t, "chunk", rw.Body.String())
		})
	}
}

type collectingHistogram struct {
	observations int
	labels       []string
}

func (h *collectingHistogram) With(labels ...string) metricsregistry.ScalableHistogram {
	h.labels = labels
	return h
}

func (h *collectingHistogram) Observe(float64) {
	h.observations++
}

func (h *collectingHistogram) ObserveFromStart(start time.Time) {
	h.Observe(time.Since(start).Seconds())
}

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
		t.Run(test.method, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(test.method, "http://example.com", nil)
			assert.Equal(t, test.expected, getMethod(request))
		})
	}
}

func Test_getRequestProtocol(t *testing.T) {
	testCases := []struct {
		desc     string
		headers  http.Header
		expected string
	}{
		{
			desc:     "default",
			expected: protoHTTP,
		},
		{
			desc: "websocket",
			headers: http.Header{
				"Connection": []string{"upgrade"},
				"Upgrade":    []string{"websocket"},
			},
			expected: protoWebsocket,
		},
		{
			desc: "SSE",
			headers: http.Header{
				"Accept": []string{"text/event-stream"},
			},
			expected: protoSSE,
		},
		{
			desc: "grpc web",
			headers: http.Header{
				"Content-Type": []string{"application/grpc-web-text"},
			},
			expected: protoGRPCWeb,
		},
		{
			desc: "grpc",
			headers: http.Header{
				"Content-Type": []string{"application/grpc-text"},
			},
			expected: protoGRPC,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "https://localhost", http.NoBody)
			req.Header = test.headers

			protocol := getRequestProtocol(req)

			assert.Equal(t, test.expected, protocol)
		})
	}
}

func Test_grpcStatusCode(t *testing.T) {
	testCases := []struct {
		desc     string
		status   string
		expected codes.Code
	}{
		{
			desc:     "invalid number",
			status:   "foo",
			expected: codes.Unknown,
		},
		{
			desc:     "number",
			status:   "1",
			expected: codes.Canceled,
		},
		{
			desc:     "invalid string",
			status:   `"foo"`,
			expected: codes.Unknown,
		},
		{
			desc:     "string",
			status:   `"OK"`,
			expected: codes.OK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rw := httptest.NewRecorder()
			rw.Header().Set("Grpc-Status", test.status)

			code := grpcStatusCode(rw)

			assert.EqualValues(t, test.expected, code)
		})
	}
}
