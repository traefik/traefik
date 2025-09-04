package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewPluginManager(t *testing.T) {
	tempDir := t.TempDir()
	opts := ManagerOptions{Output: tempDir}

	manager, err := NewManager(opts)
	require.NoError(t, err)
	assert.NotNil(t, manager)

	sourcesPath := filepath.Join(tempDir, sourcesFolder)
	assert.DirExists(t, sourcesPath)

	archivesPath := filepath.Join(tempDir, archivesFolder)
	assert.DirExists(t, archivesPath)

	assert.NotEmpty(t, manager.GoPath())
	assert.DirExists(t, manager.GoPath())
}

func TestPluginManager_GoPath(t *testing.T) {
	tempDir := t.TempDir()
	opts := ManagerOptions{Output: tempDir}

	manager, err := NewManager(opts)
	require.NoError(t, err)

	goPath := manager.GoPath()
	assert.NotEmpty(t, goPath)
	assert.Contains(t, goPath, tempDir)
}

func TestPluginManager_ReadManifest(t *testing.T) {
	tempDir := t.TempDir()
	opts := ManagerOptions{Output: tempDir}

	manager, err := NewManager(opts)
	require.NoError(t, err)

	moduleName := "github.com/test/plugin"
	pluginPath := filepath.Join(manager.goPath, "src", moduleName)
	err = os.MkdirAll(pluginPath, 0o755)
	require.NoError(t, err)

	manifest := &Manifest{
		DisplayName: "Test Plugin",
		Type:        "middleware",
		Import:      "github.com/test/plugin",
		Summary:     "A test plugin",
		TestData: map[string]interface{}{
			"test": "data",
		},
	}

	manifestPath := filepath.Join(pluginPath, pluginManifest)
	manifestData, err := yaml.Marshal(manifest)
	require.NoError(t, err)
	err = os.WriteFile(manifestPath, manifestData, 0o644)
	require.NoError(t, err)

	readManifest, err := manager.ReadManifest(moduleName)
	require.NoError(t, err)
	assert.Equal(t, manifest.DisplayName, readManifest.DisplayName)
	assert.Equal(t, manifest.Type, readManifest.Type)
	assert.Equal(t, manifest.Import, readManifest.Import)
	assert.Equal(t, manifest.Summary, readManifest.Summary)
}

func TestPluginManager_ReadManifest_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	opts := ManagerOptions{Output: tempDir}

	manager, err := NewManager(opts)
	require.NoError(t, err)

	_, err = manager.ReadManifest("nonexistent/plugin")
	assert.Error(t, err)
}

func TestReadManifest_Function(t *testing.T) {
	tempDir := t.TempDir()
	goPath := filepath.Join(tempDir, "gopath")
	sourcesPath := filepath.Join(goPath, goPathSrc)

	moduleName := "github.com/test/plugin"
	pluginPath := filepath.Join(sourcesPath, moduleName)
	err := os.MkdirAll(pluginPath, 0o755)
	require.NoError(t, err)

	manifest := &Manifest{
		DisplayName: "Test Plugin",
		Type:        "middleware",
		Import:      "github.com/test/plugin",
		Summary:     "A test plugin",
		TestData: map[string]interface{}{
			"test": "data",
		},
	}

	manifestPath := filepath.Join(pluginPath, pluginManifest)
	manifestData, err := yaml.Marshal(manifest)
	require.NoError(t, err)
	err = os.WriteFile(manifestPath, manifestData, 0o644)
	require.NoError(t, err)

	readManifest, err := ReadManifest(goPath, moduleName)
	require.NoError(t, err)
	assert.Equal(t, manifest.DisplayName, readManifest.DisplayName)
	assert.Equal(t, manifest.Type, readManifest.Type)
}

