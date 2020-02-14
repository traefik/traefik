package udp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUDPProxy(t *testing.T) {
	backendAddr := ":8081"
	go newServer(t, ":8081", HandlerFunc(func(conn *Conn) {
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

func newServer(t *testing.T, addr string, handler Handler) {
	addrL, err := net.ResolveUDPAddr("udp", addr)
	require.NoError(t, err)

	listener, err := Listen("udp", addrL)
	require.NoError(t, err)

	for {
		conn, err := listener.Accept()
		require.NoError(t, err)
		go handler.ServeUDP(conn)
	}
}
