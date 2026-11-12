package metrics

import (
	"context"
	"crypto/tls"
	"net"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"

	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

const (
	typeName   = "MetricsTCP"
	nameRouter = "metrics-router"
)

// The labels are named after the attributes set on the tracing spans,
// and after the OpenTelemetry semantic conventions attributes.
const (
	routerNameLabel         = "traefik.router.name"
	serviceNameLabel        = "traefik.service.name"
	tlsModeLabel            = "traefik.tls.mode"
	networkTypeLabel        = "network.type"
	serverPortLabel         = "server.port"
	tlsProtocolVersionLabel = "tls.protocol.version"
	tlsCipherLabel          = "tls.cipher"
)

// The values of the TLS mode label.
const (
	tlsModeNone        = "none"
	tlsModeTerminated  = "terminated"
	tlsModePassthrough = "passthrough"
)

// The gauges are given all their labels at once, when a connection is served,
// because the Prometheus gauge does not support being given its labels in several steps.
type metricsMiddleware struct {
	next tcp.Handler

	openConnectionsGauge  gokitmetrics.Gauge
	openConnectionsLabels []string

	openTLSConnectionsGauge  gokitmetrics.Gauge
	openTLSConnectionsLabels []string
}

// NewRouterMiddleware creates a new metrics middleware for a TCP Router.
func NewRouterMiddleware(ctx context.Context, next tcp.Handler, registry metrics.Registry, routerName string, serviceName string, tlsConfig *dynamic.RouterTCPTLSConfig) tcp.Handler {
	middlewares.GetLogger(ctx, nameRouter, typeName).Debug().Msg("Creating middleware")

	return &metricsMiddleware{
		next:                     next,
		openConnectionsGauge:     registry.RouterTCPOpenConnectionsGauge(),
		openConnectionsLabels:    []string{routerNameLabel, routerName, serviceNameLabel, serviceName, tlsModeLabel, getTLSMode(tlsConfig)},
		openTLSConnectionsGauge:  registry.RouterTCPOpenTLSConnectionsGauge(),
		openTLSConnectionsLabels: []string{routerNameLabel, routerName, serviceNameLabel, serviceName},
	}
}

// RouterMetricsHandler returns the metrics TCP router handler.
func RouterMetricsHandler(ctx context.Context, registry metrics.Registry, routerName string, serviceName string, tlsConfig *dynamic.RouterTCPTLSConfig) tcp.Constructor {
	return func(next tcp.Handler) (tcp.Handler, error) {
		if registry == nil || !registry.IsTCPRouterEnabled() {
			return next, nil
		}

		return NewRouterMiddleware(ctx, next, registry, routerName, serviceName, tlsConfig), nil
	}
}

func getTLSMode(tlsConfig *dynamic.RouterTCPTLSConfig) string {
	switch {
	case tlsConfig == nil:
		return tlsModeNone
	case tlsConfig.Passthrough:
		return tlsModePassthrough
	default:
		return tlsModeTerminated
	}
}

// ServeTCP serves the given TCP connection.
func (m *metricsMiddleware) ServeTCP(conn tcp.WriteCloser) {
	networkLabels := getNetworkLabels(conn)

	openConnectionsGauge := m.openConnectionsGauge.With(slices.Concat(m.openConnectionsLabels, networkLabels)...)
	openConnectionsGauge.Add(1)
	defer openConnectionsGauge.Add(-1)

	// Only the connections whose TLS is terminated by Traefik are TLS connections here,
	// the passthrough ones are forwarded without their TLS parameters being known.
	if tlsConn, ok := conn.(*tls.Conn); ok {
		mConn := &tlsMetricsConn{
			Conn:   tlsConn,
			gauge:  m.openTLSConnectionsGauge,
			labels: slices.Concat(m.openTLSConnectionsLabels, networkLabels),
		}
		defer mConn.release()

		conn = mConn
	}

	m.next.ServeTCP(conn)
}

// getNetworkLabels returns the network.type and server.port labels of the given connection.
// The values are empty when the local address is not a TCP one,
// as the metrics exporters require a constant set of labels.
func getNetworkLabels(conn net.Conn) []string {
	addr, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		return []string{networkTypeLabel, "", serverPortLabel, ""}
	}

	networkType := "ipv6"
	if addr.IP.To4() != nil {
		networkType = "ipv4"
	}

	return []string{networkTypeLabel, networkType, serverPortLabel, strconv.Itoa(addr.Port)}
}

// tlsMetricsConn counts a TLS connection as open once its handshake is done.
// The handshake is not forced, as it has to stay driven by the first read or write of the next handlers,
// and the TLS version and cipher are not known before it.
type tlsMetricsConn struct {
	*tls.Conn

	gauge  gokitmetrics.Gauge
	labels []string

	settled     atomic.Bool
	mu          sync.Mutex
	established gokitmetrics.Gauge
}

func (c *tlsMetricsConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.observe()

	return n, err
}

func (c *tlsMetricsConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.observe()

	return n, err
}

func (c *tlsMetricsConn) observe() {
	if c.settled.Load() {
		return
	}

	state := c.Conn.ConnectionState()
	if !state.HandshakeComplete {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.settled.Load() {
		return
	}

	c.established = c.gauge.With(append(c.labels, tlsProtocolVersionLabel, traefiktls.GetVersion(&state), tlsCipherLabel, traefiktls.GetCipherName(&state))...)
	c.established.Add(1)
	c.settled.Store(true)
}

func (c *tlsMetricsConn) release() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.established != nil {
		c.established.Add(-1)
		c.established = nil
	}

	c.settled.Store(true)
}
