package plugins

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// TestCombinedPlugin tests a plugin that supports BOTH HTTP and TCP protocols simultaneously.
func TestCombinedPlugin(t *testing.T) {
	ctx := t.Context()
	goPath := "fixtures"

	// Load the combined test plugin
	manifest := &Manifest{
		Import:      "testplugincombined",
		UseUnsafe:   false,
		SupportsTCP: true, // Indicates this plugin supports TCP
	}
	settings := Settings{
		UseUnsafe: false,
	}

	interpreter, err := newInterpreter(ctx, goPath, manifest, settings)
	require.NoError(t, err, "Should load combined plugin")

	builder, err := newYaegiMiddlewareBuilder(interpreter, "", "testplugincombined")
	require.NoError(t, err, "Should create builder for combined plugin")

	// Test 1: HTTP handler works
	t.Run("HTTP handler filters by IP and sets headers", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true
			rw.WriteHeader(http.StatusOK)
		})

		middleware, err := builder.newMiddleware(map[string]interface{}{
			"allowedIPPrefix": "127",
			"httpHeaderValue": "test-combined",
		}, "test")
		require.NoError(t, err)

		handler, err := middleware.NewHandler(ctx, next)
		require.NoError(t, err)

		// Test with allowed IP (127.x.x.x) - should set metadata headers
		t.Run("allows 127.x IPs and sets metadata", func(t *testing.T) {
			nextCalled = false
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			req.RemoteAddr = "127.0.0.1:1234"
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)

			// Verify request was allowed
			assert.Equal(t, http.StatusOK, rw.Code, "Should allow 127.x IP")
			assert.True(t, nextCalled, "Should call next handler for allowed IP")

			// Verify metadata headers were set
			assert.Equal(t, "test-combined", rw.Header().Get("X-Combined-Plugin"),
				"Should set X-Combined-Plugin header with configured value")
			assert.Equal(t, "127.0.0.1", rw.Header().Get("X-Client-IP"),
				"Should set X-Client-IP header with extracted client IP")
		})

		// Test with blocked IP (192.x.x.x) - should reject and not set headers
		t.Run("blocks 192.x IPs and rejects request", func(t *testing.T) {
			nextCalled = false
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			req.RemoteAddr = "192.168.1.1:1234"
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)

			// Verify request was blocked
			assert.Equal(t, http.StatusForbidden, rw.Code, "Should block 192.x IP with 403")
			assert.False(t, nextCalled, "Should not call next handler for blocked IP")

			// Verify no metadata headers were set (rejected before setting)
			assert.Empty(t, rw.Header().Get("X-Combined-Plugin"),
				"Should not set metadata headers for rejected requests")
		})

		// Test with different allowed IP prefix
		t.Run("allows IPs matching configured prefix", func(t *testing.T) {
			middleware2, err := builder.newMiddleware(map[string]interface{}{
				"allowedIPPrefix": "10",
				"httpHeaderValue": "prefix-test",
			}, "test-prefix")
			require.NoError(t, err)

			handler2, err := middleware2.NewHandler(ctx, next)
			require.NoError(t, err)

			nextCalled = false
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			rw := httptest.NewRecorder()

			handler2.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code, "Should allow 10.x IP")
			assert.True(t, nextCalled, "Should call next handler")
			assert.Equal(t, "prefix-test", rw.Header().Get("X-Combined-Plugin"),
				"Should set header with configured value")
		})
	})

	// Test 2: TCP support is detected
	t.Run("TCP support is detected", func(t *testing.T) {
		middleware, err := builder.newMiddleware(map[string]interface{}{
			"allowedIPPrefix": "127",
		}, "test-tcp")
		require.NoError(t, err)

		yaeg, ok := middleware.(*YaegiMiddleware)
		require.True(t, ok, "Should be YaegiMiddleware")

		// Verify NewTCP function was found
		assert.True(t, yaeg.builder.fnNewTCP.IsValid(),
			"NewTCP function should be detected in combined plugin")

		// Verify we can create TCP constructor
		constructor, err := yaeg.NewTCPHandler(ctx, nil)
		require.NoError(t, err, "Should create TCP constructor without error")
		require.NotNil(t, constructor, "TCP constructor should not be nil")
	})

	// Test 3: TCP handler works (mirrors HTTP tests)
	t.Run("TCP handler filters by IP and sets context metadata", func(t *testing.T) {
		nextCalled := false
		var capturedCtx context.Context
		nextHandler := tcp.HandlerFunc(func(ctx context.Context, conn tcp.WriteCloser) {
			nextCalled = true
			capturedCtx = ctx
		})

		middleware, err := builder.newMiddleware(map[string]interface{}{
			"allowedIPPrefix": "127",
			"httpHeaderValue": "test-combined",
		}, "test")
		require.NoError(t, err)

		yaeg := middleware.(*YaegiMiddleware)
		tcpHandler, err := yaeg.builder.newTCPHandler(ctx, nextHandler, yaeg.config, "test")
		require.NoError(t, err)
		require.NotNil(t, tcpHandler)

		// Test with allowed IP (127.x.x.x) - should set metadata in context
		t.Run("allows 127.x IPs and sets metadata", func(t *testing.T) {
			nextCalled = false
			capturedCtx = nil
			testConn := &testTCPConn{remoteAddr: "127.0.0.1:1234"}

			// Initialize context with metadata map (like Traefik does)
			type metadataKey string
			metadata := make(map[string]string)
			execCtx := context.WithValue(t.Context(), metadataKey("metadata"), metadata)
			tcpHandler.ServeTCP(execCtx, testConn)

			// Verify request was allowed
			assert.True(t, nextCalled, "Should call next handler for allowed IP")
			require.NotNil(t, capturedCtx, "Context should be passed to next handler")

			// Verify metadata was set in the metadata map
			metadataMap, ok := capturedCtx.Value("metadata").(map[string]string)
			if !ok {
				type contextKey string
				metadataMap, ok = capturedCtx.Value(contextKey("metadata")).(map[string]string)
			}
			require.True(t, ok, "Context should contain metadata map")
			assert.Equal(t, "test-combined", metadataMap["X-Combined-Plugin"],
				"Should set X-Combined-Plugin in context metadata map")
			assert.Equal(t, "127.0.0.1", metadataMap["X-Client-IP"],
				"Should set X-Client-IP in context metadata map with extracted client IP")
		})

		// Test with blocked IP (192.x.x.x) - should close connection
		t.Run("blocks 192.x IPs and closes connection", func(t *testing.T) {
			nextCalled = false
			capturedCtx = nil
			testConn := &testTCPConn{remoteAddr: "192.168.1.1:1234"}

			tcpHandler.ServeTCP(t.Context(), testConn)

			// Verify connection was closed (rejected)
			assert.True(t, testConn.closed, "Connection should be closed for blocked IP")
			assert.False(t, nextCalled, "Should not call next handler for blocked IP")
		})

		// Test with different allowed IP prefix
		t.Run("allows IPs matching configured prefix", func(t *testing.T) {
			middleware2, err := builder.newMiddleware(map[string]interface{}{
				"allowedIPPrefix": "10",
				"httpHeaderValue": "prefix-test",
			}, "test-prefix")
			require.NoError(t, err)

			yaeg2 := middleware2.(*YaegiMiddleware)
			tcpHandler2, err := yaeg2.builder.newTCPHandler(ctx, nextHandler, yaeg2.config, "test-prefix")
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
			assert.Equal(t, "prefix-test", metadataMap["X-Combined-Plugin"],
				"Should set metadata with configured value")
		})
	})

	// Test 4: TCP support is detected
	t.Run("TCP support is detected", func(t *testing.T) {
		middleware, err := builder.newMiddleware(map[string]interface{}{
			"allowedIPPrefix": "127",
		}, "test-tcp")
		require.NoError(t, err)

		yaeg, ok := middleware.(*YaegiMiddleware)
		require.True(t, ok, "Should be YaegiMiddleware")

		// Verify NewTCP function was found
		assert.True(t, yaeg.builder.fnNewTCP.IsValid(),
			"NewTCP function should be detected in combined plugin")

		// Verify we can create TCP constructor
		constructor, err := yaeg.NewTCPHandler(ctx, nil)
		require.NoError(t, err, "Should create TCP constructor without error")
		require.NotNil(t, constructor, "TCP constructor should not be nil")
	})

	// Test 5: Verify both use same config
	t.Run("HTTP and TCP share same configuration", func(t *testing.T) {
		sharedConfig := map[string]interface{}{
			"allowedIPPrefix": "10",
			"httpHeaderValue": "shared-config",
		}

		// Create HTTP middleware
		httpMW, err := builder.newMiddleware(sharedConfig, "http-test")
		require.NoError(t, err)

		// Create TCP middleware (same config)
		tcpMW, err := builder.newMiddleware(sharedConfig, "tcp-test")
		require.NoError(t, err)

		// Both should work with the same config
		httpYaeg := httpMW.(*YaegiMiddleware)
		tcpYaeg := tcpMW.(*YaegiMiddleware)

		// Verify both have valid configs (pointers will differ, but values should match)
		assert.True(t, httpYaeg.config.IsValid(), "HTTP config should be valid")
		assert.True(t, tcpYaeg.config.IsValid(), "TCP config should be valid")
	})
}
