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

	err = i.acquire(ip)
	if err != nil {
		logger.Debug("Connection rejected: %v", err)
		conn.Close()
		return
	}

	defer i.release(ip)

	i.next.ServeTCP(conn)
}

func (i *inFlightConn) increment(ip string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.connections[ip] >= i.maxConnections {
		return fmt.Errorf("max number of connections reached for %s", ip)
	}

	i.connections[ip]++

	return nil
}

func (i *inFlightConn) decrement(ip string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.connections[ip]--
}
