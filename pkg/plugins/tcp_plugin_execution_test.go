package plugins

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// TestTCPPluginActualExecution tests that TCP plugins ACTUALLY EXECUTE via yaegi.
func TestTCPPluginActualExecution(t *testing.T) {
	ctx := t.Context()
	goPath := "fixtures"

	manifest := &Manifest{
		Import:      "testplugincombined",
		UseUnsafe:   false,
		SupportsTCP: true,
	}
	settings := Settings{
		UseUnsafe: false,
	}

	interpreter, err := newInterpreter(ctx, goPath, manifest, settings)
	require.NoError(t, err)

	builder, err := newYaegiMiddlewareBuilder(interpreter, "", "testplugincombined")
	require.NoError(t, err)

	middleware, err := builder.newMiddleware(map[string]interface{}{
		"allowedIPPrefix": "127",
		"httpHeaderValue": "test",
	}, "test")
	require.NoError(t, err)

	// Test HTTP (sanity check - this definitely works)
	t.Run("HTTP executes correctly", func(t *testing.T) {
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
		})

		handler, err := middleware.NewHandler(ctx, next)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusOK, rw.Code, "HTTP should allow 127.x IP")
	})

	// Test TCP execution with compiled tcp.Handler
	t.Run("TCP executes with compiled next handler", func(t *testing.T) {
		yaeg := middleware.(*YaegiMiddleware)

		// Track if next handler was called and capture context
		nextCalled := false
		var capturedCtx context.Context
		nextHandler := tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
			nextCalled = true
			capturedCtx = ctx
		})

		// Build TCP handler via yaegi - uses stdlib types only (net.Conn + closeWrite callback)
		tcpHandler, err := yaeg.builder.newTCPHandler(ctx, nextHandler, yaeg.config, "test")
		require.NoError(t, err, "Should build TCP handler")
		require.NotNil(t, tcpHandler)

		// Test 1: Allowed IP with metadata
		t.Run("allows 127.x IPs and sets context metadata", func(t *testing.T) {
			nextCalled = false
			capturedCtx = nil
			testConn := &testTCPConn{remoteAddr: "127.0.0.1:1234"}

			// Execute! Call ServeTCP directly since it's a real tcp.Handler
			// Initialize context with metadata map (like Traefik does)
			type metadataKey string
			metadata := make(map[string]string)
			execCtx := context.WithValue(t.Context(), metadataKey("metadata"), metadata)
			tcpHandler.ServeTCP(execCtx, testConn)

			// Verify execution
			assert.True(t, nextCalled, "TCP plugin should have called next handler for allowed IP")
			require.NotNil(t, capturedCtx, "Context should be passed to next handler")

			// Verify metadata was set in the metadata map
			// Try string key first (what plugin uses), then typed key
			metadataMap, ok := capturedCtx.Value("metadata").(map[string]string)
			if !ok {
				type contextKey string
				metadataMap, ok = capturedCtx.Value(contextKey("metadata")).(map[string]string)
			}
			require.True(t, ok, "Context should contain metadata map")
			assert.Equal(t, "test", metadataMap["X-Combined-Plugin"],
				"Should set X-Combined-Plugin in context metadata map")
			assert.Equal(t, "127.0.0.1", metadataMap["X-Client-IP"],
				"Should set X-Client-IP in context metadata map with extracted IP")
		})

		// Test 2: Blocked IP - should close connection
		t.Run("blocks 192.x IPs and closes connection", func(t *testing.T) {
			nextCalled = false
			capturedCtx = nil
			testConn := &testTCPConn{remoteAddr: "192.168.1.1:1234"}

			// Execute with blocked IP
			tcpHandler.ServeTCP(t.Context(), testConn)

			// Verify connection was closed (rejected)
			assert.True(t, testConn.closed, "Connection should be closed for blocked IP")
			assert.False(t, nextCalled, "Should not call next handler for blocked IP")
		})

		// Test 3: Different IP prefix configuration
		t.Run("respects configured IP prefix", func(t *testing.T) {
			middleware2, err := builder.newMiddleware(map[string]interface{}{
				"allowedIPPrefix": "10",
				"httpHeaderValue": "tcp-prefix-test",
			}, "test-tcp-prefix")
			require.NoError(t, err)

			yaeg2 := middleware2.(*YaegiMiddleware)
			tcpHandler2, err := yaeg2.builder.newTCPHandler(ctx, nextHandler, yaeg2.config, "test-tcp-prefix")
			require.NoError(t, err)

			nextCalled = false
			capturedCtx = nil
			testConn := &testTCPConn{remoteAddr: "10.0.0.1:1234"}

			// Initialize context with metadata map
			type metadataKey string
			metadata := make(map[string]string)
			execCtx := context.WithValue(t.Context(), metadataKey("metadata"), metadata)
			tcpHandler2.ServeTCP(execCtx, testConn)

			assert.True(t, nextCalled, "Should allow 10.x IP")
			require.NotNil(t, capturedCtx, "Context should be passed")

			// Verify metadata with different config
			metadataMap, ok := capturedCtx.Value("metadata").(map[string]string)
			if !ok {
				type contextKey string
				metadataMap, ok = capturedCtx.Value(contextKey("metadata")).(map[string]string)
			}
			require.True(t, ok, "Context should contain metadata map")
			assert.Equal(t, "tcp-prefix-test", metadataMap["X-Combined-Plugin"],
				"Should set metadata with configured value")
		})
	})
}

// testTCPConn is a simple test connection
type testTCPConn struct {
	remoteAddr string
	closed     bool
}

func (t *testTCPConn) Read(b []byte) (n int, err error)          { return 0, nil }
func (t *testTCPConn) Write(b []byte) (n int, err error)         { return len(b), nil }
func (t *testTCPConn) Close() error                              { t.closed = true; return nil }
func (t *testTCPConn) CloseWrite() error                         { return nil }
func (t *testTCPConn) LocalAddr() net.Addr                       { return &testAddr{addr: "0.0.0.0:0"} }
func (t *testTCPConn) RemoteAddr() net.Addr                      { return &testAddr{addr: t.remoteAddr} }
func (t *testTCPConn) SetDeadline(deadline time.Time) error      { return nil }
func (t *testTCPConn) SetReadDeadline(deadline time.Time) error  { return nil }
func (t *testTCPConn) SetWriteDeadline(deadline time.Time) error { return nil }

type testAddr struct {
	addr string
}

func (t *testAddr) Network() string { return "tcp" }
func (t *testAddr) String() string  { return t.addr }
