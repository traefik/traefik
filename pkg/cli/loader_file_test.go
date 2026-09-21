package cli

import (
	"fmt"
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

	testCases := []struct {
		desc       string
		configFile string
	}{
		{desc: "missing file", configFile: "fixtures/missing-traefik.toml"},
		{desc: "spaces", configFile: "   "},
		{desc: "tabs and newlines", configFile: "\t\n"},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var config map[string]any

			configFile, err := loadConfigFiles(test.configFile, &config)

			require.EqualError(t, err, fmt.Sprintf("configuration file %q not found", test.configFile))
			require.Empty(t, configFile)
		})
	}
}

func TestLoadConfigFilesUsesDefaultConfigSearchWhenNoConfigFileProvided(t *testing.T) {
	// System configuration takes precedence over the temporary search paths below.
	for _, extension := range []string{"toml", "yaml", "yml"} {
		path := "/etc/traefik/traefik." + extension
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		}
		require.NoError(t, err)
		t.Skipf("system configuration %s prevents isolating the default search", path)
	}

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
	require.Equal(t, map[string]any{"log": map[string]any{"level": "DEBUG"}}, config)
}
