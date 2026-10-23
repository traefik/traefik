package accessgovernance

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestSecurityRejectsTableNamesAfterRuntimeResolution(t *testing.T) {
	unsafe := []string{"events` SET x=1 --", "events;DROP TABLE users", "db.events", "events/*x*/", "events\x00", "events\n", "events'", "events\"", "events\\", "../events", "{{ .Body.table }}", "${runtime.missing}", "évents", strings.Repeat("a", 65)}
	for _, table := range unsafe {
		for _, runtimeValue := range []bool{false, true} {
			for _, network := range []bool{false, true} {
				for _, autoCreate := range []bool{false, true} {
					for _, protocol := range []string{"http", "tcp", "udp"} {
						config := dynamic.AccessGovernance{Databases: map[string]dynamic.AccessGovernanceDatabase{}}
						config.Schema.AutoCreate = autoCreate
						name := table
						if runtimeValue {
							name = "runtime://table"
							config.RuntimeValues = map[string]dynamic.RuntimeValue{"table": {Value: table}}
						}
						database := dynamic.AccessGovernanceDatabase{DSN: "unused:unused@tcp(127.0.0.1:1)/unused"}
						if network {
							database.NetworkTable = name
						} else {
							database.HTTPTable = name
						}
						config.Databases["audit"] = database
						var err error
						switch protocol {
						case "http":
							_, err = New(t.Context(), nil, config, "review")
						case "tcp":
							_, err = NewTCP(t.Context(), nil, config, "review")
						case "udp":
							_, err = NewUDP(t.Context(), nil, config, "review")
						}
						require.Error(t, err)
						require.True(t, strings.Contains(err.Error(), "table names must match") || strings.Contains(err.Error(), "not declared"), err.Error())
					}
				}
			}
		}
	}
}

func TestSecurityEventValuesRemainSQLArguments(t *testing.T) {
	recordingDriverRegisterOnce.Do(func() { sql.Register("access-governance-recording", testRecordingDriver) })
	db, err := sql.Open("access-governance-recording", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	payload := "x'); DROP TABLE events; --"
	require.NoError(t, insertHTTP(t.Context(), db, "http_events", []httpEvent{{PrincipalID: payload, Path: payload, Query: payload, UserAgent: payload, Labels: map[string]string{"value": payload}, Metadata: map[string]any{"value": payload}}}))
	query, args := testRecordingDriver.snapshot()
	require.NotContains(t, query, payload)
	require.Equal(t, len(args), strings.Count(query, "?"))
	require.Equal(t, payload, args[11].Value)
	require.NoError(t, insertNetwork(t.Context(), db, "network_events", []networkEvent{{PrincipalID: payload, ClientAddr: payload, Labels: map[string]string{"value": payload}}}))
	query, args = testRecordingDriver.snapshot()
	require.NotContains(t, query, payload)
	require.Equal(t, len(args), strings.Count(query, "?"))
	require.Equal(t, payload, args[11].Value)
}

func TestSQLBoundariesRejectUnsafeTable(t *testing.T) {
	table := "events`;DROP TABLE users;--"
	require.Error(t, insertHTTP(t.Context(), nil, table, []httpEvent{{}}))
	require.Error(t, insertNetwork(t.Context(), nil, table, []networkEvent{{}}))
	require.Error(t, ensureTable(t.Context(), nil, table, nil, nil))
	_, err := readIndexes(t.Context(), nil, table)
	require.Error(t, err)
}