func TestPluginManager_CleanArchives(t *testing.T) {
	tempDir := t.TempDir()
	opts := ManagerOptions{Output: tempDir}

	manager, err := NewManager(opts)
	require.NoError(t, err)

	testPlugin1 := "test/plugin1"
	testPlugin2 := "test/plugin2"

	archive1Dir := filepath.Join(manager.archives, "test", "plugin1")
	archive2Dir := filepath.Join(manager.archives, "test", "plugin2")
	err = os.MkdirAll(archive1Dir, 0o755)
	require.NoError(t, err)
	err = os.MkdirAll(archive2Dir, 0o755)
	require.NoError(t, err)

	archive1Old := filepath.Join(archive1Dir, "v1.0.0.zip")
	archive1New := filepath.Join(archive1Dir, "v2.0.0.zip")
	archive2 := filepath.Join(archive2Dir, "v1.0.0.zip")

	err = os.WriteFile(archive1Old, []byte("old archive"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(archive1New, []byte("new archive"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(archive2, []byte("archive 2"), 0o644)
	require.NoError(t, err)

	state := map[string]string{
		testPlugin1: "v1.0.0",
		testPlugin2: "v1.0.0",
	}
	stateData, err := json.MarshalIndent(state, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(manager.stateFile, stateData, 0o600)
	require.NoError(t, err)

	currentPlugins := map[string]Descriptor{
		"plugin1": {
			ModuleName: testPlugin1,
			Version:    "v2.0.0",
		},
		"plugin2": {
			ModuleName: testPlugin2,
			Version:    "v1.0.0",
		},
	}

	err = manager.CleanArchives(currentPlugins)
	require.NoError(t, err)

	assert.NoFileExists(t, archive1Old)
	assert.FileExists(t, archive1New)
	assert.FileExists(t, archive2)
}

func TestPluginManager_WriteState(t *testing.T) {
	tempDir := t.TempDir()
	opts := ManagerOptions{Output: tempDir}

	manager, err := NewManager(opts)
	require.NoError(t, err)

	plugins := map[string]Descriptor{
		"plugin1": {
			ModuleName: "test/plugin1",
			Version:    "v1.0.0",
		},
		"plugin2": {
			ModuleName: "test/plugin2",
			Version:    "v2.0.0",
		},
	}

	err = manager.WriteState(plugins)
	require.NoError(t, err)

	assert.FileExists(t, manager.stateFile)

	data, err := os.ReadFile(manager.stateFile)
	require.NoError(t, err)

	var state map[string]string
	err = json.Unmarshal(data, &state)
	require.NoError(t, err)

	expectedState := map[string]string{
		"test/plugin1": "v1.0.0",
		"test/plugin2": "v2.0.0",
	}
	assert.Equal(t, expectedState, state)
}

func TestPluginManager_ResetAll(t *testing.T) {
	tempDir := t.TempDir()
	opts := ManagerOptions{Output: tempDir}

	manager, err := NewManager(opts)
	require.NoError(t, err)

	testFile := filepath.Join(manager.GoPath(), "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0o644)
	require.NoError(t, err)

	archiveFile := filepath.Join(manager.archives, "test.zip")
	err = os.WriteFile(archiveFile, []byte("archive"), 0o644)
	require.NoError(t, err)

	err = manager.ResetAll()
	require.NoError(t, err)

	assert.DirExists(t, manager.archives)
	assert.NoFileExists(t, testFile)
	assert.NoFileExists(t, archiveFile)
}

func TestPluginManager_ResetAll_EmptyGoPath(t *testing.T) {
	manager := &Manager{goPath: ""}

	err := manager.ResetAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "goPath is empty")
}

func Test_computeHash(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content for hashing")

	err := os.WriteFile(testFile, content, 0o644)
	require.NoError(t, err)

	hash, err := computeHash(testFile)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)

	hash2, err := computeHash(testFile)
	require.NoError(t, err)
	assert.Equal(t, hash, hash2)
}

func Test_computeHash_NonexistentFile(t *testing.T) {
	_, err := computeHash("/nonexistent/file")
	assert.Error(t, err)
}

func Test_resetDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "testdir")

	err := os.MkdirAll(testDir, 0o755)
	require.NoError(t, err)

	testFile := filepath.Join(testDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0o644)
	require.NoError(t, err)

	err = resetDirectory(testDir)
	require.NoError(t, err)

	// Directory should exist but be empty
	assert.DirExists(t, testDir)
	assert.NoFileExists(t, testFile)
}

func Test_resetDirectory_CurrentPath(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	err = resetDirectory(currentDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be deleted")
}
