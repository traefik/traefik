package server

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/udp"
)

func TestShutdownUDPConn(t *testing.T) {
	ep := static.EntryPoint{
		Address: ":0",
		Transport: &static.EntryPointsTransport{
			LifeCycle: &static.LifeCycle{
				GraceTimeOut: ptypes.Duration(5 * time.Second),
			},
		},
	}
	ep.SetDefaults()

	entryPoint, err := NewUDPEntryPoint(&ep)
	require.NoError(t, err)

	go entryPoint.Start(context.Background())
	entryPoint.Switch(udp.HandlerFunc(func(conn *udp.Conn) {
		for {
			b := make([]byte, 1024*1024)
			n, err := conn.Read(b)
			if err != nil {
				return
			}

			// We control the termination, otherwise we would block on the Read above,
			// until conn is closed by a timeout.
			// Which means we would get an error,
			// and even though we are in a goroutine and the current test might be over,
			// go test would still yell at us if this happens while other tests are still running.
			if string(b[:n]) == "CLOSE" {
				return
			}
			_, _ = conn.Write(b[:n])
		}
	}))

	conn, err := net.Dial("udp", entryPoint.listener.Addr().String())
	require.NoError(t, err)

	// Start sending packets, to create a "session" with the server.
	requireEcho(t, "TEST", conn, time.Second)

	doneChan := make(chan struct{})
	go func() {
		entryPoint.Shutdown(context.Background())
		close(doneChan)
	}()

	// Make sure that our session is still live even after the shutdown.
	requireEcho(t, "TEST2", conn, time.Second)

	// And make sure that on the other hand, opening new sessions is not possible anymore.
	conn2, err := net.Dial("udp", entryPoint.listener.Addr().String())
	require.NoError(t, err)

	_, err = conn2.Write([]byte("TEST"))
	// Packet is accepted, but dropped
	require.NoError(t, err)

	// Make sure that our session is yet again still live.
	// This is specifically to make sure we don't create a regression in listener's readLoop,
	// i.e. that we only terminate the listener's readLoop goroutine by closing its pConn.
	requireEcho(t, "TEST3", conn, time.Second)

	done := make(chan bool)
	go func() {
		defer close(done)
		b := make([]byte, 1024*1024)
		n, err := conn2.Read(b)
		require.Error(t, err)
		require.Equal(t, 0, n)
	}()

	conn2.Close()

	select {
	case <-done:
	case <-time.Tick(time.Second):
		t.Fatal("Timeout")
	}

	_, err = conn.Write([]byte("CLOSE"))
	require.NoError(t, err)

	select {
	case <-doneChan:
	case <-time.Tick(10 * time.Second):
		// In case we introduce a regression that would make the test wait forever.
		t.Fatal("Timeout during shutdown")
	}
}

// requireEcho tests that conn session is live and functional,
// by writing data through it,
// and expecting the same data as a response when reading on it.
// It fatals if the read blocks longer than timeout,
// which is useful to detect regressions that would make a test wait forever.
func requireEcho(t *testing.T, data string, conn io.ReadWriter, timeout time.Duration) {
	t.Helper()

	_, err := conn.Write([]byte(data))
	require.NoError(t, err)

	doneChan := make(chan struct{})
	go func() {
		b := make([]byte, 1024*1024)
		n, err := conn.Read(b)
		require.NoError(t, err)
		require.Equal(t, data, string(b[:n]))
		close(doneChan)
	}()

	select {
	case <-doneChan:
	case <-time.Tick(timeout):
		t.Fatalf("Timeout during echo for: %s", data)
	}
}
