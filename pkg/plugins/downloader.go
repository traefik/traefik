package plugins

import (
	zipa "archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
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
	SourcesPath  string
}

// RegistryDownloader implements PluginDownloader for HTTP-based plugin downloads.
type RegistryDownloader struct {
	httpClient *http.Client
	baseURL    *url.URL
	archives   string
	sources    string
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
		sources:    opts.SourcesPath,
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

	// Unzip the downloaded archive
	err = d.unzip(pName, pVersion)
	if err != nil {
		return "", fmt.Errorf("failed to unzip archive: %w", err)
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

// unzip unzips a plugin archive to the sources directory.
func (d *RegistryDownloader) unzip(pName, pVersion string) error {
	err := d.unzipModule(pName, pVersion)
	if err == nil {
		return nil
	}

	// Unzip as a generic archive if the module unzip fails.
	// This is useful for plugins that have vendor directories or other structures.
	// This is also useful for wasm plugins.
	return d.unzipArchive(pName, pVersion)
}

func (d *RegistryDownloader) unzipModule(pName, pVersion string) error {
	src := d.buildArchivePath(pName, pVersion)
	dest := filepath.Join(d.sources, filepath.FromSlash(pName))

	return zip.Unzip(dest, module.Version{Path: pName, Version: pVersion}, src)
}

func (d *RegistryDownloader) unzipArchive(pName, pVersion string) error {
	zipPath := d.buildArchivePath(pName, pVersion)

	archive, err := zipa.OpenReader(zipPath)
	if err != nil {
		return err
	}

	defer func() { _ = archive.Close() }()

	dest := filepath.Join(d.sources, filepath.FromSlash(pName))

	for _, f := range archive.File {
		err = d.unzipFile(f, dest)
		if err != nil {
			return fmt.Errorf("unable to unzip %s: %w", f.Name, err)
		}
	}

	return nil
}

func (d *RegistryDownloader) unzipFile(f *zipa.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}

	defer func() { _ = rc.Close() }()

	// Split to discard the first part of the path when the archive is a Yaegi go plugin with vendoring.
	// In this case the path starts with `[organization]-[project]-[release commit sha1]/`.
	pathParts := strings.SplitN(f.Name, "/", 2)
	var fileName string
	if len(pathParts) < 2 {
		fileName = pathParts[0]
	} else {
		fileName = pathParts[1]
	}

	// Validate and sanitize the file path.
	cleanName := filepath.Clean(fileName)
	if strings.Contains(cleanName, "..") {
		return fmt.Errorf("invalid file path in archive: %s", f.Name)
	}

	filePath := filepath.Join(dest, cleanName)
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolving file path: %w", err)
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("resolving destination directory: %w", err)
	}

	if !strings.HasPrefix(absFilePath, absDest) {
		return fmt.Errorf("file path escapes destination directory: %s", absFilePath)
	}

	if f.FileInfo().IsDir() {
		err = os.MkdirAll(filePath, f.Mode())
		if err != nil {
			return fmt.Errorf("unable to create archive directory %s: %w", filePath, err)
		}

		return nil
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0o750)
	if err != nil {
		return fmt.Errorf("unable to create archive directory %s for file %s: %w", filepath.Dir(filePath), filePath, err)
	}

	elt, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}

	defer func() { _ = elt.Close() }()

	_, err = io.Copy(elt, rc)
	if err != nil {
		return err
	}

	return nil
}
