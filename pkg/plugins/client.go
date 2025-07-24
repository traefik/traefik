package plugins

import (
	zipa "archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
	"gopkg.in/yaml.v3"
)

const (
	sourcesFolder  = "sources"
	archivesFolder = "archives"
	stateFilename  = "state.json"
	goPathSrc      = "src"
	pluginManifest = ".traefik.yml"
)

const pluginsURL = "https://plugins.traefik.io/public/"

const (
	hashHeader = "X-Plugin-Hash"
)

// ClientOptions the options of a Traefik plugins client.
type ClientOptions struct {
	Output string
}

// Client a Traefik plugins client.
type Client struct {
	HTTPClient *http.Client
	baseURL    *url.URL

	archives  string
	stateFile string
	goPath    string
	sources   string
}

// NewClient creates a new Traefik plugins client.
func NewClient(opts ClientOptions) (*Client, error) {
	baseURL, err := url.Parse(pluginsURL)
	if err != nil {
		return nil, err
	}

	sourcesRootPath := filepath.Join(filepath.FromSlash(opts.Output), sourcesFolder)
	err = resetDirectory(sourcesRootPath)
	if err != nil {
		return nil, err
	}

	goPath, err := os.MkdirTemp(sourcesRootPath, "gop-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create GoPath: %w", err)
	}

	archivesPath := filepath.Join(filepath.FromSlash(opts.Output), archivesFolder)
	err = os.MkdirAll(archivesPath, 0o755)
	if err != nil {
		return nil, fmt.Errorf("failed to create archives directory %s: %w", archivesPath, err)
	}

	client := retryablehttp.NewClient()
	client.Logger = logs.NewRetryableHTTPLogger(log.Logger)
	client.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	client.RetryMax = 3

	return &Client{
		HTTPClient: client.StandardClient(),
		baseURL:    baseURL,

		archives:  archivesPath,
		stateFile: filepath.Join(archivesPath, stateFilename),

		goPath:  goPath,
		sources: filepath.Join(goPath, goPathSrc),
	}, nil
}

// GoPath gets the plugins GoPath.
func (c *Client) GoPath() string {
	return c.goPath
}

// ReadManifest reads a plugin manifest.
func (c *Client) ReadManifest(moduleName string) (*Manifest, error) {
	return ReadManifest(c.goPath, moduleName)
}

// ReadManifest reads a plugin manifest.
func ReadManifest(goPath, moduleName string) (*Manifest, error) {
	p := filepath.Join(goPath, goPathSrc, filepath.FromSlash(moduleName), pluginManifest)

	file, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open the plugin manifest %s: %w", p, err)
	}

	defer func() { _ = file.Close() }()

	m := &Manifest{}
	err = yaml.NewDecoder(file).Decode(m)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the plugin manifest %s: %w", p, err)
	}

	return m, nil
}

// Download downloads a plugin archive.
func (c *Client) Download(ctx context.Context, pName, pVersion string) (string, error) {
	filename := c.buildArchivePath(pName, pVersion)

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

	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "download", pName, pVersion))
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

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call service: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotModified:
		// noop
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

		return hash, nil

	default:
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error: %d: %s", resp.StatusCode, string(data))
	}
}

// Check checks the plugin archive integrity.
func (c *Client) Check(ctx context.Context, pName, pVersion, hash string) error {
	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, "validate", pName, pVersion))
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

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call service: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.New("plugin integrity check failed")
}

// Unzip unzip a plugin archive.
func (c *Client) Unzip(pName, pVersion string) error {
	err := c.unzipModule(pName, pVersion)
	if err == nil {
		return nil
	}

	// Unzip as a generic archive if the module unzip fails.
	// This is useful for plugins that have vendor directories or other structures.
	// This is also useful for wasm plugins.
	return c.unzipArchive(pName, pVersion)
}

func (c *Client) unzipModule(pName, pVersion string) error {
	src := c.buildArchivePath(pName, pVersion)
	dest := filepath.Join(c.sources, filepath.FromSlash(pName))

	return zip.Unzip(dest, module.Version{Path: pName, Version: pVersion}, src)
}

func (c *Client) unzipArchive(pName, pVersion string) error {
	zipPath := c.buildArchivePath(pName, pVersion)

	archive, err := zipa.OpenReader(zipPath)
	if err != nil {
		return err
	}

	defer func() { _ = archive.Close() }()

	dest := filepath.Join(c.sources, filepath.FromSlash(pName))

	for _, f := range archive.File {
		err = unzipFile(f, dest)
		if err != nil {
			return fmt.Errorf("unable to unzip %s: %w", f.Name, err)
		}
	}

	return nil
}

func unzipFile(f *zipa.File, dest string) error {
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

// CleanArchives cleans plugins archives.
func (c *Client) CleanArchives(plugins map[string]Descriptor) error {
	if _, err := os.Stat(c.stateFile); os.IsNotExist(err) {
		return nil
	}

	stateFile, err := os.Open(c.stateFile)
	if err != nil {
		return fmt.Errorf("failed to open state file %s: %w", c.stateFile, err)
	}

	previous := make(map[string]string)
	err = json.NewDecoder(stateFile).Decode(&previous)
	if err != nil {
		return fmt.Errorf("failed to decode state file %s: %w", c.stateFile, err)
	}

	for pName, pVersion := range previous {
		for _, desc := range plugins {
			if desc.ModuleName == pName && desc.Version != pVersion {
				archivePath := c.buildArchivePath(pName, pVersion)
				if err = os.RemoveAll(archivePath); err != nil {
					return fmt.Errorf("failed to remove archive %s: %w", archivePath, err)
				}
			}
		}
	}

	return nil
}

// WriteState writes the plugins state files.
func (c *Client) WriteState(plugins map[string]Descriptor) error {
	m := make(map[string]string)

	for _, descriptor := range plugins {
		m[descriptor.ModuleName] = descriptor.Version
	}

	mp, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal plugin state: %w", err)
	}

	return os.WriteFile(c.stateFile, mp, 0o600)
}

// ResetAll resets all plugins related directories.
func (c *Client) ResetAll() error {
	if c.goPath == "" {
		return errors.New("goPath is empty")
	}

	err := resetDirectory(filepath.Join(c.goPath, ".."))
	if err != nil {
		return fmt.Errorf("unable to reset plugins GoPath directory %s: %w", c.goPath, err)
	}

	err = resetDirectory(c.archives)
	if err != nil {
		return fmt.Errorf("unable to reset plugins archives directory: %w", err)
	}

	return nil
}

func (c *Client) buildArchivePath(pName, pVersion string) string {
	return filepath.Join(c.archives, filepath.FromSlash(pName), pVersion+".zip")
}

func resetDirectory(dir string) error {
	dirPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("unable to get absolute path of %s: %w", dir, err)
	}

	currentPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get the current directory: %w", err)
	}

	if strings.HasPrefix(currentPath, dirPath) {
		return fmt.Errorf("cannot be deleted: the directory path %s is the parent of the current path %s", dirPath, currentPath)
	}

	err = os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("unable to remove directory %s: %w", dirPath, err)
	}

	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("unable to create directory %s: %w", dirPath, err)
	}

	return nil
}

func computeHash(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	sum := hash.Sum(nil)

	return hex.EncodeToString(sum), nil
}
