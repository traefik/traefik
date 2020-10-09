package tcp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeRedis(t *testing.T, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		fmt.Println("Accept on server")
		require.NoError(t, err)
		for {
			withErr := false
			buf := make([]byte, 64)
			if _, err := conn.Read(buf); err != nil {
				withErr = true
			}

			if string(buf[:4]) == "ping" {
				time.Sleep(1 * time.Millisecond)
				if _, err := conn.Write([]byte("PONG")); err != nil {
					conn.Close()
					return
				}
			}
			if withErr {
				conn.Close()
				return
			}
		}
	}
}

func TestCloseWrite(t *testing.T) {
	backendListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go fakeRedis(t, backendListener)
	_, port, err := net.SplitHostPort(backendListener.Addr().String())
	require.NoError(t, err)

	proxy, err := NewProxy(":"+port, 10*time.Millisecond, 10*time.Second)
	require.NoError(t, err)

	proxyListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		for {
			conn, err := proxyListener.Accept()
			require.NoError(t, err)
			proxy.ServeTCP(conn.(*net.TCPConn))
		}
	}()

	_, port, err = net.SplitHostPort(proxyListener.Addr().String())
	require.NoError(t, err)

	conn, err := net.Dial("tcp", ":"+port)
	require.NoError(t, err)

	_, err = conn.Write([]byte("ping\n"))
	require.NoError(t, err)

	err = conn.(*net.TCPConn).CloseWrite()
	require.NoError(t, err)

	var buf []byte
	buffer := bytes.NewBuffer(buf)
	n, err := io.Copy(buffer, conn)
	require.NoError(t, err)
	require.Equal(t, int64(4), n)
	require.Equal(t, "PONG", buffer.String())
}

// Copy of internal/testenv.go
func mustHaveExternalNetwork(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skipf("skipping test: no external network on %s", runtime.GOOS)
	}
	if testing.Short() {
		t.Skipf("skipping test: no external network in -short mode")
	}
}

func TestLookupAddressCache(t *testing.T) {
	mustHaveExternalNetwork(t)

	cacheDuration := 50 * time.Millisecond
	proxy, err := NewProxy("dns.google:53", 10*time.Millisecond, cacheDuration)
	require.NoError(t, err)

	proxyListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	var wg sync.WaitGroup
	go func(wg *sync.WaitGroup) {
		for {
			conn, err := proxyListener.Accept()
			require.NoError(t, err)
			proxy.ServeTCP(conn.(*net.TCPConn))
			wg.Done()
		}
	}(&wg)

	require.Nil(t, proxy.target)

	wg.Add(1)
	conn, err := net.Dial("tcp", proxyListener.Addr().String())
	require.NoError(t, err)

	_, err = conn.Write([]byte("ping\n"))
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)
	wg.Wait()

	require.NotNil(t, proxy.target)

	firstLookup := proxy.resolvedAt
	lookupExpiry := firstLookup.Add(cacheDuration)
	for {
		wg.Add(1)
		conn, err := net.Dial("tcp", proxyListener.Addr().String())
		require.NoError(t, err)

		_, err = conn.Write([]byte("ping\n"))
		require.NoError(t, err)

		err = conn.Close()
		require.NoError(t, err)
		wg.Wait()

		lastLookup := proxy.resolvedAt
		lastTarget := proxy.target

		if !lastLookup.Equal(firstLookup) {
			assert.True(t, time.Now().After(lookupExpiry), "address should be updated")
			return
		}
		assert.Equal(t, lastTarget, proxy.target)
	}
}
