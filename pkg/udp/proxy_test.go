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
	backendAddr := ":8081"
	go newServer(t, backendAddr, HandlerFunc(func(conn *Conn) {
		for {
			b := make([]byte, 1024*1024)
			n, err := conn.Read(b)
			require.NoError(t, err)

			_, err = conn.Write(b[:n])
			require.NoError(t, err)
		}
	}))

	proxy, err := NewProxy(backendAddr)
	require.NoError(t, err)

	proxyAddr := ":8080"
	go newServer(t, proxyAddr, proxy)

	time.Sleep(time.Second)

	udpConn, err := net.Dial("udp", proxyAddr)
	require.NoError(t, err)

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

	backendAddr := ":8083"
	go newServer(t, backendAddr, HandlerFunc(func(conn *Conn) {
		buffer := make([]byte, dataSize)

		n, err := conn.Read(buffer)
		require.NoError(t, err)

		_, err = conn.Write(buffer[:n])
		require.NoError(t, err)
	}))

	proxy, err := NewProxy(backendAddr)
	require.NoError(t, err)

	proxyAddr := ":8082"
	go newServer(t, proxyAddr, proxy)

	time.Sleep(time.Second)

	udpConn, err := net.Dial("udp", proxyAddr)
	require.NoError(t, err)

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

func newServer(t *testing.T, addr string, handler Handler) {
	t.Helper()

	listener, err := Listen(net.ListenConfig{}, "udp", addr, 3*time.Second)
	require.NoError(t, err)

	for {
		conn, err := listener.Accept()
		require.NoError(t, err)

		go handler.ServeUDP(conn)
	}
}
