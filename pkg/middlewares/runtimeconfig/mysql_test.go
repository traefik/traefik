package runtimeconfig

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestMySQLManagerSurvivesOneOwnerAndStopsAfterLastOwner(t *testing.T) {
	first, cancelFirst := context.WithCancel(t.Context())
	defer cancelFirst()
	second, cancelSecond := context.WithCancel(t.Context())
	defer cancelSecond()
	key := StableKey(t.Name(), t.TempDir())
	settings := MySQLSettings{DSN: "tester:example@tcp(127.0.0.1:1)/test?timeout=1ms"}
	manager, err := GetMySQLManager(first, key, settings, nil)
	require.NoError(t, err)
	same, err := GetMySQLManager(second, key, settings, nil)
	require.NoError(t, err)
	require.Same(t, manager, same)
	cancelFirst()
	require.Eventually(t, func() bool {
		mysqlRegistry.Lock()
		defer mysqlRegistry.Unlock()
		entry := mysqlRegistry.values[key]
		return entry != nil && entry.refs == 1 && entry.timer == nil
	}, time.Second, time.Millisecond)
	select {
	case <-manager.done:
		t.Fatal("another middleware still owns the manager")
	default:
	}
	cancelSecond()
	require.Eventually(t, func() bool {
		mysqlRegistry.Lock()
		defer mysqlRegistry.Unlock()
		entry := mysqlRegistry.values[key]
		if entry == nil || entry.refs != 0 || entry.timer == nil {
			return false
		}
		entry.timer.Reset(0)
		return true
	}, time.Second, time.Millisecond)
	select {
	case <-manager.done:
	case <-time.After(time.Second):
		t.Fatal("manager did not stop after its retention elapsed")
	}
}
