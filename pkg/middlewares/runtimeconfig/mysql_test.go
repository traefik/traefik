package runtimeconfig

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStableKeyDoesNotContainMaterial(t *testing.T) {
	key := StableKey("access", "secret:apps/audit:dsn", "password")
	require.NotContains(t, key, "password")
	require.Len(t, key, 64)
}

func TestMySQLMaterialRevision(t *testing.T) {
	firstDSN, firstRevision, err := mysqlMaterial(MySQLSettings{DSN: "user:one@tcp(db)/audit"})
	require.NoError(t, err)
	secondDSN, secondRevision, err := mysqlMaterial(MySQLSettings{DSN: "user:two@tcp(db)/audit"})
	require.NoError(t, err)
	require.NotEqual(t, firstDSN, secondDSN)
	require.NotEqual(t, firstRevision, secondRevision)
}

func TestMySQLRejectedFileRotationKeepsActivePool(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dsn")
	require.NoError(t, os.WriteFile(path, []byte(strings.Repeat("x", maxFileSize+1)), 0o600))
	pool := &mysqlPool{db: &sql.DB{}}
	manager := &MySQLManager{current: pool, activeRevision: "active", desired: MySQLSettings{DSNFile: path}}
	manager.reconcile(t.Context(), true)
	require.Same(t, pool, manager.current)
	require.Equal(t, "active", manager.activeRevision)
	require.ErrorContains(t, manager.lastErr, "exceeds 1 MiB")
	db, release, err := manager.Acquire(t.Context())
	require.NoError(t, err)
	require.Same(t, pool.db, db)
	release()
}
