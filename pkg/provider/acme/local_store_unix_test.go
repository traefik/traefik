//go:build !windows

package acme

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/safe"
)

func TestLocalStore_SaveAccount_failedWriteKeepsPreviousFile(t *testing.T) {
	// RLIMIT_FSIZE is process-wide and would also break go test's own log writes, so run in a child process.
	if os.Getenv("ACME_FSIZE_CHILD") == "" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestLocalStore_SaveAccount_failedWriteKeepsPreviousFile$", "-test.v")
		cmd.Env = append(os.Environ(), "ACME_FSIZE_CHILD=1")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return
	}

	acmeFile := filepath.Join(t.TempDir(), "acme.json")

	previous := `{"test":{"Account":{"Email":"previous@email.com"}}}`
	err := os.WriteFile(acmeFile, []byte(previous), 0o600)
	require.NoError(t, err)

	s := NewLocalStore(acmeFile, safe.NewPool(t.Context()))

	// Load the existing file before limiting writes.
	_, err = s.GetAccount("test")
	require.NoError(t, err)

	// Make every write past 16 bytes fail, like a full disk or a crash mid-write.
	var limit syscall.Rlimit
	require.NoError(t, syscall.Getrlimit(syscall.RLIMIT_FSIZE, &limit))
	restricted := limit
	restricted.Cur = 16
	require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_FSIZE, &restricted))
	t.Cleanup(func() { _ = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit) })

	err = s.SaveAccount("test", &Account{Email: strings.Repeat("a", 1024) + "@email.com"})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit))

	file, err := os.ReadFile(acmeFile)
	require.NoError(t, err)
	assert.Equal(t, previous, string(file))

	entries, err := os.ReadDir(filepath.Dir(acmeFile))
	require.NoError(t, err)
	assert.Len(t, entries, 1, "temporary file left behind")
}

func TestLocalStore_SaveAccount_keepsExistingFileMode(t *testing.T) {
	acmeFile := filepath.Join(t.TempDir(), "acme.json")

	err := os.WriteFile(acmeFile, []byte(`{}`), 0o600)
	require.NoError(t, err)
	// The store only accepts owner-only modes; use one other than the 0600 of a new file.
	require.NoError(t, os.Chmod(acmeFile, 0o400))

	s := NewLocalStore(acmeFile, safe.NewPool(t.Context()))

	err = s.SaveAccount("test", &Account{Email: "some@email.com"})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	file, err := os.ReadFile(acmeFile)
	require.NoError(t, err)
	assert.Contains(t, string(file), `"Email": "some@email.com"`)

	fi, err := os.Stat(acmeFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o400), fi.Mode().Perm())
}
