//go:build linux

package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginManager_CleanArchivesClosesStateFile(t *testing.T) {
	t.Parallel()

	stateFile := filepath.Join(t.TempDir(), "plugins-state.json")
	require.NoError(t, os.WriteFile(stateFile, []byte(`{"github.com/traefik/plugintest":"v0.1.0"}`), 0o600))

	manager := &Manager{stateFile: stateFile}
	for range 10 {
		require.NoError(t, manager.CleanArchives(nil))
	}

	require.Zero(t, countOpenDescriptorsTo(t, stateFile))
}

func countOpenDescriptorsTo(t *testing.T, path string) int {
	t.Helper()

	entries, err := os.ReadDir("/proc/self/fd")
	require.NoError(t, err)

	var count int
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err != nil {
			continue
		}

		if target == path {
			count++
		}
	}

	return count
}
