package accesslog

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

const (
	tcpProtocol              = "TCP"
	connectionStatusSuccess  = "success"
	connectionStatusRejected = "rejected"

	// Written when a middleware closes the connection without recording a reason.
	tcpClosedBeforeService = "connection closed before reaching the service"
)

// WrapTCP logs one access-log line per TCP connection, using the same handler as HTTP.
// The wrapper sits outside the middleware chain, so a middleware that closes the
// connection without calling next is still recorded.
func (h *Handler) WrapTCP(entryPoint, routerName, serviceName string, next tcp.Handler) tcp.Handler {
	if h == nil || next == nil {
		return next
	}
	return &tcpAccessLogHandler{
		handler:     h,
		entryPoint:  entryPoint,
		routerName:  routerName,
		serviceName: serviceName,
		next:        next,
	}
}

// IncludeInternalRouters reports whether TCP routers whose name ends with @internal are logged.
func (h *Handler) IncludeInternalRouters() bool {
	return h != nil && h.config != nil && h.config.AddInternals
}

type tcpAccessLogHandler struct {
	handler     *Handler
	entryPoint  string
	routerName  string
	serviceName string
	next        tcp.Handler
}

func (t *tcpAccessLogHandler) ServeTCP(ctx context.Context, conn tcp.WriteCloser) {
	ctx = tcp.WithMetadata(ctx)
	start := time.Now().UTC()
	state := &tcpLogState{}
	observed := &observingConn{WriteCloser: conn, state: state}

	t.next.ServeTCP(ctx, observed)

	t.handler.logTCP(ctx, t.entryPoint, t.routerName, t.serviceName, observed, start)
}

type tcpLogState struct {
	mu        sync.Mutex
	reached   bool
	rejection error
	read      atomic.Int64
	written   atomic.Int64
}

// observingConn counts bytes and records whether the service was reached
// and why a middleware rejected the connection.
type observingConn struct {
	tcp.WriteCloser
	state *tcpLogState
}

func (c *observingConn) Read(p []byte) (int, error) {
	n, err := c.WriteCloser.Read(p)
	if n > 0 {
		c.state.read.Add(int64(n))
	}
	return n, err
}

func (c *observingConn) Write(p []byte) (int, error) {
	n, err := c.WriteCloser.Write(p)
	if n > 0 {
		c.state.written.Add(int64(n))
	}
	return n, err
}

// MarkBackend records that the TCP service handler was invoked.
func (c *observingConn) MarkBackend() {
	c.state.mu.Lock()
	c.state.reached = true
	c.state.mu.Unlock()
}

// SetRejection records the first reason a middleware closed the connection.
func (c *observingConn) SetRejection(err error) {
	if err == nil {
		return
	}
	c.state.mu.Lock()
	if c.state.rejection == nil {
		c.state.rejection = err
	}
	c.state.mu.Unlock()
}

func (h *Handler) logTCP(ctx context.Context, entryPoint, routerName, serviceName string, conn *observingConn, start time.Time) {
	if h == nil || h.config == nil || h.logger == nil || conn == nil {
		return
	}

	duration := time.Since(start)
	if !h.keepTCPAccessLog(duration) {
		return
	}

	conn.state.mu.Lock()
	reached := conn.state.reached
	rejection := conn.state.rejection
	conn.state.mu.Unlock()

	status := connectionStatusSuccess
	if !reached {
		status = connectionStatusRejected
	}

	clientAddr, clientHost, clientPort := "-", "-", "-"
	if conn.RemoteAddr() != nil {
		clientAddr = conn.RemoteAddr().String()
		clientHost, clientPort = silentSplitHostPort(clientAddr)
	}

	core := map[string]any{
		logs.EntryPointName:   entryPoint,
		RouterName:            routerName,
		ServiceName:           serviceName,
		ClientAddr:            clientAddr,
		ClientHost:            clientHost,
		ClientPort:            clientPort,
		RequestProtocol:       tcpProtocol,
		RequestMethod:         "-",
		RequestPath:           "-",
		StartUTC:              start,
		Duration:              duration,
		RequestContentSize:    conn.state.read.Load(),
		DownstreamContentSize: conn.state.written.Load(),
		ConnectionStatus:      status,
	}
	if !reached {
		if rejection != nil {
			core[Error] = rejection.Error()
		} else {
			core[Error] = tcpClosedBeforeService
		}
	}

	fields := logrus.Fields{}
	taken := map[string]struct{}{}
	for k, v := range core {
		taken[strings.ToLower(k)] = struct{}{}
		if h.config.Fields.Keep(strings.ToLower(k)) {
			fields[k] = v
		}
	}

	for k, v := range tcp.Metadata(ctx) {
		if _, reserved := taken[strings.ToLower(k)]; reserved || k == "" {
			continue
		}
		if h.config.Fields.Keep(strings.ToLower(k)) {
			fields[k] = v
		}
	}

	h.writeEntry(ctx, core, fields)
}

// keepTCPAccessLog applies only minDuration. HTTP status codes and retry attempts
// do not describe a TCP connection, so a filter that sets only those still keeps the line.
func (h *Handler) keepTCPAccessLog(duration time.Duration) bool {
	if h.config.Filters == nil || h.config.Filters.MinDuration == 0 {
		return true
	}
	return ptypes.Duration(duration) > h.config.Filters.MinDuration
}
