package inflightconn

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

func TestInFlightConn_ServeTCP(t *testing.T) {
	proceedCh := make(chan struct{})
	waitCh := make(chan struct{})
	finishCh := make(chan struct{})

	next := tcp.HandlerFunc(func(conn tcp.WriteCloser) {
		proceedCh <- struct{}{}

		if fc, ok := conn.(fakeConn); !ok || !fc.wait {
			return
		}

		<-waitCh
		finishCh <- struct{}{}
	})

	middleware, err := New(t.Context(), next, dynamic.TCPInFlightConn{Amount: 1}, "foo")
	require.NoError(t, err)

	// The first connection should succeed and wait.
	go middleware.ServeTCP(fakeConn{addr: "127.0.0.1:9000", wait: true})
	requireMessage(t, proceedCh)

	closeCh := make(chan struct{})

	// The second connection from the same remote address should be closed as the maximum number of connections is exceeded.
	go middleware.ServeTCP(fakeConn{addr: "127.0.0.1:9000", closeCh: closeCh})
	requireMessage(t, closeCh)

	// The connection from another remote address should succeed.
	go middleware.ServeTCP(fakeConn{addr: "127.0.0.2:9000"})
	requireMessage(t, proceedCh)

	// Once the first connection is closed, next connection with the same remote address should succeed.
	close(waitCh)
	requireMessage(t, finishCh)

	go middleware.ServeTCP(fakeConn{addr: "127.0.0.1:9000"})
	requireMessage(t, proceedCh)
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
	close(c.closeCh)
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
