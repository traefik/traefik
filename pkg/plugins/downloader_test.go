package plugins

import (
	"archive/zip"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPPluginDownloader_Download(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedHash   string
		expectError    bool
	}{
		{
			name: "successful download",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/zip")
				w.WriteHeader(http.StatusOK)
				// Create a simple zip content
				writer := zip.NewWriter(w)
				file, err := writer.Create("test.txt")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				_, _ = file.Write([]byte("test content"))
				_ = writer.Close()
			},
			expectError: false,
		},
		{
			name: "not modified response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				hash := r.Header.Get(hashHeader)
				if hash != "" {
					w.WriteHeader(http.StatusNotModified)
					return
				}
				w.WriteHeader(http.StatusOK)
				writer := zip.NewWriter(w)
				file, err := writer.Create("test.txt")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				_, _ = file.Write([]byte("test content"))
				_ = writer.Close()
			},
			expectError: false,
		},
		{
			name: "server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(test.serverResponse))
			defer server.Close()

			tempDir := t.TempDir()
			archivesPath := filepath.Join(tempDir, "archives")

			baseURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			downloader := &RegistryDownloader{
				httpClient: server.Client(),
				baseURL:    baseURL,
				archives:   archivesPath,
				}

			ctx := t.Context()
			hash, err := downloader.Download(ctx, "test/plugin", "v1.0.0")

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				// Check if archive file was created
				archivePath := downloader.buildArchivePath("test/plugin", "v1.0.0")
				assert.FileExists(t, archivePath)
			}
		})
	}
}

func TestHTTPPluginDownloader_Check(t *testing.T) {
	tests := []struct {
		name           string
		pHash          string
		hash           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
		expectedError  string
	}{
		{
			name:  "successful check",
			pHash: "",
			hash:  "testhash",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectError: false,
		},
		{
			name:  "failed check",
			pHash: "",
			hash:  "testhash",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			expectError: true,
		},
		{
			name:  "hash validation success",
			pHash: "correcthash",
			hash:  "correcthash",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectError: false,
		},
		{
			name:  "hash validation failure",
			pHash: "expectedhash",
			hash:  "actualhash",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectError:   true,
			expectedError: "invalid hash for plugin test/plugin, expected expectedhash, got actualhash",
		},
		{
			name:  "empty pHash skips validation",
			pHash: "",
			hash:  "anyhash",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(test.serverResponse))
			defer server.Close()

			tempDir := t.TempDir()
			archivesPath := filepath.Join(tempDir, "archives")

			baseURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			downloader := &RegistryDownloader{
				httpClient: server.Client(),
				baseURL:    baseURL,
				archives:   archivesPath,
				}

			ctx := t.Context()

			err = downloader.Check(ctx, "test/plugin", "v1.0.0", test.pHash, test.hash)

			if test.expectError {
				assert.Error(t, err)
				if test.expectedError != "" {
					assert.Equal(t, test.expectedError, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPPluginDownloader_buildArchivePath(t *testing.T) {
	downloader := &RegistryDownloader{
		archives: "/tmp/archives",
	}

	path := downloader.buildArchivePath("test/plugin", "v1.0.0")
	expected := filepath.Join("/tmp/archives", "test", "plugin", "v1.0.0.zip")
	assert.Equal(t, expected, path)
}

func TestNewHTTPPluginDownloader(t *testing.T) {
	archivesPath := "/tmp/archives"

	downloader, err := NewRegistryDownloader(RegistryDownloaderOptions{
		HTTPClient:   http.DefaultClient,
		ArchivesPath: archivesPath,
	})
	assert.NoError(t, err)
	assert.NotNil(t, downloader)
	assert.Equal(t, archivesPath, downloader.archives)
	assert.NotNil(t, downloader.httpClient)
	assert.NotNil(t, downloader.baseURL)
}
