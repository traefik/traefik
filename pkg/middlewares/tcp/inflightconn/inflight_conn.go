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
// The connections are identified and grouped by remote IP.
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

	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		logger.Errorf("Cannot parse IP from remote addr: %v", err)
		conn.Close()
		return
	}

	if err = i.increment(ip); err != nil {
		logger.Errorf("Connection rejected: %v", err)
		conn.Close()
		return
	}

	defer i.decrement(ip)

	i.next.ServeTCP(conn)
}

// increment increases the counter for the number of connections tracked for the
// given IP.
// It returns an error if the counter would go above the max allowed number of
// connections.
func (i *inFlightConn) increment(ip string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.connections[ip] >= i.maxConnections {
		return fmt.Errorf("max number of connections reached for %s", ip)
	}

	i.connections[ip]++

	return nil
}

// decrement decreases the counter for the number of connections tracked for the
// given IP.
// It ensures that the counter does not go below zero.
func (i *inFlightConn) decrement(ip string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.connections[ip] <= 0 {
		return
	}

	i.connections[ip]--
}
