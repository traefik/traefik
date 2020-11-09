package tcp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
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

	proxy, err := NewProxy(":"+port, 10*time.Millisecond, &dynamic.ProxyProtocol{Version: ""})
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
	testCases := []string{"", "1", "2"}
	for _, proxyversion := range testCases {
		func(proxyprotocolversion string) {
			backendListener, err := net.Listen("tcp", ":0")
			require.NoError(t, err)
			proxybackendListener := proxyproto.Listener{
				Listener: backendListener,
				ValidateHeader: func(h *proxyproto.Header) error {
					versionFromHeader := int(h.Version)
					expectVersion, _ := strconv.Atoi(proxyprotocolversion)
					if versionFromHeader != expectVersion {
						t.Fatalf("Expected version %s got: %d", proxyprotocolversion, versionFromHeader)
					}
					return nil
				},
				Policy: func(upstream net.Addr) (proxyproto.Policy, error) {
					switch proxyprotocolversion {
					case "1", "2":
						return proxyproto.USE, nil
					case "":
						return proxyproto.IGNORE, nil
					default:
						return proxyproto.REQUIRE, fmt.Errorf("fail")
					}
				},
			}
			defer proxybackendListener.Close()

			go fakeRedis(t, &proxybackendListener)
			_, port, err := net.SplitHostPort(proxybackendListener.Addr().String())
			require.NoError(t, err)

			proxy, err := NewProxy(":"+port, 10*time.Millisecond, &dynamic.ProxyProtocol{Version: proxyprotocolversion})
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
		}(proxyversion)
	}
}
