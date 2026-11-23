package acme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/safe"
)

func TestLocalStoreRejectsOversizedFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "acme.json")
	require.NoError(t, os.WriteFile(filename, make([]byte, maxStorageFileSize+1), 0o600))

	store := NewLocalStore(filename, safe.NewPool(t.Context()))
	_, err := store.GetAccount("test")
	require.ErrorContains(t, err, "exceeds")
}
