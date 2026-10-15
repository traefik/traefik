package plugins

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestSettingsWithoutSocket(t *testing.T) {
	cache := wazero.NewCompilationCache()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx := log.Logger.WithContext(t.Context())

	t.Setenv("PLUGIN_TEST", "MY-TEST")
	t.Setenv("PLUGIN_TEST_B", "MY-TEST_B")

	testCases := []struct {
		desc        string
		getSettings func(t *testing.T) (Settings, map[string]any)
		expected    string
	}{
		{
			desc: "mounts path",
			getSettings: func(t *testing.T) (Settings, map[string]any) {
				t.Helper()

				tempDir := t.TempDir()
				filePath := path.Join(tempDir, "hello.txt")
				err := os.WriteFile(filePath, []byte("content_test"), 0o644)
				require.NoError(t, err)

				return Settings{Mounts: []string{tempDir}}, map[string]any{"file": filePath}
			},
			expected: "content_test",
		},
		{
			desc: "mounts src to dest",
			getSettings: func(t *testing.T) (Settings, map[string]any) {
				t.Helper()

				tempDir := t.TempDir()
				filePath := path.Join(tempDir, "hello.txt")
				err := os.WriteFile(filePath, []byte("content_test"), 0o644)
				require.NoError(t, err)

				return Settings{Mounts: []string{tempDir + ":/tmp"}}, map[string]any{"file": "/tmp/hello.txt"}
			},
			expected: "content_test",
		},
		{
			desc: "one env",
			getSettings: func(t *testing.T) (Settings, map[string]any) {
				t.Helper()

				envs := []string{"PLUGIN_TEST"}
				return Settings{Envs: envs}, map[string]any{
					"envs": envs,
				}
			},
			expected: "MY-TEST\n",
		},
		{
			desc: "two env",
			getSettings: func(t *testing.T) (Settings, map[string]any) {
				t.Helper()

				envs := []string{"PLUGIN_TEST", "PLUGIN_TEST_B"}
				return Settings{Envs: envs}, map[string]any{
					"envs": envs,
				}
			},
			expected: "MY-TEST\nMY-TEST_B\n",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			settings, config := test.getSettings(t)

			builder := &wasmMiddlewareBuilder{path: "./fixtures/withoutsocket/plugin.wasm", cache: cache, settings: settings}

			cfg := reflect.ValueOf(config)

			m, applyCtx, err := builder.buildMiddleware(ctx, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusTeapot)
			}), cfg, "test")
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(applyCtx(ctx), "GET", "/", http.NoBody)

			m.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)
			assert.Equal(t, test.expected, rw.Body.String())
		})
	}
}

// TestHTTPWasm_requiresWrapExternal documents why Wasm plugins are built with middlewares.WrapExternal.
// The http-wasm host asserts http.Flusher on the ResponseWriter it receives instead of using http.ResponseController,
// so the flush done after the guest wrote the response body does not reach a ResponseWriter which only implements Unwrap,
// as Traefik's wrappers do.
// If the "only implements Unwrap" case fails after bumping http-wasm/http-wasm-host-go, the host follows Unwrap
// and WrapExternal may no longer be needed in newHandler.
func TestHTTPWasm_requiresWrapExternal(t *testing.T) {
	testCases := []struct {
		desc          string
		unwrapOnly    bool
		expectFlushed bool
	}{
		{
			desc:          "ResponseWriter implementing Flush",
			expectFlushed: true,
		},
		{
			desc:       "ResponseWriter which only implements Unwrap",
			unwrapOnly: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			filePath := path.Join(tempDir, "hello.txt")
			err := os.WriteFile(filePath, []byte("content_test"), 0o644)
			require.NoError(t, err)

			ctx := log.Logger.WithContext(t.Context())

			builder := &wasmMiddlewareBuilder{
				path:     "./fixtures/withoutsocket/plugin.wasm",
				cache:    wazero.NewCompilationCache(),
				settings: Settings{Mounts: []string{tempDir}},
			}

			// The guest writes the response body itself and does not call the next handler.
			handler, applyCtx, err := builder.buildMiddleware(ctx, http.NotFoundHandler(), reflect.ValueOf(map[string]any{"file": filePath}), "test")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			var rw http.ResponseWriter = recorder
			if test.unwrapOnly {
				rw = &testhelpers.UnwrapOnlyResponseWriter{ResponseWriter: recorder}
			}

			handler.ServeHTTP(rw, httptest.NewRequestWithContext(applyCtx(ctx), http.MethodGet, "/", http.NoBody))

			assert.Equal(t, "content_test", recorder.Body.String())
			assert.Equal(t, test.expectFlushed, recorder.Flushed)
		})
	}
}
