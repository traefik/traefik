package metrics

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	gokitmetrics "github.com/go-kit/kit/metrics"
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

// The TLS modes of a TCP router.
const (
	// TLSModeNone is the mode of a router without TLS.
	TLSModeNone = "none"
	// TLSModeTerminated is the mode of a router whose TLS is terminated by Traefik.
	TLSModeTerminated = "terminated"
	// TLSModePassthrough is the mode of a router that forwards the TLS connections as is.
	TLSModePassthrough = "passthrough"
)

type metricsMiddleware struct {
	next                    tcp.Handler
	openConnectionsGauge    gokitmetrics.Gauge
	openTLSConnectionsGauge gokitmetrics.Gauge
}

// NewRouterMiddleware creates a new metrics middleware for a TCP Router.
func NewRouterMiddleware(ctx context.Context, next tcp.Handler, registry metrics.Registry, routerName string, serviceName string, tlsMode string) tcp.Handler {
	middlewares.GetLogger(ctx, nameRouter, typeName).Debug().Msg("Creating middleware")

	return &metricsMiddleware{
		next:                    next,
		openConnectionsGauge:    registry.RouterTCPOpenConnectionsGauge().With(routerNameLabel, routerName, serviceNameLabel, serviceName, tlsModeLabel, tlsMode),
		openTLSConnectionsGauge: registry.RouterTCPOpenTLSConnectionsGauge().With(routerNameLabel, routerName, serviceNameLabel, serviceName),
	}
}

// RouterMetricsHandler returns the metrics TCP router handler.
func RouterMetricsHandler(ctx context.Context, registry metrics.Registry, routerName string, serviceName string, tlsMode string) tcp.Constructor {
	return func(next tcp.Handler) (tcp.Handler, error) {
		if registry == nil || !registry.IsTCPRouterEnabled() {
			return next, nil
		}

		return NewRouterMiddleware(ctx, next, registry, routerName, serviceName, tlsMode), nil
	}
}

// ServeTCP serves the given TCP connection.
func (m *metricsMiddleware) ServeTCP(conn tcp.WriteCloser) {
	networkLabels := getNetworkLabels(conn)

	openConnectionsGauge := m.openConnectionsGauge.With(networkLabels...)
	openConnectionsGauge.Add(1)
	defer openConnectionsGauge.Add(-1)

	// Only the connections whose TLS is terminated by Traefik are TLS connections here,
	// the passthrough ones are forwarded without their TLS parameters being known.
	if tlsConn, ok := conn.(*tls.Conn); ok {
		mConn := &tlsMetricsConn{Conn: tlsConn, gauge: m.openTLSConnectionsGauge.With(networkLabels...)}
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

	gauge gokitmetrics.Gauge

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

	c.established = c.gauge.With(tlsProtocolVersionLabel, traefiktls.GetVersion(&state), tlsCipherLabel, traefiktls.GetCipherName(&state))
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
