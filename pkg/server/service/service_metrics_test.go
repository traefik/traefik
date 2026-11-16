package service

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"github.com/traefik/traefik/v3/pkg/observability/tracing"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type recordingServiceCounter struct {
	counts map[string]float64
	labels []string
}

func (c *recordingServiceCounter) Add(delta float64) {
	c.counts[strings.Join(c.labels, "\x00")] += delta
}

func (c *recordingServiceCounter) With(_ http.Header, labels ...string) metrics.CounterWithHeaders {
	return &recordingServiceCounter{counts: c.counts, labels: labels}
}

type recordingServiceRegistry struct {
	metrics.Registry

	counter *recordingServiceCounter
}

func (r *recordingServiceRegistry) IsSvcEnabled() bool {
	return true
}

func (r *recordingServiceRegistry) ServiceReqsCounter() metrics.CounterWithHeaders {
	return r.counter
}

func TestBuildHTTPServiceTracingSpans(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	testCases := []struct {
		desc     string
		servers  []dynamic.Server
		expected []string
	}{
		{
			desc:     "available server",
			servers:  []dynamic.Server{{URL: backend.URL}},
			expected: []string{"ReverseProxy", "Metrics", "Service", "Root"},
		},
		{
			desc:     "no available server",
			servers:  []dynamic.Server{},
			expected: []string{"Metrics", "Service", "Root"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			registry := &recordingServiceRegistry{
				Registry: metrics.NewVoidRegistry(),
				counter:  &recordingServiceCounter{counts: make(map[string]float64)},
			}
			manager := NewManager(map[string]*runtime.ServiceInfo{
				"test@file": {
					Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: test.servers}},
				},
			}, middleware.NewObservabilityMgr(static.Configuration{}, registry, nil, nil, nil, nil), nil, &transportManagerMock{}, httputil.NewProxyBuilder(&transportManagerMock{}, nil))

			handler, err := manager.BuildHTTP(t.Context(), "test@file")
			require.NoError(t, err)

			handler, err = capture.Wrap(handler)
			require.NoError(t, err)
			handler = observability.WithObservabilityHandler(handler, observability.Observability{
				MetricsEnabled:         true,
				TracingEnabled:         true,
				DetailedTracingEnabled: true,
			})

			spanRecorder := tracetest.NewSpanRecorder()
			tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
			tracer := tracing.NewTracer(tracerProvider.Tracer("github.com/traefik/traefik"), nil, nil, nil)

			ctx, rootSpan := tracer.Start(t.Context(), "Root")
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com/", nil).WithContext(ctx))
			rootSpan.End()

			var names []string
			for _, span := range spanRecorder.Ended() {
				names = append(names, span.Name())
			}
			assert.Equal(t, test.expected, names)
		})
	}
}

func TestLoadBalancerServiceMetrics(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	registry := &recordingServiceRegistry{
		Registry: metrics.NewVoidRegistry(),
		counter:  &recordingServiceCounter{counts: make(map[string]float64)},
	}
	manager := NewManager(map[string]*runtime.ServiceInfo{
		"unavailable@file": {
			Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{}},
		},
		"available@file": {
			Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{
				Servers: []dynamic.Server{{URL: backend.URL}},
			}},
		},
	}, middleware.NewObservabilityMgr(static.Configuration{}, registry, nil, nil, nil, nil), nil, &transportManagerMock{}, httputil.NewProxyBuilder(&transportManagerMock{}, nil))

	for _, test := range []struct {
		name string
		code int
	}{
		{name: "unavailable@file", code: http.StatusServiceUnavailable},
		{name: "available@file", code: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler, err := manager.BuildHTTP(t.Context(), test.name)
			require.NoError(t, err)

			handler, err = capture.Wrap(handler)
			require.NoError(t, err)
			handler = observability.WithObservabilityHandler(handler, observability.Observability{MetricsEnabled: true})

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "http://example.com/", nil))
			assert.Equal(t, test.code, rw.Code)

			labels := []string{"service", test.name, "method", http.MethodGet, "protocol", "http", "code", strconv.Itoa(test.code)}
			assert.InDelta(t, 1, registry.counter.counts[strings.Join(labels, "\x00")], 0)
		})
	}
}
