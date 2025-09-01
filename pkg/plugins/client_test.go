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

func TestNewClient(t *testing.T) {
	tests := []struct {
		desc      string
		options   ClientOptions
		wantError bool
	}{
		{
			desc: "valid options",
			options: ClientOptions{
				HTTPClient: &http.Client{},
				BaseURL:    "https://plugins.example.com/",
				Output:     t.TempDir(),
			},
		},
		{
			desc: "invalid base URL",
			options: ClientOptions{
				HTTPClient: &http.Client{},
				BaseURL:    "://invalid-url",
				Output:     t.TempDir(),
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			client, err := NewClient(test.options)

			if test.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.options.HTTPClient, client.httpClient)
			assert.Equal(t, test.options.BaseURL, client.baseURL.String())
		})
	}
}

func TestClient_GoPath(t *testing.T) {
	tempDir := t.TempDir()
	client, err := NewClient(ClientOptions{
		HTTPClient: &http.Client{},
		BaseURL:    "https://plugins.example.com/",
		Output:     tempDir,
	})
	require.NoError(t, err)

	goPath := client.GoPath()
	assert.NotEmpty(t, goPath)
	assert.Contains(t, goPath, tempDir)
}

func TestClient_Download(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/download/test/plugin/v1.0.0":
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("fake-zip-content"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		HTTPClient: server.Client(),
		BaseURL:    server.URL + "/",
		Output:     t.TempDir(),
	})
	require.NoError(t, err)

	t.Run("successful download", func(t *testing.T) {
		hash, err := client.Download(context.Background(), "test/plugin", "v1.0.0")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		expectedPath := client.buildArchivePath("test/plugin", "v1.0.0")
		assert.FileExists(t, expectedPath)
	})

	t.Run("plugin not found", func(t *testing.T) {
		_, err = client.Download(context.Background(), "test/notfound", "v1.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error: 404")
	})
}

func TestClient_Check(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/validate/test/valid/v1.0.0":
			w.WriteHeader(http.StatusOK)
		case "/validate/test/invalid/v1.0.0":
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		HTTPClient: server.Client(),
		BaseURL:    server.URL + "/",
		Output:     t.TempDir(),
	})
	require.NoError(t, err)

	t.Run("valid plugin", func(t *testing.T) {
		err = client.Check(context.Background(), "test/valid", "v1.0.0", "test-hash")
		assert.NoError(t, err)
	})

	t.Run("invalid plugin", func(t *testing.T) {
		err = client.Check(context.Background(), "test/invalid", "v1.0.0", "test-hash")
		assert.Error(t, err)
	})
}

func TestClient_ReadManifest(t *testing.T) {
	client, err := NewClient(ClientOptions{
		HTTPClient: &http.Client{},
		BaseURL:    "https://plugins.example.com/",
		Output:     t.TempDir(),
	})
	require.NoError(t, err)

	moduleName := "github.com/test/plugin"
	manifestPath := filepath.Join(client.GoPath(), "src", filepath.FromSlash(moduleName), ".traefik.yml")
	err = os.MkdirAll(filepath.Dir(manifestPath), 0755)
	require.NoError(t, err)

	manifestContent := `displayName: Test Plugin
type: middleware
runtime: yaegi
import: github.com/test/plugin
summary: A test plugin
`
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	manifest, err := client.ReadManifest(moduleName)
	require.NoError(t, err)
	assert.Equal(t, "Test Plugin", manifest.DisplayName)
	assert.Equal(t, "middleware", manifest.Type)
	assert.Equal(t, "yaegi", manifest.Runtime)
	assert.Equal(t, "github.com/test/plugin", manifest.Import)
	assert.Equal(t, "A test plugin", manifest.Summary)
}
