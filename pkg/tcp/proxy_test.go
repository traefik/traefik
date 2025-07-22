package tcp

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func fakeRedis(t *testing.T, listener net.Listener) {
	t.Helper()

	for {
		conn, err := listener.Accept()
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
					_ = conn.Close()
					return
				}
			}

			if withErr {
				_ = conn.Close()
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

	dialer := tcpDialer{&net.Dialer{}, 10 * time.Millisecond}

	proxy, err := NewProxy(":"+port, nil, dialer)
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

func TestProxyProtocol(t *testing.T) {
	testCases := []struct {
		desc    string
		version int
	}{
		{
			desc:    "PROXY protocol v1",
			version: 1,
		},
		{
			desc:    "PROXY protocol v2",
			version: 2,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			backendListener, err := net.Listen("tcp", ":0")
			require.NoError(t, err)

			var version int
			proxyBackendListener := proxyproto.Listener{
				Listener: backendListener,
				ValidateHeader: func(h *proxyproto.Header) error {
					version = int(h.Version)
					return nil
				},
				Policy: func(upstream net.Addr) (proxyproto.Policy, error) {
					switch test.version {
					case 1, 2:
						return proxyproto.USE, nil
					default:
						return proxyproto.REQUIRE, errors.New("unsupported version")
					}
				},
			}
			defer proxyBackendListener.Close()

			go fakeRedis(t, &proxyBackendListener)

			_, port, err := net.SplitHostPort(proxyBackendListener.Addr().String())
			require.NoError(t, err)

			dialer := tcpDialer{&net.Dialer{}, 10 * time.Millisecond}

			proxy, err := NewProxy(":"+port, &dynamic.ProxyProtocol{Version: test.version}, dialer)
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

			assert.Equal(t, int64(4), n)
			assert.Equal(t, "PONG", buffer.String())

			assert.Equal(t, test.version, version)
		})
	}
}
