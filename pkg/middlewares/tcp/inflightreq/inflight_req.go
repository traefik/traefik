package tcpinflightreq

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/tcp"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	typeName = "InFlightReqTCP"
)

type inFlightReq struct {
	ctx context.Context
	// maxConnections maximum amount of allowed simultaneous in-flight request
	maxConnections int64
	mutex          *sync.Mutex
	name           string
	next           tcp.Handler
	sourceMatcher  func(conn tcp.WriteCloser) (string, error)
	// totalConnections current number of connections
	totalConnections int64
}

// New creates a max request middleware.
// If no source criterion is provided in the config, it defaults to RequestHost.
func New(ctx context.Context, next tcp.Handler, config dynamic.TCPInFlightReq, name string) (tcp.Handler, error) {
	ctxLog := log.With(ctx, log.Str(log.MiddlewareName, name), log.Str(log.MiddlewareType, typeName))
	log.FromContext(ctxLog).Debug("Creating middleware")

	sourceMatcher := func(conn tcp.WriteCloser) (string, error) {
		ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		return ip, err
	}

	return &inFlightReq{
		ctx:            ctx,
		maxConnections: config.Amount,
		name:           name,
		next:           next,
		sourceMatcher:  sourceMatcher,
	}, nil
}

func (i *inFlightReq) GetTracingInformation() (string, ext.SpanKindEnum) {
	return i.name, tracing.SpanKindNoneEnum
}

func (i *inFlightReq) ServeTCP(conn tcp.WriteCloser) {
	token, err := i.sourceMatcher(conn)
	if err != nil {
		log.FromContext(i.ctx).Debugf("Failed to extract source of the connection: %v", err)
		conn.Close()
		return
	}

	if i.totalConnections >= i.maxConnections {
		log.FromContext(i.ctx).Debugf("Limiting request source %s: no more space available", token)
		conn.Close()
		return
	}

	atomic.AddInt64(&i.totalConnections, 1)

	defer atomic.AddInt64(&i.totalConnections, -1)

	i.next.ServeTCP(conn)
}
