package runtimeconfig

import (
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
