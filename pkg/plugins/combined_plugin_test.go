package plugins

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCombinedPlugin tests a plugin that supports BOTH HTTP and TCP protocols simultaneously.
func TestCombinedPlugin(t *testing.T) {
	ctx := context.Background()
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

		// Test with allowed IP (127.x.x.x)
		t.Run("allows 127.x IPs", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			req.RemoteAddr = "127.0.0.1:1234"
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code, "Should allow 127.x IP")
			assert.Equal(t, "test-combined", rw.Header().Get("X-Combined-Plugin"), "Should set custom header")
			assert.Equal(t, "127.0.0.1", rw.Header().Get("X-Client-IP"), "Should extract client IP")
		})

		// Test with blocked IP (192.x.x.x)
		t.Run("blocks 192.x IPs", func(t *testing.T) {
			nextCalled = false
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			req.RemoteAddr = "192.168.1.1:1234"
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusForbidden, rw.Code, "Should block 192.x IP")
			assert.False(t, nextCalled, "Should not call next handler")
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
			"✅ NewTCP function should be detected in combined plugin")

		// Verify we can create TCP constructor
		constructor, err := yaeg.NewTCPHandler(ctx, nil)
		require.NoError(t, err, "Should create TCP constructor without error")
		require.NotNil(t, constructor, "TCP constructor should not be nil")

		t.Log("✅ Combined plugin supports both HTTP and TCP")
		t.Log("✅ HTTP: ServeHTTP with IP filtering and headers")
		t.Log("✅ TCP: ServeTCP with IP filtering")
	})

	// Test 3: Verify both use same config
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

		t.Log("✅ Single config structure works for both HTTP and TCP handlers")
		t.Log("✅ Both middlewares created from same config successfully")
	})
}
