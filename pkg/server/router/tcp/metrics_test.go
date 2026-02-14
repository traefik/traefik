package tcp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	traefikmetrics "github.com/traefik/traefik/v3/pkg/observability/metrics"
	tcpmiddleware "github.com/traefik/traefik/v3/pkg/server/middleware/tcp"
	"github.com/traefik/traefik/v3/pkg/server/service/tcp"
	tcp2 "github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

type mockRegistry struct {
	traefikmetrics.Registry
	routerOpenConnectionsGauge *mockGauge
}

func (m *mockRegistry) IsRouterEnabled() bool {
	return true
}

func (m *mockRegistry) RouterOpenConnectionsGauge() metrics.Gauge {
	return m.routerOpenConnectionsGauge
}

type mockGauge struct {
	value       float64
	labelValues []string
}

func (m *mockGauge) With(labelValues ...string) metrics.Gauge {
	m.labelValues = labelValues
	return m
}

func (m *mockGauge) Set(value float64) {
	m.value = value
}

func (m *mockGauge) Add(delta float64) {
	m.value += delta
}

func TestRouterOpenConnectionsMetrics(t *testing.T) {
	routerName := "test-router"
	serviceName := "test-service"
	entryPointName := "web"

	conf := &runtime.Configuration{
		TCPServices: map[string]*runtime.TCPServiceInfo{
			serviceName: {
				TCPService: &dynamic.TCPService{
					LoadBalancer: &dynamic.TCPServersLoadBalancer{
						Servers: []dynamic.TCPServer{
							{Address: "127.0.0.1:9999"},
						},
					},
				},
			},
		},
		TCPRouters: map[string]*runtime.TCPRouterInfo{
			routerName: {
				TCPRouter: &dynamic.TCPRouter{
					EntryPoints: []string{entryPointName},
					Service:     serviceName,
					Rule:        "HostSNI(`*`)",
				},
			},
		},
	}

	gauge := &mockGauge{}
	registry := &mockRegistry{
		routerOpenConnectionsGauge: gauge,
	}

	dialerManager := tcp2.NewDialerManager(nil)
	dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
	svcManager := tcp.NewManager(conf, dialerManager)

	middlewaresBuilder := tcpmiddleware.NewBuilder(conf.TCPMiddlewares)

	manager := NewManager(conf,
		svcManager,
		middlewaresBuilder, nil, nil,
		traefiktls.NewManager(nil),
		registry,
	)

	handlers := manager.BuildHandlers(context.Background(), []string{entryPointName})
	_, ok := handlers[entryPointName]
	require.True(t, ok)

	// Verify that the gauge's With method was called with the correct labels
	// during handler construction. The With method on our mock stores the label values.
	assert.Equal(t, []string{"router", routerName, "service", serviceName}, gauge.labelValues)
}

func TestRouterOpenConnectionsMetrics_ServeTCP(t *testing.T) {
	// Test the metricsMiddleware directly to verify increment/decrement behavior.
	gauge := &mockGauge{}

	mockHandler := &mockTCPHandler{}
	middleware := &metricsMiddleware{
		next:                 mockHandler,
		openConnectionsGauge: gauge,
	}

	conn := &mockConn{}
	middleware.ServeTCP(conn)

	// After ServeTCP completes, the gauge should have been incremented and decremented.
	assert.Equal(t, float64(0), gauge.value)
	assert.True(t, mockHandler.called, "next handler should have been called")
}

type mockTCPHandler struct {
	called bool
}

func (m *mockTCPHandler) ServeTCP(conn tcp2.WriteCloser) {
	m.called = true
}

type mockConn struct {
	tcp2.WriteCloser
}

func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) CloseWrite() error                  { return nil }
func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return 0, nil }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1234} }
func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 80} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
