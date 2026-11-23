package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigFilesRejectsOversizedFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "traefik.yml")
	require.NoError(t, os.WriteFile(filename, make([]byte, maxConfigFileSize+1), 0o600))

	_, err := loadConfigFiles(filename, &struct{}{})
	require.ErrorContains(t, err, "exceeds")
}
