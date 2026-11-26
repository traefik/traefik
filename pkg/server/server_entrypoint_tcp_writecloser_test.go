package server

import (
	"net"
	"testing"

	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// closeWriteConn is a connection that is not a *net.TCPConn but does close
// its write half, which is what a userspace network stack hands over.
type closeWriteConn struct {
	net.Conn

	closedWrite bool
}

func (c *closeWriteConn) CloseWrite() error {
	c.closedWrite = true
	return nil
}

func TestWriteCloserAcceptsAnyConnThatClosesItsWriteHalf(t *testing.T) {
	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	conn := &closeWriteConn{Conn: server}

	writeCloser, err := writeCloser(conn)
	require.NoError(t, err)

	require.NoError(t, writeCloser.CloseWrite())
	assert.True(t, conn.closedWrite)
}

func TestWriteCloserAcceptsProxyProtocolOverAnyConnThatClosesItsWriteHalf(t *testing.T) {
	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	conn := &closeWriteConn{Conn: server}

	writeCloser, err := writeCloser(proxyproto.NewConn(conn))
	require.NoError(t, err)

	require.NoError(t, writeCloser.CloseWrite())
	assert.True(t, conn.closedWrite)
}

// A connection that cannot close its write half is still rejected, rather
// than silently proxied without the half-close the TCP path relies on.
func TestWriteCloserRejectsConnWithoutCloseWrite(t *testing.T) {
	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	_, err := writeCloser(server)
	require.ErrorContains(t, err, "unknown connection type")
}
