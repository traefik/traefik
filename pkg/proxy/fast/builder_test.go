package fast

import (
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

type transportManagerWithTimeoutsMock struct {
	forwardingTimeouts *dynamic.ForwardingTimeouts
}

func (r *transportManagerWithTimeoutsMock) GetTLSConfig(_ string) (*tls.Config, error) {
	return nil, nil
}

func (r *transportManagerWithTimeoutsMock) Get(_ string) (*dynamic.ServersTransport, error) {
	return &dynamic.ServersTransport{ForwardingTimeouts: r.forwardingTimeouts}, nil
}

func TestProxyBuilder_ConnReadWriteTimeouts(t *testing.T) {
	testCases := []struct {
		desc               string
		forwardingTimeouts *dynamic.ForwardingTimeouts
		expectedWrapped    bool
	}{
		{
			desc: "no forwarding timeouts",
		},
		{
			desc:               "no read/write timeout set",
			forwardingTimeouts: &dynamic.ForwardingTimeouts{},
		},
		{
			desc: "read timeout set",
			forwardingTimeouts: &dynamic.ForwardingTimeouts{
				ReadTimeout: ptypes.Duration(50 * time.Millisecond),
			},
			expectedWrapped: true,
		},
		{
			desc: "write timeout set",
			forwardingTimeouts: &dynamic.ForwardingTimeouts{
				WriteTimeout: ptypes.Duration(50 * time.Millisecond),
			},
			expectedWrapped: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			t.Cleanup(func() { _ = ln.Close() })

			go func() {
				for {
					conn, err := ln.Accept()
					if err != nil {
						return
					}
					_ = conn.Close()
				}
			}()

			builder := NewProxyBuilder(&transportManagerWithTimeoutsMock{forwardingTimeouts: test.forwardingTimeouts}, static.FastProxyConfig{})

			cfg, err := builder.transportManager.Get("test")
			require.NoError(t, err)

			pool := builder.getPool("test", cfg, nil, testhelpers.MustParseURL("http://"+ln.Addr().String()), nil)

			conn, err := pool.dialer()
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })

			if test.expectedWrapped {
				require.IsType(t, &connWithTimeouts{}, conn)
			} else {
				require.IsType(t, &net.TCPConn{}, conn)
			}
		})
	}
}

func TestConnWithTimeouts(t *testing.T) {
	testCases := []struct {
		desc                      string
		readTimeout               time.Duration
		writeTimeout              time.Duration
		serverWriteDelay          time.Duration
		serverReads               bool
		expectedReadTimeoutError  bool
		expectedWriteTimeoutError bool
	}{
		{
			desc:                     "read timeout - server delays longer than client timeout",
			readTimeout:              50 * time.Millisecond,
			serverWriteDelay:         150 * time.Millisecond,
			expectedReadTimeoutError: true,
		},
		{
			desc:             "read succeeds with sufficient timeout",
			readTimeout:      500 * time.Millisecond,
			serverWriteDelay: 100 * time.Millisecond,
		},
		{
			desc:                      "write timeout triggered when reader stops reading",
			writeTimeout:              50 * time.Millisecond,
			expectedWriteTimeoutError: true,
		},
		{
			desc:         "write succeeds within timeout",
			writeTimeout: 500 * time.Millisecond,
			serverReads:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// net.Pipe has no OS buffering: reads and writes block until the other side is ready,
			// which allows the read and write deadlines to be exercised deterministically.
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()

			serverDone := make(chan struct{})
			go func() {
				defer close(serverDone)
				_, _ = server.Write([]byte("HELLO1"))
				if test.serverWriteDelay > 0 {
					time.Sleep(test.serverWriteDelay)
				}
				_, _ = server.Write([]byte("HELLO2"))
				if test.serverReads {
					buf := make([]byte, 5)
					_, _ = server.Read(buf)
				}
			}()

			conn := &connWithTimeouts{
				Conn:         client,
				readTimeout:  test.readTimeout,
				writeTimeout: test.writeTimeout,
			}

			buf := make([]byte, 6)
			_, err := conn.Read(buf)
			require.NoError(t, err)
			require.Equal(t, "HELLO1", string(buf))

			_, err = conn.Read(buf)
			if test.expectedReadTimeoutError {
				var netErr net.Error
				require.ErrorAs(t, err, &netErr)
				require.True(t, netErr.Timeout())
				client.Close()
				<-serverDone
				return
			}
			require.NoError(t, err)
			require.Equal(t, "HELLO2", string(buf))

			if test.serverReads || test.expectedWriteTimeoutError {
				_, err = conn.Write([]byte("HI"))
				if test.expectedWriteTimeoutError {
					var netErr net.Error
					require.ErrorAs(t, err, &netErr)
					require.True(t, netErr.Timeout())
				} else {
					require.NoError(t, err)
				}
			}

			client.Close()
			<-serverDone
		})
	}
}
