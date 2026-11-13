package udp

import (
	"crypto/rand"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxy_ServeUDP(t *testing.T) {
	backendAddr := newServer(t, HandlerFunc(func(conn *Conn) {
		for {
			b := make([]byte, 1024*1024)
			n, err := conn.Read(b)
			if err != nil {
				return
			}

			if _, err := conn.Write(b[:n]); err != nil {
				return
			}
		}
	}))

	proxy, err := NewProxy(backendAddr)
	require.NoError(t, err)

	proxyAddr := newServer(t, proxy)

	udpConn, err := net.Dial("udp", proxyAddr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = udpConn.Close() })

	_, err = udpConn.Write([]byte("DATAWRITE"))
	require.NoError(t, err)

	b := make([]byte, 1024*1024)
	n, err := udpConn.Read(b)
	require.NoError(t, err)

	assert.Equal(t, "DATAWRITE", string(b[:n]))
}

func TestProxy_ServeUDP_MaxDataSize(t *testing.T) {
	if runtime.GOOS == "darwin" {
		// sudo sysctl -w net.inet.udp.maxdgram=65507
		t.Skip("Skip test on darwin as the maximum dgram size is set to 9216 bytes by default")
	}

	// Theoretical maximum size of data in a UDP datagram.
	// 65535 − 8 (UDP header) − 20 (IP header).
	dataSize := 65507

	backendAddr := newServer(t, HandlerFunc(func(conn *Conn) {
		buffer := make([]byte, dataSize)

		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		_, _ = conn.Write(buffer[:n])
	}))

	proxy, err := NewProxy(backendAddr)
	require.NoError(t, err)

	proxyAddr := newServer(t, proxy)

	udpConn, err := net.Dial("udp", proxyAddr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = udpConn.Close() })

	want := make([]byte, dataSize)

	_, err = rand.Read(want)
	require.NoError(t, err)

	_, err = udpConn.Write(want)
	require.NoError(t, err)

	got := make([]byte, dataSize)

	_, err = udpConn.Read(got)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

// newServer serves handler on a free port until the test ends, and returns its address.
// The handlers outlive the test until the listener closes their sessions,
// so they return on errors rather than reporting them to a finished test.
func newServer(t *testing.T, handler Handler) string {
	t.Helper()

	listener, err := Listen(net.ListenConfig{}, "udp", "127.0.0.1:0", 3*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go handler.ServeUDP(conn)
		}
	}()

	return listener.Addr().String()
}
