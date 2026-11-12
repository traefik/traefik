package metrics

import (
	"crypto/tls"
	"io"
	"maps"
	"net"
	"slices"
	"strings"
	"testing"

	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/traefik/v3/pkg/tls/generate"
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
			registry: newRegistryMock(false),
		},
		{
			desc:     "TCP routers metrics enabled",
			registry: newRegistryMock(true),
			wrapped:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := tcp.HandlerFunc(func(conn tcp.WriteCloser) {})

			handler, err := RouterMetricsHandler(t.Context(), test.registry, "router@file", "service@file", nil)(next)
			require.NoError(t, err)

			_, wrapped := handler.(*metricsMiddleware)
			assert.Equal(t, test.wrapped, wrapped)
		})
	}
}

func TestMetricsMiddleware_ServeTCP(t *testing.T) {
	registry := newRegistryMock(true)

	var openConnections map[string]float64
	next := tcp.HandlerFunc(func(conn tcp.WriteCloser) {
		openConnections = maps.Clone(registry.gauge.values)
	})

	handler := NewRouterMiddleware(t.Context(), next, registry, "router@file", "service@file", &dynamic.RouterTCPTLSConfig{Passthrough: true})

	handler.ServeTCP(connMock{localAddr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}})

	labels := "traefik.router.name=router@file,traefik.service.name=service@file,traefik.tls.mode=passthrough,network.type=ipv4,server.port=8080"

	// The connection is counted while it is served, and released afterwards.
	assert.Equal(t, map[string]float64{labels: 1}, openConnections)
	assert.Equal(t, map[string]float64{labels: 0}, registry.gauge.values)
	assert.Empty(t, registry.tlsGauge.values)
}

func TestMetricsMiddleware_ServeTCP_tls(t *testing.T) {
	cert, err := generate.DefaultCertificate()
	require.NoError(t, err)

	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() {
		_ = serverConn.Close()
		_ = clientConn.Close()
	})

	go func() {
		client := tls.Client(clientConn, &tls.Config{InsecureSkipVerify: true})
		_, _ = client.Write([]byte("ping"))
	}()

	registry := newRegistryMock(true)

	var beforeHandshake, afterHandshake map[string]float64
	next := tcp.HandlerFunc(func(conn tcp.WriteCloser) {
		beforeHandshake = maps.Clone(registry.tlsGauge.values)

		_, err := io.ReadFull(conn, make([]byte, 4))
		assert.NoError(t, err)

		afterHandshake = maps.Clone(registry.tlsGauge.values)
	})

	handler := NewRouterMiddleware(t.Context(), next, registry, "router@file", "service@file", &dynamic.RouterTCPTLSConfig{})

	handler.ServeTCP(tls.Server(serverConn, &tls.Config{
		Certificates: []tls.Certificate{*cert},
		MaxVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
	}))

	// The network labels are empty because the local address of a pipe is not a TCP one.
	labels := "traefik.router.name=router@file,traefik.service.name=service@file,network.type=,server.port=,tls.protocol.version=1.2,tls.cipher=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"

	assert.Contains(t, registry.gauge.values, "traefik.router.name=router@file,traefik.service.name=service@file,traefik.tls.mode=terminated,network.type=,server.port=")

	// The TLS connection is counted once the handshake is done, and released afterwards.
	assert.Empty(t, beforeHandshake)
	assert.Equal(t, map[string]float64{labels: 1}, afterHandshake)
	assert.Equal(t, map[string]float64{labels: 0}, registry.tlsGauge.values)
}

type connMock struct {
	tcp.WriteCloser

	localAddr net.Addr
}

func (c connMock) LocalAddr() net.Addr {
	return c.localAddr
}

type registryMock struct {
	metrics.Registry

	routerEnabled bool
	gauge         *gaugeMock
	tlsGauge      *gaugeMock
}

func newRegistryMock(routerEnabled bool) *registryMock {
	return &registryMock{
		routerEnabled: routerEnabled,
		gauge:         &gaugeMock{values: make(map[string]float64)},
		tlsGauge:      &gaugeMock{values: make(map[string]float64)},
	}
}

func (r *registryMock) IsTCPRouterEnabled() bool {
	return r.routerEnabled
}

func (r *registryMock) RouterTCPOpenConnectionsGauge() gokitmetrics.Gauge {
	return r.gauge
}

func (r *registryMock) RouterTCPOpenTLSConnectionsGauge() gokitmetrics.Gauge {
	return r.tlsGauge
}

// gaugeMock holds the values of a gauge by labels.
type gaugeMock struct {
	values map[string]float64
	labels []string
}

func (g *gaugeMock) With(labelValues ...string) gokitmetrics.Gauge {
	labels := slices.Clone(g.labels)
	for i := 0; i < len(labelValues); i += 2 {
		labels = append(labels, labelValues[i]+"="+labelValues[i+1])
	}

	return &gaugeMock{values: g.values, labels: labels}
}

func (g *gaugeMock) Set(value float64) {
	g.values[strings.Join(g.labels, ",")] = value
}

func (g *gaugeMock) Add(delta float64) {
	g.values[strings.Join(g.labels, ",")] += delta
}
