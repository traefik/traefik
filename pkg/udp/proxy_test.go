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


func TestProxyProtocol_UDP(t *testing.T) {
	backendAddr := ":8091"
	received := make(chan struct {
		addr net.Addr
		data []byte
	}, 1)

	// Backend server records remote address and payload
	go newServer(t, backendAddr, HandlerFunc(func(conn *Conn) {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		require.NoError(t, err)
		received <- struct {
			addr net.Addr
			data []byte
		}{conn.rAddr, buf[:n]}
	}))

	proxy, err := NewProxy(backendAddr)
	require.NoError(t, err)

	proxyAddr := ":8090"
	go newServer(t, proxyAddr, proxy)

	time.Sleep(time.Second)

	// Prepare a UDP packet with a PROXY protocol v1 header

	srcAddr, err := net.ResolveUDPAddr("udp", "127.0.0.2:55555")
	require.NoError(t, err)
	dstAddr, err := net.ResolveUDPAddr("udp", proxyAddr)
	require.NoError(t, err)

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	require.NoError(t, err)

	header := "PROXY UDP4 127.0.0.2 127.0.0.1 55555 8090\r\n"
	payload := []byte("PROXYTEST")
	packet := append([]byte(header), payload...)

	_, err = conn.Write(packet)
	require.NoError(t, err)

	select {
	case got := <-received:
		assert.Equal(t, payload, got.data)
		// The backend should see the remote address as 127.0.0.2:55555
		udpAddr, ok := got.addr.(*net.UDPAddr)
		require.True(t, ok)
		assert.Equal(t, "127.0.0.2", udpAddr.IP.String())
		assert.Equal(t, 55555, udpAddr.Port)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for backend to receive packet")
	}
}
