package plugins

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/yaegi/v2/vm"
)

// A Go update that changes a package copied by Yaegi v2 needs a Yaegi v2 update.
func TestYaegiV2_AOTCopies(t *testing.T) {
	require.NoError(t, vm.CheckAOT(false))
}

func TestResolveYaegiVersion(t *testing.T) {
	testCases := []struct {
		desc           string
		defaultVersion string
		manifest       string
		expected       string
		wantErr        bool
	}{
		{desc: "default v1", defaultVersion: "v1", expected: YaegiV1},
		{desc: "default v2", defaultVersion: "v2", expected: YaegiV2},
		{desc: "manifest v2 with default v1", defaultVersion: "v1", manifest: "v2", expected: YaegiV2},
		{desc: "manifest v1 is refused", defaultVersion: "v2", manifest: "v1", wantErr: true},
		{desc: "unknown default", defaultVersion: "v3", wantErr: true},
		{desc: "empty default", wantErr: true},
		{desc: "unknown manifest", defaultVersion: "v2", manifest: "foo", wantErr: true},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			version, err := resolveYaegiVersion(test.defaultVersion, &Manifest{RuntimeVersion: test.manifest})
			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, version)
		})
	}
}

func TestNewMiddlewareBuilder_YaegiV2(t *testing.T) {
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
			desc:          "unsafe plugin without useUnsafe in the manifest",
			pluginType:    "unsafe-only",
			expectedError: "unsafe",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			pkg := getPluginPackage(test.pluginType)
			manifest := &Manifest{Import: pkg, UseUnsafe: test.manifestUnsafe, RuntimeVersion: YaegiV2}
			settings := Settings{UseUnsafe: test.settingsUnsafe}

			builder, err := newMiddlewareBuilder(t.Context(), "fixtures", manifest, pkg, settings, YaegiV1, false)
			if test.expectedError != "" {
				require.ErrorContains(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)
			require.IsType(t, yaegiV2MiddlewareBuilder{}, builder)

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

func TestNewMiddlewareBuilder_YaegiVersion(t *testing.T) {
	manifest := &Manifest{Import: "testpluginsafe"}

	builder, err := newMiddlewareBuilder(t.Context(), "fixtures", manifest, "testpluginsafe", Settings{}, YaegiV2, false)
	require.NoError(t, err)
	assert.IsType(t, yaegiV2MiddlewareBuilder{}, builder)

	builder, err = newMiddlewareBuilder(t.Context(), "fixtures", manifest, "testpluginsafe", Settings{}, YaegiV1, false)
	require.NoError(t, err)
	assert.IsType(t, &yaegiMiddlewareBuilder{}, builder)

	_, err = newMiddlewareBuilder(t.Context(), "fixtures", manifest, "testpluginsafe", Settings{}, "v3", false)
	require.Error(t, err)
}

func TestNewProviderBuilder_YaegiV2(t *testing.T) {
	manifest := &Manifest{Import: "testpluginprovider", RuntimeVersion: YaegiV2}

	builder, err := newProviderBuilder(t.Context(), manifest, "fixtures", "testpluginprovider", Settings{}, YaegiV1, false)
	require.NoError(t, err)
	require.NotNil(t, builder.plugin)

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
