package plugins

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMiddlewareBuilder_Wy(t *testing.T) {
	testCases := []struct {
		desc           string
		pluginType     string
		manifestUnsafe bool
		settingsUnsafe bool
		expectedHeader string
		expectedError  string
	}{
		{
			desc:           "safe plugin",
			pluginType:     "safe",
			expectedHeader: "safe",
		},
		{
			desc:           "unsafe plugin with useUnsafe",
			pluginType:     "unsafe-only",
			manifestUnsafe: true,
			settingsUnsafe: true,
			expectedHeader: "unsafe-only",
		},
		{
			desc:           "unsafe and syscall plugin with useUnsafe",
			pluginType:     "unsafe+syscall",
			manifestUnsafe: true,
			settingsUnsafe: true,
			expectedHeader: "syscall",
		},
		{
			desc:           "unsafe plugin without useUnsafe in the settings",
			pluginType:     "unsafe-only",
			manifestUnsafe: true,
			expectedError:  "restricted imports",
		},
		{
			desc:           "unsafe plugin without useUnsafe in the manifest",
			pluginType:     "unsafe-only",
			settingsUnsafe: true,
			expectedError:  "restricted imports",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			pkg := getPluginPackage(test.pluginType)
			manifest := &Manifest{Import: pkg, UseUnsafe: test.manifestUnsafe, RuntimeVersion: Wy}
			settings := Settings{UseUnsafe: test.settingsUnsafe}

			builder, err := newMiddlewareBuilder(t.Context(), "fixtures", manifest, pkg, settings, YaegiV1, false)
			if test.expectedError != "" {
				require.ErrorContains(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)
			require.IsType(t, &wyMiddlewareBuilder{}, builder)

			middleware, err := builder.newMiddleware(map[string]any{"message": "test"}, "test-middleware")
			require.NoError(t, err)

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusTeapot)
			})

			handler, err := middleware.NewHandler(t.Context(), next)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://localhost", nil))

			assert.Equal(t, http.StatusTeapot, recorder.Code)
			assert.Equal(t, test.expectedHeader, recorder.Header().Get("X-Test-Plugin"))
		})
	}
}

func TestNewMiddlewareBuilder_WyDefaultRuntimeVersion(t *testing.T) {
	manifest := &Manifest{Import: "testpluginsafe"}

	builder, err := newMiddlewareBuilder(t.Context(), "fixtures", manifest, "testpluginsafe", Settings{}, Wy, false)
	require.NoError(t, err)
	assert.IsType(t, &wyMiddlewareBuilder{}, builder)
}

// The plugin declares no useUnsafe but imports unsafe from one of its own packages.
func TestNewWyInterpreter_UndeclaredRestrictedImport(t *testing.T) {
	manifest := &Manifest{Import: "testpluginundeclared"}

	_, err := newWyInterpreter(t.Context(), "fixtures", manifest, Settings{UseUnsafe: true})
	require.ErrorContains(t, err, "restricted imports")
	assert.ErrorContains(t, err, "import of package unsafe is not allowed")
}

// A plugin wrapping the ResponseWriter keeps the optional interfaces of the writer it wraps.
func TestWyMiddleware_ResponseWriterWrapper(t *testing.T) {
	manifest := &Manifest{Import: "testpluginwrapper", RuntimeVersion: Wy}

	builder, err := newMiddlewareBuilder(t.Context(), "fixtures", manifest, manifest.Import, Settings{}, YaegiV1, false)
	require.NoError(t, err)

	middleware, err := builder.newMiddleware(nil, "test")
	require.NoError(t, err)

	var flusher, hijacker bool
	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, hijacker = rw.(http.Hijacker)
		rw.WriteHeader(http.StatusTeapot)
		flusher = http.NewResponseController(rw).Flush() == nil
	})

	handler, err := middleware.NewHandler(t.Context(), next)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusTeapot, recorder.Code)
	assert.Equal(t, "418", recorder.Header().Get("X-Status"))
	assert.True(t, flusher)
	assert.True(t, hijacker)
}

func TestNewProviderBuilder_Wy(t *testing.T) {
	manifest := &Manifest{Import: "testpluginprovider", RuntimeVersion: Wy}

	builder, err := newProviderBuilder(t.Context(), manifest, "fixtures", "testpluginprovider", Settings{}, YaegiV1, false)
	require.NoError(t, err)
	require.NotNil(t, builder.wyInterpreter)

	provider, err := newProvider(builder, map[string]any{"routerName": "foo"}, "plugin-test")
	require.NoError(t, err)

	require.NoError(t, provider.pp.Init())

	cfgChan := make(chan json.Marshaler)
	require.NoError(t, provider.pp.Provide(cfgChan))

	data, err := (<-cfgChan).MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, "{\"http\":{\"routers\":{\"foo\":{\"rule\":\"Host(`example.com`)\",\"service\":\"noop@internal\"}}}}", string(data))

	require.NoError(t, provider.pp.Stop())
}
