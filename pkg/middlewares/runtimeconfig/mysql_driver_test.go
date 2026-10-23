package runtimeconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMySQLDriverConfigSeparatesStatementsAndValues(t *testing.T) {
	config, err := mysqlDriverConfig("audit:example@tcp(db.example:3306)/audit?parseTime=true&timeout=3s&readTimeout=4s&writeTimeout=5s&tls=true&charset=utf8mb4&multiStatements=true&interpolateParams=true&allowAllFiles=true")
	require.NoError(t, err)
	require.Equal(t, "audit", config.User)
	require.Equal(t, "example", config.Passwd)
	require.Equal(t, "tcp", config.Net)
	require.Equal(t, "db.example:3306", config.Addr)
	require.Equal(t, "audit", config.DBName)
	require.True(t, config.ParseTime)
	require.Equal(t, 3*time.Second, config.Timeout)
	require.Equal(t, 4*time.Second, config.ReadTimeout)
	require.Equal(t, 5*time.Second, config.WriteTimeout)
	require.Equal(t, "true", config.TLSConfig)
	require.Contains(t, config.FormatDSN(), "charset=utf8mb4")
	require.False(t, config.MultiStatements)
	require.False(t, config.InterpolateParams)
	require.False(t, config.AllowAllFiles)
}

func TestMySQLDriverConfigErrorsDoNotExposeMaterial(t *testing.T) {
	for _, dsn := range []string{
		"private-material",
		"audit:private-material@tcp(db.example)/audit?timeout=private-material",
		"audit:private-material@tcp(db.example)/audit?tls=private-material",
	} {
		_, err := mysqlDriverConfig(dsn)
		require.EqualError(t, err, "invalid MySQL DSN")
		require.NotContains(t, err.Error(), "private-material")
	}
}

func TestMySQLDriverConfigPreservesDefaultOptions(t *testing.T) {
	config, err := mysqlDriverConfig("audit:example@tcp(db.example)/audit?parseTime=true")
	require.NoError(t, err)
	require.False(t, config.MultiStatements)
	require.False(t, config.InterpolateParams)
	require.False(t, config.AllowAllFiles)
	require.True(t, config.ParseTime)
}
