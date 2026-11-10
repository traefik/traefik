package metrics

import (
	"testing"

	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

func TestRouterMetricsHandler(t *testing.T) {
	testCases := []struct {
		desc     string
		registry metrics.Registry
		wrapped  bool
	}{
		{
			desc: "no registry",
		},
		{
			desc:     "TCP routers metrics disabled",
			registry: &registryMock{gauge: &gaugeMock{}},
		},
		{
			desc:     "TCP routers metrics enabled",
			registry: &registryMock{routerEnabled: true, gauge: &gaugeMock{}},
			wrapped:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := tcp.HandlerFunc(func(conn tcp.WriteCloser) {})

			handler, err := RouterMetricsHandler(t.Context(), test.registry, "router@file", "service@file")(next)
			require.NoError(t, err)

			_, wrapped := handler.(*metricsMiddleware)
			assert.Equal(t, test.wrapped, wrapped)
		})
	}
}

func TestMetricsMiddleware_ServeTCP(t *testing.T) {
	gauge := &gaugeMock{}
	registry := &registryMock{routerEnabled: true, gauge: gauge}

	var openConnections float64
	next := tcp.HandlerFunc(func(conn tcp.WriteCloser) {
		openConnections = gauge.value
	})

	handler := NewRouterMiddleware(t.Context(), next, registry, "router@file", "service@file")

	assert.Equal(t, []string{"router", "router@file", "service", "service@file"}, gauge.labelValues)

	handler.ServeTCP(nil)

	// The connection is counted while it is served, and released afterwards.
	assert.InDelta(t, 1, openConnections, 0)
	assert.InDelta(t, 0, gauge.value, 0)
}

type registryMock struct {
	metrics.Registry

	routerEnabled bool
	gauge         *gaugeMock
}

func (r *registryMock) IsTCPRouterEnabled() bool {
	return r.routerEnabled
}

func (r *registryMock) RouterTCPOpenConnectionsGauge() gokitmetrics.Gauge {
	return r.gauge
}

type gaugeMock struct {
	value       float64
	labelValues []string
}

func (g *gaugeMock) With(labelValues ...string) gokitmetrics.Gauge {
	g.labelValues = labelValues
	return g
}

func (g *gaugeMock) Set(value float64) {
	g.value = value
}

func (g *gaugeMock) Add(delta float64) {
	g.value += delta
}
