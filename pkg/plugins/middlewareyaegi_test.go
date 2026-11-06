package plugins

import (
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/yaegi/interp"
)

// TestNewInterpreter_SyscallErrorCase - Tests the security gate logic
func TestNewInterpreter_SyscallErrorCase(t *testing.T) {
	manifest := &Manifest{
		Import:    "does-not-matter-will-error-before-import",
		UseUnsafe: true, // Plugin wants unsafe access
	}
	settings := Settings{
		UseUnsafe: false, // But admin doesn't allow it
	}

	ctx := t.Context()
	_, err := newInterpreter(ctx, "/tmp", manifest, settings)

	// This proves our security gate logic works
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restricted imports", "Our error message should be returned")
}

// TestNewYaegiMiddlewareBuilder_WithSyscallSupport - Tests the ACTUAL production code!
func TestNewYaegiMiddlewareBuilder_WithSyscallSupport(t *testing.T) {
	tests := []struct {
		name           string
		pluginType     string
		manifestUnsafe bool
		settingsUnsafe bool
		shouldSucceed  bool
		expectedError  string
	}{
		{
			name:           "Should work with safe plugin when useUnsafe disabled",
			pluginType:     "safe",
			manifestUnsafe: false,
			settingsUnsafe: false,
			shouldSucceed:  true,
		},
		{
			name:           "Should work with unsafe-only plugin when useUnsafe enabled",
			pluginType:     "unsafe-only",
			manifestUnsafe: true,
			settingsUnsafe: true,
			shouldSucceed:  true,
		},
		{
			name:           "Should work with unsafe+syscall plugin when useUnsafe enabled",
			pluginType:     "unsafe+syscall",
			manifestUnsafe: true,
			settingsUnsafe: true,
			shouldSucceed:  true,
		},
		{
			name:           "Should fail when plugin needs unsafe but setting disabled",
			pluginType:     "unsafe-only",
			manifestUnsafe: true,
			settingsUnsafe: false,
			shouldSucceed:  false,
			expectedError:  "restricted imports",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()

			// Set GOPATH to include our fixtures directory
			goPath := "fixtures"

			// Create interpreter using our ACTUAL newInterpreter function
			// This will automatically import the real test plugin!
			interpreter, err := createInterpreterForTesting(ctx, goPath, tc.pluginType, tc.manifestUnsafe, tc.settingsUnsafe)

			if tc.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, interpreter)

				// Test actual middleware building using newYaegiMiddlewareBuilder
				// The plugin is already loaded by newInterpreter!
				basePkg := getPluginPackage(tc.pluginType)

				builder, err := newYaegiMiddlewareBuilder(interpreter, basePkg, basePkg)
				require.NoError(t, err)
				require.NotNil(t, builder)

				// Verify that unsafe/syscall functions actually work if the plugin uses them
				if tc.pluginType != "safe" {
					verifyMiddlewareWorks(t, builder)
				}
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

// Helper that uses the ACTUAL newInterpreter function with real test plugins
func createInterpreterForTesting(ctx context.Context, goPath, pluginType string, manifestUnsafe, settingsUnsafe bool) (*interp.Interpreter, error) {
	pluginImport := getPluginPackage(pluginType)

	manifest := &Manifest{
		Import:    pluginImport,
		UseUnsafe: manifestUnsafe,
	}
	settings := Settings{
		UseUnsafe: settingsUnsafe,
	}

	// Call the ACTUAL production newInterpreter function - no workarounds needed!
	return newInterpreter(ctx, goPath, manifest, settings)
}

// Helper to get the correct plugin package name based on type
func getPluginPackage(pluginType string) string {
	switch pluginType {
	case "safe":
		return "testpluginsafe"
	case "unsafe-only":
		return "testpluginunsafe"
	case "unsafe+syscall":
		return "testpluginsyscall"
	default:
		return "testpluginsafe"
	}
}

// Helper to verify that unsafe/syscall functions actually work by invoking the middleware
func verifyMiddlewareWorks(t *testing.T, builder *yaegiMiddlewareBuilder) {
	t.Helper()
	// Create a middleware instance - this will call the plugin's New() function
	// which uses unsafe/syscall, proving they work
	middleware, err := builder.newMiddleware(map[string]any{
		"message": "test",
	}, "test-middleware")
	require.NoError(t, err, "Should be able to create middleware that uses unsafe/syscall")
	require.NotNil(t, middleware, "Middleware should not be nil")

	// The fact that we got here without crashing proves unsafe/syscall work!
}

func TestBuilder_BuildTCP(t *testing.T) {
	testCases := []struct {
		desc            string
		config          map[string]any
		wantBanner      string
		wantOriginal    bool
		wantCloseWrites int
	}{
		{
			desc:            "connection passed as is",
			config:          map[string]any{"banner": "hello"},
			wantBanner:      "hello",
			wantOriginal:    true,
			wantCloseWrites: 1,
		},
		{
			desc:            "connection wrapped by the plugin",
			config:          map[string]any{"banner": "hello", "wrap": true},
			wantBanner:      "hello",
			wantCloseWrites: 1,
		},
	}

	builder := newTCPTestBuilder(t)

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			plug, err := builder.BuildTCP("tcp", test.config, "test-tcp")
			require.NoError(t, err)

			var got tcp.WriteCloser
			next := tcp.HandlerFunc(func(conn tcp.WriteCloser) {
				got = conn
			})

			handler, err := plug(t.Context(), next)
			require.NoError(t, err)

			conn := &fakeWriteCloser{}
			handler.ServeTCP(conn)

			assert.Equal(t, test.wantBanner, conn.written.String())
			require.NotNil(t, got)
			if test.wantOriginal {
				assert.Same(t, conn, got)
			} else {
				assert.NotSame(t, conn, got)
			}

			require.NoError(t, got.CloseWrite())
			assert.Equal(t, test.wantCloseWrites, conn.closeWrites)
		})
	}
}

func TestBuilder_BuildTCP_errors(t *testing.T) {
	builder := newTCPTestBuilder(t)

	_, err := builder.BuildTCP("unknown", nil, "test-tcp")
	require.ErrorContains(t, err, "unknown plugin type: unknown")

	plug, err := builder.BuildTCP("httponly", nil, "test-tcp")
	require.NoError(t, err)

	_, err = plug(t.Context(), tcp.HandlerFunc(func(tcp.WriteCloser) {}))
	require.ErrorContains(t, err, "does not export a NewTCP function")
}

func newTCPTestBuilder(t *testing.T) Builder {
	t.Helper()

	builders := map[string]middlewareBuilder{}
	for name, pkg := range map[string]string{"tcp": "testplugintcp", "httponly": "testpluginsafe"} {
		i, err := newInterpreter(t.Context(), "fixtures", &Manifest{Import: pkg}, Settings{})
		require.NoError(t, err)

		b, err := newYaegiMiddlewareBuilder(i, pkg, pkg)
		require.NoError(t, err)

		builders[name] = b
	}

	return Builder{middlewareBuilders: builders}
}

type fakeWriteCloser struct {
	net.Conn

	written     bytes.Buffer
	closeWrites int
}

func (c *fakeWriteCloser) Write(b []byte) (int, error) {
	return c.written.Write(b)
}

func (c *fakeWriteCloser) CloseWrite() error {
	c.closeWrites++
	return nil
}
