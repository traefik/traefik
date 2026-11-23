package metrics

import (
	"context"
	"crypto/tls"
	"net"
	"slices"
	"sync"

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

	// The labels are named after the attributes set on the tracing spans,
	// and after the OpenTelemetry semantic conventions attributes.
	entryPointNameLabel     = "traefik.entrypoint.name"
	routerNameLabel         = "traefik.router.name"
	serviceNameLabel        = "traefik.service.name"
	tlsModeLabel            = "traefik.tls.mode"
	networkTypeLabel        = "network.type"
	tlsProtocolVersionLabel = "tls.protocol.version"
	tlsCipherLabel          = "tls.cipher"

	// The values of the TLS mode label.
	tlsModeNone        = "none"
	tlsModeTerminated  = "terminated"
	tlsModePassthrough = "passthrough"
)

// The gauges are given all their labels at once, when a connection is served,
// because the Prometheus gauge does not support being given its labels in several steps.
// The counters share the labels of the gauges, so that both can be compared.
type metricsMiddleware struct {
	next tcp.Handler

	connectionsCounter    gokitmetrics.Counter
	openConnectionsGauge  gokitmetrics.Gauge
	receivedBytesCounter  gokitmetrics.Counter
	sentBytesCounter      gokitmetrics.Counter
	openConnectionsLabels []string

	tlsConnectionsCounter    gokitmetrics.Counter
	tlsOpenConnectionsGauge  gokitmetrics.Gauge
	tlsReceivedBytesCounter  gokitmetrics.Counter
	tlsSentBytesCounter      gokitmetrics.Counter
	tlsOpenConnectionsLabels []string
}

// NewRouterMiddleware creates a new metrics middleware for a TCP Router.
func NewRouterMiddleware(ctx context.Context, next tcp.Handler, registry metrics.Registry, entryPointName, routerName string, serviceName string, tlsConfig *dynamic.RouterTCPTLSConfig) tcp.Handler {
	middlewares.GetLogger(ctx, nameRouter, typeName).Debug().Msg("Creating middleware")

	tlsMode := tlsModeNone
	if tlsConfig != nil {
		if tlsConfig.Passthrough {
			tlsMode = tlsModePassthrough
		} else {
			tlsMode = tlsModeTerminated
		}
	}

	return &metricsMiddleware{
		next:                     next,
		connectionsCounter:       registry.RouterTCPConnectionsCounter(),
		openConnectionsGauge:     registry.RouterTCPOpenConnectionsGauge(),
		receivedBytesCounter:     registry.RouterTCPReceivedBytesCounter(),
		sentBytesCounter:         registry.RouterTCPSentBytesCounter(),
		openConnectionsLabels:    []string{entryPointNameLabel, entryPointName, routerNameLabel, routerName, serviceNameLabel, serviceName, tlsModeLabel, tlsMode},
		tlsConnectionsCounter:    registry.RouterTCPTLSConnectionsCounter(),
		tlsOpenConnectionsGauge:  registry.RouterTCPTLSOpenConnectionsGauge(),
		tlsReceivedBytesCounter:  registry.RouterTCPTLSReceivedBytesCounter(),
		tlsSentBytesCounter:      registry.RouterTCPTLSSentBytesCounter(),
		tlsOpenConnectionsLabels: []string{entryPointNameLabel, entryPointName, routerNameLabel, routerName, serviceNameLabel, serviceName},
	}
}

// RouterMetricsHandler returns the metrics TCP router handler.
func RouterMetricsHandler(ctx context.Context, registry metrics.Registry, entryPointName, routerName string, serviceName string, tlsConfig *dynamic.RouterTCPTLSConfig) tcp.Constructor {
	return func(next tcp.Handler) (tcp.Handler, error) {
		if registry == nil || !registry.IsTCPRouterEnabled() {
			return next, nil
		}

		return NewRouterMiddleware(ctx, next, registry, entryPointName, routerName, serviceName, tlsConfig), nil
	}
}

// ServeTCP serves the given TCP connection.
func (m *metricsMiddleware) ServeTCP(conn tcp.WriteCloser) {
	networkLabels := getNetworkLabels(conn)

	labels := slices.Concat(m.openConnectionsLabels, networkLabels)

	m.connectionsCounter.With(labels...).Add(1)

	openConnectionsGauge := m.openConnectionsGauge.With(labels...)
	openConnectionsGauge.Add(1)
	defer openConnectionsGauge.Add(-1)

	// Only the connections whose TLS is terminated by Traefik are TLS connections here,
	// the passthrough ones are forwarded without their TLS parameters being known.
	if tlsConn, ok := conn.(*tls.Conn); ok {
		mConn := &tlsMetricsConn{
			Conn:                    tlsConn,
			tlsConnectionsCounter:   m.tlsConnectionsCounter,
			tlsOpenConnectionsGauge: m.tlsOpenConnectionsGauge,
			tlsReceivedBytesCounter: m.tlsReceivedBytesCounter,
			tlsSentBytesCounter:     m.tlsSentBytesCounter,
			labels:                  slices.Concat(m.tlsOpenConnectionsLabels, networkLabels),
		}
		defer mConn.release()

		conn = mConn
	}

	conn = &bytesMetricsConn{
		WriteCloser:   conn,
		receivedBytes: m.receivedBytesCounter.With(labels...),
		sentBytes:     m.sentBytesCounter.With(labels...),
	}

	m.next.ServeTCP(conn)
}

