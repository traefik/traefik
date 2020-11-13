package tcp

import (
	"bytes"
	"fmt"
	"io"
	"net"
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

	proxy, err := NewProxy(":"+port, 10*time.Millisecond)
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

func TestLookupAddress(t *testing.T) {
	testCases := []struct {
		desc       string
		address    string
		expectSame bool
	}{
		{
			desc:       "IP doesn't need refresh",
			address:    "8.8.4.4:53",
			expectSame: true,
		},
		{
			desc:       "Hostname needs refresh",
			address:    "dns.google:53",
			expectSame: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			proxy, err := NewProxy(test.address, 10*time.Millisecond)
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

			require.NotNil(t, proxy.target)

			var lastTarget *net.TCPAddr

			for i := 0; i < 3; i++ {
				wg.Add(1)
				conn, err := net.Dial("tcp", proxyListener.Addr().String())
				require.NoError(t, err)

				_, err = conn.Write([]byte("ping\n"))
				require.NoError(t, err)

				err = conn.Close()
				require.NoError(t, err)
				wg.Wait()

				assert.NotNil(t, proxy.target)
				if lastTarget != nil {
					if test.expectSame {
						assert.Same(t, lastTarget, proxy.target)
					} else {
						assert.NotSame(t, lastTarget, proxy.target)
					}
				}

				lastTarget = proxy.target
			}
		})
	}
}
