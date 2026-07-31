package plugins

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
)

// TestWasmMiddleware_ConcurrentSockets is a regression test for
// https://github.com/traefik/traefik/issues/11629: intermittent TCP failures
// reported by a wasm plugin that dials an outbound TCP connection (e.g. to a
// Kafka broker) on every request.
//
// The guest fixture (./fixtures/withsocket) dials ECHO_ADDR, writes a
// request-scoped payload, reads back the echo, and reports a 409 if the
// echoed bytes don't match what it sent. Firing many of these concurrently
// checks whether the host ever hands one guest's socket data to another.
func TestWasmMiddleware_ConcurrentSockets(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skipf("wasm socket support is only wired up on linux/darwin (see wasip.go); on %s InstantiateHost is a no-op (wasip_mock.go)", runtime.GOOS)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	ctx := log.Logger.WithContext(t.Context())

	echoAddr := startEchoServer(t)
	t.Setenv("ECHO_ADDR", echoAddr)

	cache := wazero.NewCompilationCache()
	builder := &wasmMiddlewareBuilder{
		path:     "./fixtures/withsocket/plugin.wasm",
		cache:    cache,
		settings: Settings{Envs: []string{"ECHO_ADDR"}},
	}

	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusTeapot)
	})

	m, applyCtx, err := builder.buildMiddleware(ctx, next, reflect.ValueOf(map[string]any{}), "test")
	require.NoError(t, err)

	const requests = 100
	const concurrency = 20

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var mismatches []string

	for i := range requests {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			reqCtx, cleanup, err := applyCtx(ctx)
			if err != nil {
				mu.Lock()
				mismatches = append(mismatches, fmt.Sprintf("id=%d applyCtx error=%v", i, err))
				mu.Unlock()
				return
			}
			defer cleanup()

			req := httptest.NewRequestWithContext(reqCtx, http.MethodGet, fmt.Sprintf("/?id=%d", i), http.NoBody)
			rw := httptest.NewRecorder()

			m.ServeHTTP(rw, req)

			if rw.Code != http.StatusOK {
				mu.Lock()
				mismatches = append(mismatches, fmt.Sprintf("id=%d status=%d body=%s", i, rw.Code, rw.Body.String()))
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if len(mismatches) > 0 {
		t.Errorf("%d/%d requests got corrupted socket data instead of their own echo:\n%s", len(mismatches), requests, strings.Join(mismatches, "\n"))
	}
}

// startEchoServer starts a TCP server that reads whatever a connection sends
// within a short window and writes the same bytes back, after a small random
// delay. The delay widens the window during which many concurrent
// connections are in flight, to increase the odds of exposing any mixing of
// socket state on the host side.
func startEchoServer(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()

				_ = conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
				buf := make([]byte, 4096)
				n, err := conn.Read(buf)
				if err != nil && n == 0 {
					return
				}
				received := append([]byte(nil), buf[:n]...)

				time.Sleep(time.Duration(rand.Intn(80)) * time.Millisecond)

				_, _ = conn.Write(received)
			}()
		}
	}()

	return ln.Addr().String()
}
