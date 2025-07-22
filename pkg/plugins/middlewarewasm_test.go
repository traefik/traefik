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
		getSettings func(t *testing.T) (Settings, map[string]interface{})
		expected    string
	}{
		{
			desc: "mounts path",
			getSettings: func(t *testing.T) (Settings, map[string]interface{}) {
				t.Helper()

				tempDir := t.TempDir()
				filePath := path.Join(tempDir, "hello.txt")
				err := os.WriteFile(filePath, []byte("content_test"), 0o644)
				require.NoError(t, err)

				return Settings{Mounts: []string{
						tempDir,
					}}, map[string]interface{}{
						"file": filePath,
					}
			},
			expected: "content_test",
		},
		{
			desc: "mounts src to dest",
			getSettings: func(t *testing.T) (Settings, map[string]interface{}) {
				t.Helper()

				tempDir := t.TempDir()
				filePath := path.Join(tempDir, "hello.txt")
				err := os.WriteFile(filePath, []byte("content_test"), 0o644)
				require.NoError(t, err)

				return Settings{Mounts: []string{
						tempDir + ":/tmp",
					}}, map[string]interface{}{
						"file": "/tmp/hello.txt",
					}
			},
			expected: "content_test",
		},
		{
			desc: "one env",
			getSettings: func(t *testing.T) (Settings, map[string]interface{}) {
				t.Helper()

				envs := []string{"PLUGIN_TEST"}
				return Settings{Envs: envs}, map[string]interface{}{
					"envs": envs,
				}
			},
			expected: "MY-TEST\n",
		},
		{
			desc: "two env",
			getSettings: func(t *testing.T) (Settings, map[string]interface{}) {
				t.Helper()

				envs := []string{"PLUGIN_TEST", "PLUGIN_TEST_B"}
				return Settings{Envs: envs}, map[string]interface{}{
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
