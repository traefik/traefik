package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/cmd"
)

func TestLoadConfigFiles(t *testing.T) {
	dir := t.TempDir()
	existingFile := filepath.Join(dir, "traefik.toml")
	require.NoError(t, os.WriteFile(existingFile, []byte("[log]\n  level = \"DEBUG\"\n"), 0o600))

	tests := []struct {
		desc       string
		configFile string
		wantErr    bool
	}{
		{
			desc:       "existing file",
			configFile: existingFile,
		},
		{
			desc:       "non-existent explicit file",
			configFile: filepath.Join(dir, "does-not-exist.toml"),
			wantErr:    true,
		},
		{
			desc:       "no file specified",
			configFile: "",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tconfig := cmd.NewTraefikConfiguration()
			got, err := loadConfigFiles(test.configFile, tconfig)
			if test.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
				return
			}
			require.NoError(t, err)
			if test.configFile != "" {
				assert.Equal(t, test.configFile, got)
			}
		})
	}
}
