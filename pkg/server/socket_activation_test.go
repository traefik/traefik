package server

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
)

// withSocketActivation temporarily replaces the package-level
// socketActivation with the given value for the duration of the test.
func withSocketActivation(t *testing.T, sa *SocketActivation) {
	t.Helper()
	prev := socketActivation
	socketActivation = sa
	t.Cleanup(func() { socketActivation = prev })
}

// TestBuildListenerSocketActivationUnix verifies that buildListener
// returns a clear error instead of panicking when systemd socket
// activation provides a *net.UnixListener for an entrypoint that
// expects *net.TCPListener. Regression test for #10924.
func TestBuildListenerSocketActivationUnix(t *testing.T) {
	// Unix sockets have a per-platform sun_path length limit (104 bytes on
	// Darwin) so t.TempDir is too long; place the socket under /tmp.
	dir, err := os.MkdirTemp("/tmp", "traefik-sa-test")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	addr, err := net.ResolveUnixAddr("unix", filepath.Join(dir, "test.sock"))
	require.NoError(t, err)

	unixLn, err := net.ListenUnix("unix", addr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = unixLn.Close() })

	withSocketActivation(t, &SocketActivation{
		enabled:   true,
		listeners: map[string]net.Listener{"web": unixLn},
	})

	_, err = buildListener(context.Background(), "web", &static.EntryPoint{Address: ":0"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "socket activation listener for entrypoint \"web\"")
	assert.Contains(t, err.Error(), "*net.UnixListener")
}

// TestBuildListenerSocketActivationTCP confirms the happy path: a
// *net.TCPListener provided through socket activation is wrapped in
// tcpKeepAliveListener and buildListener succeeds.
func TestBuildListenerSocketActivationTCP(t *testing.T) {
	tcpLn, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	require.NoError(t, err)
	t.Cleanup(func() { _ = tcpLn.Close() })

	withSocketActivation(t, &SocketActivation{
		enabled:   true,
		listeners: map[string]net.Listener{"web": tcpLn},
	})

	ln, err := buildListener(context.Background(), "web", &static.EntryPoint{Address: ":0"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
}
