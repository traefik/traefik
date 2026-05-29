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

				return Settings{Mounts: []string{
						tempDir,
					}}, map[string]any{
						"file": filePath,
					}
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

				return Settings{Mounts: []string{
						tempDir + ":/tmp",
					}}, map[string]any{
						"file": "/tmp/hello.txt",
					}
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

			builder := &wasmMiddlewareBuilder{
				path:      "./fixtures/withoutsocket/plugin.wasm",
				cache:     cache,
				settings:  settings,
				instances: &wasmInstanceCache{m: make(map[string]*cachedWasmInstance)},
			}

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

func newTestWasmBuilder(t *testing.T) *wasmMiddlewareBuilder {
	t.Helper()

	return &wasmMiddlewareBuilder{
		path:      "./fixtures/withoutsocket/plugin.wasm",
		cache:     wazero.NewCompilationCache(),
		settings:  Settings{},
		instances: &wasmInstanceCache{m: make(map[string]*cachedWasmInstance)},
	}
}

func testNextHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusTeapot)
	})
}

// TestWasmMiddlewareInstanceReuse verifies that re-publishing a middleware with
// an unchanged guest config reuses the cached http-wasm instance (and therefore
// its wazero runtime) instead of allocating a new one.
func TestWasmMiddlewareInstanceReuse(t *testing.T) {
	ctx := log.Logger.WithContext(t.Context())
	builder := newTestWasmBuilder(t)
	cfg := reflect.ValueOf(map[string]any{"envs": []string{}})

	_, _, err := builder.buildMiddleware(ctx, testNextHandler(), cfg, "test")
	require.NoError(t, err)
	require.Len(t, builder.instances.m, 1)
	first := builder.instances.m["test"]

	// Second reload with the same name + config must reuse the same instance.
	_, _, err = builder.buildMiddleware(ctx, testNextHandler(), cfg, "test")
	require.NoError(t, err)
	require.Len(t, builder.instances.m, 1)
	second := builder.instances.m["test"]

	assert.Same(t, first, second, "instance must be reused on unchanged config")
	assert.Equal(t, first.mw, second.mw, "wazero-backed middleware must be reused")
}

// TestWasmMiddlewareConfigChangeRebuilds verifies that a changed guest config
// for the same middleware name produces a fresh instance.
func TestWasmMiddlewareConfigChangeRebuilds(t *testing.T) {
	ctx := log.Logger.WithContext(t.Context())
	builder := newTestWasmBuilder(t)

	_, _, err := builder.buildMiddleware(ctx, testNextHandler(), reflect.ValueOf(map[string]any{"envs": []string{}}), "test")
	require.NoError(t, err)
	first := builder.instances.m["test"]

	_, _, err = builder.buildMiddleware(ctx, testNextHandler(), reflect.ValueOf(map[string]any{"envs": []string{"PLUGIN_TEST"}}), "test")
	require.NoError(t, err)
	second := builder.instances.m["test"]

	require.Len(t, builder.instances.m, 1, "one cache entry per middleware name")
	assert.NotEqual(t, first.confHash, second.confHash, "config hash must change")
	assert.NotSame(t, first, second, "a changed config must rebuild the instance")
}

// TestWasmMiddlewareConfigChangeEvictsOldInstance verifies that when the guest
// config for a middleware name changes, the previous cached instance is
// markEvicted'd so its wazero runtime will be closed once its last outstanding
// handler is released. This is deterministic and does not depend on GC.
func TestWasmMiddlewareConfigChangeEvictsOldInstance(t *testing.T) {
	ctx := log.Logger.WithContext(t.Context())
	builder := newTestWasmBuilder(t)

	_, _, err := builder.buildMiddleware(ctx, testNextHandler(), reflect.ValueOf(map[string]any{"envs": []string{}}), "test")
	require.NoError(t, err)
	oldCI := builder.instances.m["test"]

	_, _, err = builder.buildMiddleware(ctx, testNextHandler(), reflect.ValueOf(map[string]any{"envs": []string{"PLUGIN_TEST"}}), "test")
	require.NoError(t, err)

	oldCI.mu.Lock()
	evicted := oldCI.evicted
	oldCI.mu.Unlock()
	assert.True(t, evicted, "old instance must be marked evicted when guest config changes")
}

// TestWasmMiddlewareNoRuntimeLeakOnReload reproduces the scenario from
// https://github.com/traefik/traefik/issues/13235: a provider that re-publishes
// the same middleware on every reload. With instance memoization, repeated
// reloads of an unchanged middleware must not accumulate runtimes — exactly one
// instance is created regardless of the number of reloads.
func TestWasmMiddlewareNoRuntimeLeakOnReload(t *testing.T) {
	ctx := log.Logger.WithContext(t.Context())
	builder := newTestWasmBuilder(t)
	cfg := reflect.ValueOf(map[string]any{"envs": []string{}})

	const reloads = 50
	for range reloads {
		_, _, err := builder.buildMiddleware(ctx, testNextHandler(), cfg, "test")
		require.NoError(t, err)
	}

	assert.Len(t, builder.instances.m, 1, "%d reloads of an unchanged middleware must create a single instance", reloads)
}
