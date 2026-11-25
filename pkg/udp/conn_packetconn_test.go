package udp

import (
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wrappedPacketConn is a net.PacketConn that is not a *net.UDPConn, which is
// what a userspace network stack hands over: its packets do not come from a
// host socket.
type wrappedPacketConn struct {
	net.PacketConn
}

// A listener must work on any net.PacketConn. It used to require a
// *net.UDPConn, which silently ruled out every packet source that is not a
// host socket.
func TestListenPacketConnAcceptsAnyPacketConn(t *testing.T) {
	underlying, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	listener, err := ListenPacketConn(&wrappedPacketConn{PacketConn: underlying}, 3*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Shutdown(0) })

	assert.Equal(t, underlying.LocalAddr().String(), listener.Addr().String())

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		_, _ = conn.Write(buf[:n])
	}()

	client, err := net.Dial("udp", listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.Write([]byte("hello"))
	require.NoError(t, err)

	require.NoError(t, client.SetReadDeadline(time.Now().Add(5*time.Second)))

	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	require.NoError(t, err)

	assert.Equal(t, "hello", string(buf[:n]))
}

func TestListenPacketConnRequiresAConn(t *testing.T) {
	_, err := ListenPacketConn(nil, time.Second)
	require.ErrorContains(t, err, "packet conn is required")
}

// A packet conn does not always name the sender: a unix datagram socket does
// not, for a peer that is not bound. Such a datagram opens no session, and
// must not stop the listener either.
func TestListenPacketConnSkipsADatagramWithoutSender(t *testing.T) {
	path := filepath.Join(t.TempDir(), "udp.sock")

	server, err := net.ListenPacket("unixgram", path)
	require.NoError(t, err)

	listener, err := ListenPacketConn(server, 3*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Shutdown(0) })

	unbound, err := net.Dial("unixgram", path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = unbound.Close() })

	_, err = unbound.Write([]byte("anonymous"))
	require.NoError(t, err)

	bound, err := net.DialUnix("unixgram", &net.UnixAddr{Name: filepath.Join(t.TempDir(), "client.sock"), Net: "unixgram"}, &net.UnixAddr{Name: path, Net: "unixgram"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = bound.Close() })

	_, err = bound.Write([]byte("hello"))
	require.NoError(t, err)

	accepted := make(chan *Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	select {
	case conn := <-accepted:
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		require.NoError(t, err)

		assert.Equal(t, "hello", string(buf[:n]))
	case <-time.After(5 * time.Second):
		t.Fatal("the listener stopped on the datagram that had no sender")
	}
}

// failingPacketConn is a packet conn whose reads fail while it is open,
// as the conn of a userspace network stack may, and a host socket hardly ever does.
type failingPacketConn struct {
	net.PacketConn

	closed chan struct{}
}

func (c *failingPacketConn) ReadFrom([]byte) (int, net.Addr, error) {
	return 0, nil, errors.New("network unavailable")
}

func (c *failingPacketConn) Close() error {
	close(c.closed)

	return c.PacketConn.Close()
}

// A read error that is not the closing of the listener ends the reading for good:
// the listener must close then, so that Accept returns rather than wait on a listener that serves nothing.
func TestListenPacketConnClosesOnAReadError(t *testing.T) {
	underlying, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	conn := &failingPacketConn{PacketConn: underlying, closed: make(chan struct{})}

	listener, err := ListenPacketConn(conn, 3*time.Second)
	require.NoError(t, err)

	accepted := make(chan error, 1)
	go func() {
		_, err := listener.Accept()
		accepted <- err
	}()

	select {
	case err := <-accepted:
		require.ErrorIs(t, err, errClosedListener)
	case <-time.After(5 * time.Second):
		t.Fatal("Accept is still waiting on a listener that reads nothing")
	}

	select {
	case <-conn.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("the packet conn was left open")
	}

	// Shutting down a listener that closed by itself is not an error, and does not close it twice.
	require.NoError(t, listener.Shutdown(0))
}