// getNetworkLabels returns the network.type label of the given connection.
// The value is empty when the local address is not a TCP one,
// as the metrics exporters require a constant set of labels.
func getNetworkLabels(conn net.Conn) []string {
	addr, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		return []string{networkTypeLabel, ""}
	}

	networkType := "ipv6"
	if addr.IP.To4() != nil {
		networkType = "ipv4"
	}

	return []string{networkTypeLabel, networkType}
}

// bytesMetricsConn counts the bytes read from and written to a connection as they are transferred,
// rather than when the connection is closed, as a TCP connection can stay open for a long time.
type bytesMetricsConn struct {
	tcp.WriteCloser

	receivedBytes gokitmetrics.Counter
	sentBytes     gokitmetrics.Counter
}

func (c *bytesMetricsConn) Read(b []byte) (int, error) {
	n, err := c.WriteCloser.Read(b)
	if n > 0 {
		c.receivedBytes.Add(float64(n))
	}

	return n, err
}

func (c *bytesMetricsConn) Write(b []byte) (int, error) {
	n, err := c.WriteCloser.Write(b)
	if n > 0 {
		c.sentBytes.Add(float64(n))
	}

	return n, err
}

// tlsMetricsConn counts a TLS connection, and counts it as open, once its handshake is done,
// and then counts the application bytes read from and written to it.
// The handshake is not forced, as it has to stay driven by the first read or write of the next handlers,
// and the TLS version and cipher are not known before it.
type tlsMetricsConn struct {
	*tls.Conn

	tlsConnectionsCounter   gokitmetrics.Counter
	tlsOpenConnectionsGauge gokitmetrics.Gauge
	tlsReceivedBytesCounter gokitmetrics.Counter
	tlsSentBytesCounter     gokitmetrics.Counter
	labels                  []string

	once sync.Once
	// The gauge is kept with the labels it was given when the connection was counted,
	// so that the connection is released with the very same labels.
	// It is nil as long as the connection is not counted.
	tlsOpenConnectionsGaugeWithLabels gokitmetrics.Gauge
	// The bytes counters are kept with their labels as well, as they are added to on every read and write.
	// They are nil as long as the connection is not counted.
	tlsReceivedBytesCounterWithLabels gokitmetrics.Counter
	tlsSentBytesCounterWithLabels     gokitmetrics.Counter
}

// A read or a write of a TLS connection only succeeds, or transfers bytes, once the handshake is done.
// When it fails, the handshake state tells whether the failure happened after the handshake,
// e.g. when the client closes the connection without sending any data.
func (c *tlsMetricsConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err == nil || c.Conn.ConnectionState().HandshakeComplete {
		c.once.Do(c.observe)
	}

	// The connection is not counted when it was released before its handshake was done.
	if n > 0 && c.tlsReceivedBytesCounterWithLabels != nil {
		c.tlsReceivedBytesCounterWithLabels.Add(float64(n))
	}

	return n, err
}

func (c *tlsMetricsConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if err == nil || c.Conn.ConnectionState().HandshakeComplete {
		c.once.Do(c.observe)
	}

	if n > 0 && c.tlsSentBytesCounterWithLabels != nil {
		c.tlsSentBytesCounterWithLabels.Add(float64(n))
	}

	return n, err
}

func (c *tlsMetricsConn) observe() {
	state := c.Conn.ConnectionState()
	labels := slices.Concat(c.labels, []string{tlsProtocolVersionLabel, traefiktls.GetVersion(&state), tlsCipherLabel, traefiktls.GetCipherName(&state)})

	c.tlsConnectionsCounter.With(labels...).Add(1)

	c.tlsOpenConnectionsGaugeWithLabels = c.tlsOpenConnectionsGauge.With(labels...)
	c.tlsOpenConnectionsGaugeWithLabels.Add(1)

	c.tlsReceivedBytesCounterWithLabels = c.tlsReceivedBytesCounter.With(labels...)
	c.tlsSentBytesCounterWithLabels = c.tlsSentBytesCounter.With(labels...)
}

func (c *tlsMetricsConn) release() {
	// Waits for an ongoing observation, and prevents the connection from being counted afterwards.
	c.once.Do(func() {})

	if c.tlsOpenConnectionsGaugeWithLabels != nil {
		c.tlsOpenConnectionsGaugeWithLabels.Add(-1)
	}
}
