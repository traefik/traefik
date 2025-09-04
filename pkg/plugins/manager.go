package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

// ManagerOptions the options of a Traefik plugins manager.
type ManagerOptions struct {
	Output string
}

// Manager manages Traefik plugins lifecycle operations including storage, and manifest reading.
type Manager struct {
	archives  string
	stateFile string
	goPath    string
}

// NewManager creates a new Traefik plugins manager.
func NewManager(opts ManagerOptions) (*Manager, error) {
	sourcesRootPath := filepath.Join(filepath.FromSlash(opts.Output), sourcesFolder)
	err := resetDirectory(sourcesRootPath)
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

	return &Manager{
		archives:  archivesPath,
		stateFile: filepath.Join(archivesPath, stateFilename),
		goPath:    goPath,
	}, nil
}

// GoPath gets the plugins GoPath.
func (c *Manager) GoPath() string {
	return c.goPath
}

// ReadManifest reads a plugin manifest.
func (c *Manager) ReadManifest(moduleName string) (*Manifest, error) {
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

// CleanArchives cleans plugins archives.
func (c *Manager) CleanArchives(plugins map[string]Descriptor) error {
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
func (c *Manager) WriteState(plugins map[string]Descriptor) error {
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
func (c *Manager) ResetAll() error {
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

func (c *Manager) buildArchivePath(pName, pVersion string) string {
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
