package plugins

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTCPPluginFunctional tests that a TCP yaegi plugin actually works end-to-end.
func TestTCPPluginFunctional(t *testing.T) {
	ctx := t.Context()
	goPath := "fixtures"

	// Load the test TCP plugin
	manifest := &Manifest{
		Import:    "testplugintcp",
		UseUnsafe: false,
	}
	settings := Settings{
		UseUnsafe: false,
	}

	interpreter, err := newInterpreter(ctx, goPath, manifest, settings)
	require.NoError(t, err)

	builder, err := newYaegiMiddlewareBuilder(interpreter, "", "testplugintcp")
	require.NoError(t, err)

	// Test HTTP handler functionality
	t.Run("HTTP plugin actually works", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true
			rw.WriteHeader(http.StatusOK)
		})

		middleware, err := builder.newMiddleware(map[string]interface{}{
			"headerValue": "test-value",
		}, "test")
		require.NoError(t, err)

		handler, err := middleware.NewHandler(ctx, next)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		rw := httptest.NewRecorder()

		// Actually call the plugin!
		handler.ServeHTTP(rw, req)

		assert.True(t, nextCalled, "Plugin should call next handler")
		assert.Equal(t, "test-value", rw.Header().Get("X-Test-Plugin-TCP"), "Plugin should set header")
	})

	// Test TCP plugin detection and compilation
	t.Run("TCP plugin is detected and compiles", func(t *testing.T) {
		middleware, err := builder.newMiddleware(map[string]interface{}{
			"ipPrefix": "test-",
		}, "test-tcp")
		require.NoError(t, err)

		// Verify middleware is a YaegiMiddleware
		yaeg, ok := middleware.(*YaegiMiddleware)
		require.True(t, ok, "Middleware should be a YaegiMiddleware")

		// Verify that NewTCP function was detected
		assert.True(t, yaeg.builder.fnNewTCP.IsValid(), "NewTCP function should be detected in plugin")

		// Try to get TCP constructor - this proves the function exists and is callable
		constructor, err := yaeg.NewTCPHandler(ctx, nil)
		require.NoError(t, err, "Should be able to get TCP constructor from plugin with NewTCP function")
		require.NotNil(t, constructor, "TCP constructor should not be nil")
	})
}
