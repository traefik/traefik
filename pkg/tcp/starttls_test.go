package tcp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/log"
)

var (
	testmessage  = "message8"
	testresponse = "response"
)

func fakeBackend(t *testing.T, listener net.Listener, withStartTLS bool) {
	t.Helper()
	for {
		conn, err := listener.Accept()
		fmt.Println("Accept on server")
		require.NoError(t, err)
		for {
			if withStartTLS {
				withErr := false
				buf := make([]byte, len(PostgresStartTLSMsg))
				log.Info("backend: reading start tls ")
				if _, err := conn.Read(buf); err != nil {
					withErr = true
				}

				if bytes.Equal(buf, PostgresStartTLSMsg) {
					time.Sleep(1 * time.Millisecond)
					log.Info("backend: writing start tls ")
					if _, err := conn.Write(PostgresStartTLSReply); err != nil {
						conn.Close()
						return
					}
				} else {
					conn.Close()
					return
				}
				if withErr {
					conn.Close()
					return
				}
			}

			log.Info("backend: reading message")
			buf := make([]byte, len([]byte(testmessage)))
			_, err := io.ReadFull(conn, buf)
			require.NoError(t, err)

			if !bytes.Equal(buf, []byte(testmessage)) {
				conn.Close()
				t.Logf("backend: received bad message: want %s, got %s", testmessage, string(buf))
				return
			}

			log.Info("backend: writing response")
			_, err = conn.Write([]byte(testresponse))
			require.NoError(t, err)
		}
	}
}

// In this test we test the the proxy's configuration against different backends.
// The proxy must match the the backend. I.e. if the backend expects StartTLS we have
// to enable StartTLS in the proxy. In this case the client does not do any StartTLS
// handshake, because this side is handled by the tcp.Router and not the Proxy.
func TestProxyPostgresStartTLS(t *testing.T) {
	tests := []struct {
		name                 string
		backendWantsStartTLS bool
		proxyDoesStartTLS    bool
		wantErr              bool
	}{
		{
			name:                 "proxy and backend want starttls",
			backendWantsStartTLS: true,
			proxyDoesStartTLS:    true,
			wantErr:              false,
		},
		{
			name:                 "neither proxy nor backend want starttls",
			backendWantsStartTLS: false,
			proxyDoesStartTLS:    false,
			wantErr:              false,
		},
		{
			name:                 "proxy does starttls, backend w/o starttls",
			backendWantsStartTLS: false,
			proxyDoesStartTLS:    true,
			wantErr:              true,
		},
		{
			name:                 "proxy w/o starttls, backend wants starttls",
			backendWantsStartTLS: true,
			proxyDoesStartTLS:    false,
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backendListener, err := net.Listen("tcp", ":0")
			require.NoError(t, err)

			go fakeBackend(t, backendListener, tt.backendWantsStartTLS)
			_, port, err := net.SplitHostPort(backendListener.Addr().String())
			require.NoError(t, err)

			startTLS := ""
			if tt.proxyDoesStartTLS {
				startTLS = "postgres"
			}
			proxy, err := NewProxy(":"+port, 10*time.Millisecond, nil, startTLS)
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

			log.Info("client: writing message")
			_, err = io.WriteString(conn, testmessage)
			require.NoError(t, err)

			log.Info("client: reading response")
			buf := make([]byte, len([]byte(testresponse)))
			_, err = io.ReadFull(conn, buf)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected: want err: %v, got err: %v", tt.wantErr, err)
			}

			if (string(buf) != testresponse) && !tt.wantErr {
				t.Fatalf("bad response: got %v, want %v", string(buf), testresponse)
			}

			if (string(buf) == testresponse) && tt.wantErr {
				t.Fatalf("got correct response but expected something else")
			}
		})
	}
}
