package metrics

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	traefikmetrics "github.com/traefik/traefik/v3/pkg/observability/metrics"
	"google.golang.org/grpc/codes"
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
		trailer  bool
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
		{
			desc:     "unannounced trailer",
			status:   "0",
			trailer:  true,
			expected: codes.OK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rw := httptest.NewRecorder()
			if test.trailer {
				rw.Header().Set(http.TrailerPrefix+"Grpc-Status", test.status)
			} else {
				rw.Header().Set("Grpc-Status", test.status)
			}

			code := grpcStatusCode(rw)

			assert.EqualValues(t, test.expected, code)
		})
	}
}

func TestMetricsMiddleware_grpcUnannouncedTrailer(t *testing.T) {
	backend := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/grpc")
		rw.WriteHeader(http.StatusOK)
		rw.Header().Set(http.TrailerPrefix+"Grpc-Status", "0")
	}))
	backend.Config.Protocols = new(http.Protocols)
	backend.Config.Protocols.SetUnencryptedHTTP2(true)
	backend.Start()
	t.Cleanup(backend.Close)

	backendURL, err := url.Parse(backend.URL)
	require.NoError(t, err)

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.Transport = &http.Transport{Protocols: protocols}

	reqsCounter := &CollectingCounterWithHeaders{}
	registry := &collectingRegistry{Registry: traefikmetrics.NewVoidRegistry(), reqsCounter: reqsCounter}

	handler, err := capture.Wrap(NewServiceMiddleware(t.Context(), proxy, registry, "service"))
	require.NoError(t, err)
	handler = observability.WithObservabilityHandler(handler, observability.Observability{MetricsEnabled: true})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/grpc")
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	codeIndex := slices.Index(reqsCounter.LastLabelValues, "code")
	require.NotEqual(t, -1, codeIndex)
	assert.Equal(t, "0", reqsCounter.LastLabelValues[codeIndex+1])
}

// CollectingCounterWithHeaders is a traefikmetrics.CounterWithHeaders implementation that enables access to the LastLabelValues.
type CollectingCounterWithHeaders struct {
	LastLabelValues []string
}

// With is there to satisfy the traefikmetrics.CounterWithHeaders interface.
func (c *CollectingCounterWithHeaders) With(_ http.Header, labelValues ...string) traefikmetrics.CounterWithHeaders {
	c.LastLabelValues = labelValues
	return c
}

// Add is there to satisfy the traefikmetrics.CounterWithHeaders interface.
func (c *CollectingCounterWithHeaders) Add(_ float64) {}

type collectingRegistry struct {
	traefikmetrics.Registry

	reqsCounter traefikmetrics.CounterWithHeaders
}

func (r *collectingRegistry) ServiceReqsCounter() traefikmetrics.CounterWithHeaders {
	return r.reqsCounter
}
