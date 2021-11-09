package tcpinflightconn

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

const typeName = "InFlightConnTCP"

type inFlightConn struct {
	name           string
	next           tcp.Handler
	maxConnections int64

	mu          sync.Mutex
	connections map[string]int64 // current number of connections by remote IP.
}

// New creates a max connections middleware.
// The connections are limited by remote IP.
func New(ctx context.Context, next tcp.Handler, config dynamic.TCPInFlightConn, name string) (tcp.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName))
	logger.Debug("Creating middleware")

	return &inFlightConn{
		name:           name,
		next:           next,
		connections:    make(map[string]int64),
		maxConnections: config.Amount,
	}, nil
}

// ServeTCP serves the given TCP connection.
func (i *inFlightConn) ServeTCP(conn tcp.WriteCloser) {
	ctx := middlewares.GetLoggerCtx(context.Background(), i.name, typeName)
	logger := log.FromContext(ctx)

	err := i.acquire(conn.RemoteAddr())
	if err != nil {
		logger.Debug("Connection rejected: %v", err)
		conn.Close()
		return
	}

	defer i.release(conn.RemoteAddr())

	i.next.ServeTCP(conn)
}

func (i *inFlightConn) acquire(addr net.Addr) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	ip := addr.String()

	if i.connections[ip] >= i.maxConnections {
		return fmt.Errorf("max connection reached for %s", ip)
	}

	i.connections[ip]++

	return nil
}

func (i *inFlightConn) release(addr net.Addr) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.connections[addr.String()]--
}
