package server

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
)

func TestBuildListenerSocketActivation(t *testing.T) {
	// Unix sockets have a per-platform sun_path length limit (104 bytes on
	// Darwin) so t.TempDir is too long; place the socket under /tmp.
	dir, err := os.MkdirTemp("/tmp", "traefik-sa-test") //nolint:usetesting // Keep the socket path below the Unix socket path length limit.
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	testCases := []struct {
		desc          string
		network       string
		address       string
		expectedError string
	}{
		{
			desc:          "rejects Unix listener",
			network:       "unix",
			address:       filepath.Join(dir, "test.sock"),
			expectedError: "listener type *net.UnixListener is not supported for TCP entrypoints",
		},
		{
			desc:    "preserves TCP listener",
			network: "tcp",
			address: "127.0.0.1:0",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			listener, err := net.Listen(test.network, test.address)
			require.NoError(t, err)
			t.Cleanup(func() { _ = listener.Close() })

			prev := socketActivation
			socketActivation = &SocketActivation{
				enabled:   true,
				listeners: map[string]net.Listener{"web": listener},
			}
			t.Cleanup(func() { socketActivation = prev })

			ln, err := buildListener(t.Context(), "web", &static.EntryPoint{Address: ":0"})
			if test.expectedError != "" {
				require.EqualError(t, err, test.expectedError)
				return
			}
			require.NoError(t, err)
			t.Cleanup(func() { _ = ln.Close() })

			require.IsType(t, &onceCloseListener{}, ln)
			wrappedListener := ln.(*onceCloseListener).Listener
			require.IsType(t, tcpKeepAliveListener{}, wrappedListener)
			assert.Same(t, listener, wrappedListener.(tcpKeepAliveListener).TCPListener)
		})
	}
}
