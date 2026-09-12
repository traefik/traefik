package inflightconn

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

func TestInFlightConn_ServeTCP(t *testing.T) {
	proceedCh := make(chan struct{})
	waitCh := make(chan struct{})

	next := tcp.HandlerFunc(func(conn tcp.WriteCloser) {
		proceedCh <- struct{}{}

		if fc, ok := conn.(fakeConn); !ok || !fc.wait {
			return
		}

		<-waitCh
	})

	middleware, err := New(t.Context(), next, dynamic.TCPInFlightConn{Amount: 1}, "foo")
	require.NoError(t, err)

	// The first connection should succeed and wait.
	firstServedCh := make(chan struct{})
	go func() {
		middleware.ServeTCP(fakeConn{addr: "127.0.0.1:9000", wait: true})
		close(firstServedCh)
	}()
	requireMessage(t, proceedCh)

	closeCh := make(chan struct{})

	// The second connection from the same remote address should be closed as the maximum number of connections is exceeded.
	go middleware.ServeTCP(fakeConn{addr: "127.0.0.1:9000", closeCh: closeCh})
	requireMessage(t, closeCh)

	// The connection from another remote address should succeed.
	go middleware.ServeTCP(fakeConn{addr: "127.0.0.2:9000"})
	requireMessage(t, proceedCh)

	// Once the first connection is closed, next connection with the same remote address should succeed.
	// Waiting for ServeTCP to return, and not only for the handler to be released,
	// is what guarantees that the deferred decrement has already run.
	close(waitCh)
	requireMessage(t, firstServedCh)

	go middleware.ServeTCP(fakeConn{addr: "127.0.0.1:9000"})
	requireMessage(t, proceedCh)
}

func TestInFlightConn_ReleasedConnectionsAreForgotten(t *testing.T) {
	middleware, err := New(t.Context(), tcp.HandlerFunc(func(tcp.WriteCloser) {}), dynamic.TCPInFlightConn{Amount: 1}, "foo")
	require.NoError(t, err)

	inFlight, ok := middleware.(*inFlightConn)
	require.True(t, ok)

	// Each connection is fully served, and therefore released, before the next one starts.
	for i := range 256 {
		inFlight.ServeTCP(fakeConn{addr: fmt.Sprintf("10.0.0.%d:9000", i)})
	}

	inFlight.mu.Lock()
	defer inFlight.mu.Unlock()

	assert.Empty(t, inFlight.connections)
}

func requireMessage(t *testing.T, c chan struct{}) {
	t.Helper()
	select {
	case <-c:
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

type fakeConn struct {
	net.Conn

	addr    string
	wait    bool
	closeCh chan struct{}
}

func (c fakeConn) RemoteAddr() net.Addr {
	return fakeAddr{addr: c.addr}
}

func (c fakeConn) Close() error {
	// The middleware closes every connection it rejects, including the ones
	// the test does not watch.
	if c.closeCh != nil {
		close(c.closeCh)
	}

	return nil
}

func (c fakeConn) CloseWrite() error {
	panic("implement me")
}

type fakeAddr struct {
	addr string
}

func (a fakeAddr) Network() string {
	return "tcp"
}

func (a fakeAddr) String() string {
	return a.addr
}
