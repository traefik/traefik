package tcpinflightconn

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/tcp"
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

	middleware, err := New(context.Background(), next, dynamic.TCPInFlightConn{Amount: 1}, "foo")
	require.NoError(t, err)

	// The first connection should succeed and wait.
	go middleware.ServeTCP(fakeConn{addr: "127.0.0.1", wait: true})
	<-proceedCh

	// The second connection should be closed as the maximum number of connections is exceeded.
	closeCh := make(chan struct{})

	go middleware.ServeTCP(fakeConn{addr: "127.0.0.1", closeCh: closeCh})
	<-closeCh

	// The connection with another remote address should succeed.
	go middleware.ServeTCP(fakeConn{addr: "127.0.0.2"})
	<-proceedCh

	// Once the first connection is closed, next connection with the same remote address should succeed.
	close(waitCh)
	<-finishCh

	go middleware.ServeTCP(fakeConn{addr: "127.0.0.1"})
	<-proceedCh
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
