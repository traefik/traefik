package plugins

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const jsonPluginManifest = `displayName: JSON plugin
type: middleware
import: github.com/traefik/jsonplugin
`

// The plugin uses a type that implements json.Marshaler.
// Since Go 1.27, json.Marshaler is an alias of the jsonv2 interface,
// and yaegi before v0.16.3 panics when it wraps this type.
const jsonPluginSource = `package jsonplugin

import (
	"context"
	"encoding/json"
	"net/http"
)

type Config struct{}

func CreateConfig() *Config {
	return &Config{}
}

type number int

func (n number) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]int{"num": int(n)})
}

type jsonPlugin struct{}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &jsonPlugin{}, nil
}

func (p *jsonPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	data, err := json.Marshal(number(3))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = rw.Write(data)
}
`

func TestBuilder_localPluginWithJSONMarshaler(t *testing.T) {
	dir := t.TempDir()

	pluginDir := filepath.Join(dir, localGoPath, goPathSrc, "github.com", "traefik", "jsonplugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, pluginManifest), []byte(jsonPluginManifest), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "jsonplugin.go"), []byte(jsonPluginSource), 0o600))

	t.Chdir(dir)

	builder, err := NewBuilder(nil, nil, map[string]LocalDescriptor{
		"jsonplugin": {ModuleName: "github.com/traefik/jsonplugin"},
	})
	require.NoError(t, err)

	constructor, err := builder.Build("jsonplugin", nil, "test")
	require.NoError(t, err)

	handler, err := constructor(context.Background(), http.NotFoundHandler())
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.JSONEq(t, `{"num":3}`, rw.Body.String())
}
