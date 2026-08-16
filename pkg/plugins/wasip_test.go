//go:build linux || darwin

package plugins

import (
	"bytes"
	"context"
	_ "embed"
	"net"
	"strconv"
	"testing"

	"github.com/samyfodil/wazy"
	"github.com/stretchr/testify/require"
)

//go:embed fixtures/withsocket/sockguest.wasm
var sockGuestWasm []byte

// TestInstantiateHost_socketsExtension proves the actual integration point
// Traefik calls (InstantiateHost) works end to end against a real,
// independently-built guest that uses the WasmEdge sockets extension - not
// just that nothing else broke.
func TestInstantiateHost_socketsExtension(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	served := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			served <- "accept: " + err.Error()
			return
		}
		defer conn.Close()
		buf := make([]byte, 64)
		n, err := conn.Read(buf)
		if err != nil {
			served <- "read: " + err.Error()
			return
		}
		served <- string(buf[:n])
		_, _ = conn.Write([]byte("pong"))
	}()

	ctx := context.Background()
	rt := wazy.NewRuntimeWithConfig(ctx, wazy.NewRuntimeConfig())
	defer rt.Close(ctx)

	guestModule, err := rt.CompileModule(ctx, sockGuestWasm)
	require.NoError(t, err)

	applyCtx, err := InstantiateHost(ctx, rt, guestModule, Settings{})
	require.NoError(t, err)

	port := ln.Addr().(*net.TCPAddr).Port
	var stdout bytes.Buffer
	config := wazy.NewModuleConfig().
		WithArgs("sockguest", "client", strconv.Itoa(port)).
		WithStdout(&stdout).
		WithStderr(&stdout)

	mod, err := rt.InstantiateModule(applyCtx(ctx), guestModule, config)
	require.NoError(t, err)
	require.NoError(t, mod.Close(ctx))

	require.Equal(t, "ping", <-served, "the host listener should have received the guest's bytes")
	require.Contains(t, stdout.String(), "sent 4")
	require.Contains(t, stdout.String(), "recv pong")
}
