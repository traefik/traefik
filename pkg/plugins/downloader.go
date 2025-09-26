package plugins

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

// PluginDownloader defines the interface for downloading and validating plugins from remote sources.
type PluginDownloader interface {
	// Download downloads a plugin archive and returns its hash.
	Download(ctx context.Context, pName, pVersion string) (string, error)
	// Check checks the plugin archive integrity against a known hash.
	Check(ctx context.Context, pName, pVersion, hash string) error
}

// RegistryDownloaderOptions holds configuration options for creating a RegistryDownloader.
type RegistryDownloaderOptions struct {
	HTTPClient   *http.Client
	ArchivesPath string
}

// RegistryDownloader implements PluginDownloader for HTTP-based plugin downloads.
type RegistryDownloader struct {
	httpClient *http.Client
	baseURL    *url.URL
	archives   string
}

// NewRegistryDownloader creates a new HTTP-based plugin downloader.
func NewRegistryDownloader(opts RegistryDownloaderOptions) (*RegistryDownloader, error) {
	baseURL, err := url.Parse(pluginsURL)
	if err != nil {
		return nil, err
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &RegistryDownloader{
		httpClient: httpClient,
		baseURL:    baseURL,
		archives:   opts.ArchivesPath,
	}, nil
}

// Download downloads a plugin archive.
func (d *RegistryDownloader) Download(ctx context.Context, pName, pVersion string) (string, error) {
	filename := d.buildArchivePath(pName, pVersion)

	var hash string
	_, err := os.Stat(filename)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to read archive %s: %w", filename, err)
	}

	if err == nil {
		hash, err = computeHash(filename)
		if err != nil {
			return "", fmt.Errorf("failed to compute hash: %w", err)
		}
	}

	endpoint, err := d.baseURL.Parse(path.Join(d.baseURL.Path, "download", pName, pVersion))
	if err != nil {
		return "", fmt.Errorf("failed to parse endpoint URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if hash != "" {
		req.Header.Set(hashHeader, hash)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call service: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotModified:
		return hash, nil
	case http.StatusOK:
		err = os.MkdirAll(filepath.Dir(filename), 0o755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		var file *os.File
		file, err = os.Create(filename)
		if err != nil {
			return "", fmt.Errorf("failed to create file %q: %w", filename, err)
		}

		defer func() { _ = file.Close() }()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to write response: %w", err)
		}

		hash, err = computeHash(filename)
		if err != nil {
			return "", fmt.Errorf("failed to compute hash: %w", err)
		}
	default:
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error: %d: %s", resp.StatusCode, string(data))
	}

	return hash, nil
}

// Check checks the plugin archive integrity.
func (d *RegistryDownloader) Check(ctx context.Context, pName, pVersion, hash string) error {
	endpoint, err := d.baseURL.Parse(path.Join(d.baseURL.Path, "validate", pName, pVersion))
	if err != nil {
		return fmt.Errorf("failed to parse endpoint URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if hash != "" {
		req.Header.Set(hashHeader, hash)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call service: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.New("plugin integrity check failed")
}

// buildArchivePath builds the path to a plugin archive file.
func (d *RegistryDownloader) buildArchivePath(pName, pVersion string) string {
	return filepath.Join(d.archives, filepath.FromSlash(pName), pVersion+".zip")
}
