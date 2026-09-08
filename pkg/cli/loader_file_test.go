package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigFilesReturnsErrorForMissingExplicitConfigFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Chdir(t.TempDir())

	// Ensure an explicit missing file is not silently replaced by a default file.
	require.NoError(t, os.WriteFile("traefik.toml", []byte("[log]\nlevel = \"DEBUG\"\n"), 0o600))

	var config map[string]any

	configFile, err := loadConfigFiles("fixtures/missing-traefik.toml", &config)

	require.Error(t, err)
	require.Empty(t, configFile)
	require.Contains(t, err.Error(), "fixtures/missing-traefik.toml")
}

func TestLoadConfigFilesUsesDefaultConfigSearchWhenNoConfigFileProvided(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	defaultConfig := filepath.Join(tmpDir, "traefik.toml")
	require.NoError(t, os.WriteFile(defaultConfig, []byte("[log]\nlevel = \"DEBUG\"\n"), 0o600))

	var config map[string]any

	configFile, err := loadConfigFiles("", &config)

	require.NoError(t, err)
	require.Equal(t, defaultConfig, configFile)
}
