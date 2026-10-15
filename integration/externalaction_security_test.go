package integration

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

func (s *ExternalActionSuite) TestProjectedCredentialsAndSQLPayloads() {
	const credentialName = "example'); DROP TABLE security_binding; --"
	directory := s.T().TempDir()
	for _, name := range []string{"token", "project", "dsn"} {
		require.NoError(s.T(), os.Symlink(filepath.Join("..data", name), filepath.Join(directory, name)))
	}
	rotate := func(version, token string) {
		s.T().Helper()
		require.NoError(s.T(), os.Mkdir(filepath.Join(directory, version), 0o700))
		for name, value := range map[string]string{"token": token, "project": "example", "dsn": s.dsn} {
			require.NoError(s.T(), os.WriteFile(filepath.Join(directory, version, name), []byte(value), 0o600))
		}
		require.NoError(s.T(), os.Symlink(version, filepath.Join(directory, "..next")))
		require.NoError(s.T(), os.Rename(filepath.Join(directory, "..next"), filepath.Join(directory, "..data")))
	}
	rotate("one", "first-token")
	var mu sync.Mutex
	var tokens []string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		tokens = append(tokens, r.Header.Get("X-Token"))
		mu.Unlock()
		assert.Equal(s.T(), "example", r.Header.Get("X-Project"))
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer backend.Close()
	config := &dynamic.ExternalAction{
		CredentialSets: map[string]dynamic.ExternalActionCredentialSet{credentialName: {
			ReloadInterval:      ptypes.Duration(time.Nanosecond),
			ImmutableParameters: []string{"X-Project"},
			Auth: dynamic.ExternalActionAuth{Type: "headers", ParameterFiles: map[string]string{
				"X-Token": filepath.Join(directory, "token"), "X-Project": filepath.Join(directory, "project"),
			}},
		}},
		Rules: []dynamic.ExternalActionRule{{
			Name: "events'; --", Phase: "before", Mode: "sync", FailurePolicy: "fail",
			Match: dynamic.ExternalActionMatch{PathRegex: "^/events$"},
			Body:  dynamic.ExternalActionBody{ReadMode: "eager", Format: "json"},
			HTTP:  &dynamic.ExternalActionHTTP{URL: target.URL, Auth: &dynamic.ExternalActionAuth{CredentialSet: credentialName}},
			Deduplication: &dynamic.ExternalActionDeduplication{
				KeyTemplate: `{{ jsonPath .Body "$.event" }}`,
				MySQL:       &dynamic.ExternalActionDedupMySQL{DSNFile: filepath.Join(directory, "dsn"), Table: "security_dedup", AutoCreate: true},
			},
			Binding: &dynamic.ExternalActionBinding{
				Mode: "create", KeyTemplate: `{{ jsonPath .Body "$.flow" }}`,
				AllowedCredentialSets: []string{credentialName},
				Match:                 dynamic.ExternalActionBindingMatch{BodyJSONPaths: map[string]string{"X-Project": "$.project"}},
				MySQL:                 &dynamic.ExternalActionBindingMySQL{DSNFile: filepath.Join(directory, "dsn"), Table: "security_binding", AutoCreate: true},
			},
		}},
	}
	resolve := config.Rules[0]
	resolve.Name = "resolve"
	resolve.Match.PathRegex = "^/resolve$"
	binding := *resolve.Binding
	binding.Mode = "resolve"
	binding.Match = dynamic.ExternalActionBindingMatch{}
	resolve.Binding = &binding
	config.Rules = append(config.Rules, resolve)
	s.startGovernance(externalactionHTTPConfig(backend.URL, dynamic.Middleware{ExternalAction: config}))
	payload := "flow'); DROP TABLE security_binding; --"
	request := func(event string) {
		s.T().Helper()
		body, err := json.Marshal(map[string]string{"event": event, "flow": payload, "project": "example"})
		require.NoError(s.T(), err)
		s.request("/events", string(body), nil)
	}
	request("one'; --")
	rotate("invalid", strings.Repeat("x", (1<<20)+1))
	request("two'; --")
	rotate("two", "second-token")
	request("three'; --")
	body, err := json.Marshal(map[string]string{"event": "resolve'; --", "flow": payload})
	require.NoError(s.T(), err)
	s.request("/resolve", string(body), nil)
	request("three'; --")
	// Reclaim an expired row rather than inserting a new one.
	_, err = s.db.ExecContext(s.T().Context(), "UPDATE security_dedup SET expires_at=DATE_SUB(UTC_TIMESTAMP(6), INTERVAL 1 SECOND)")
	require.NoError(s.T(), err)
	request("three'; --")
	mu.Lock()
	actual := append([]string(nil), tokens...)
	mu.Unlock()
	assert.Equal(s.T(), []string{"first-token", "first-token", "second-token", "second-token", "second-token"}, actual)
	var count int
	require.NoError(s.T(), s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM security_binding WHERE binding_kind='lifecycle' AND binding_key=?", payload).Scan(&count))
	assert.Equal(s.T(), 1, count)
	require.NoError(s.T(), s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM security_binding WHERE binding_kind='credential' AND credential_set=?", credentialName).Scan(&count))
	assert.Equal(s.T(), 1, count)
	require.NoError(s.T(), s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM security_dedup WHERE state='completed'").Scan(&count))
	assert.Equal(s.T(), 4, count)
}

func (s *ExternalActionSuite) TestMySQLSecretRotationPreservesLeasesAndDriverSafety() {
	directory := s.T().TempDir()
	path := filepath.Join(directory, "dsn")
	require.NoError(s.T(), os.Symlink(filepath.Join("..data", "dsn"), path))
	rotate := func(revision, dsn string) {
		s.T().Helper()
		require.NoError(s.T(), os.Mkdir(filepath.Join(directory, revision), 0o700))
		require.NoError(s.T(), os.WriteFile(filepath.Join(directory, revision, "dsn"), []byte(dsn), 0o600))
		require.NoError(s.T(), os.Symlink(revision, filepath.Join(directory, "..next")))
		require.NoError(s.T(), os.Rename(filepath.Join(directory, "..next"), filepath.Join(directory, "..data")))
	}
	config, err := mysql.ParseDSN(s.dsn)
	require.NoError(s.T(), err)
	config.MultiStatements, config.InterpolateParams, config.AllowAllFiles = true, true, true
	rotate("first", config.FormatDSN())
	manager, err := runtimeconfig.GetMySQLManager(s.T().Context(), directory, runtimeconfig.MySQLSettings{
		DSNFile: path, RefreshInterval: time.Millisecond,
	}, nil)
	require.NoError(s.T(), err)
	oldDB, release, err := manager.Acquire(s.T().Context())
	require.NoError(s.T(), err)
	defer release()
	_, err = oldDB.ExecContext(s.T().Context(), "SELECT 1; SELECT 2")
	require.Error(s.T(), err)
	stmt, err := oldDB.PrepareContext(s.T().Context(), "SELECT ?")
	require.NoError(s.T(), err)
	defer stmt.Close()
	const payload = "'; DROP DATABASE governance; --"
	var actual string
	require.NoError(s.T(), stmt.QueryRowContext(s.T().Context(), payload).Scan(&actual))
	assert.Equal(s.T(), payload, actual)

	rotate("invalid", "malformed-dsn")
	time.Sleep(6 * time.Second)
	active, done, err := manager.Acquire(s.T().Context())
	require.NoError(s.T(), err)
	assert.Same(s.T(), oldDB, active)
	done()
	config.ReadTimeout = 3 * time.Second
	rotate("second", config.FormatDSN())
	var newDB *sql.DB
	require.Eventually(s.T(), func() bool {
		db, done, err := manager.Acquire(s.T().Context())
		if err != nil {
			return false
		}
		defer done()
		if db == oldDB {
			return false
		}
		newDB = db
		_, err = db.ExecContext(s.T().Context(), "SELECT 1; SELECT 2")
		return err != nil
	}, 15*time.Second, 100*time.Millisecond)
	require.NotSame(s.T(), oldDB, newDB)
	// A prepared operation keeps its old lease usable during the pool swap.
	require.NoError(s.T(), stmt.QueryRowContext(s.T().Context(), payload).Scan(&actual))
	assert.Equal(s.T(), payload, actual)
	require.NoError(s.T(), newDB.QueryRowContext(s.T().Context(), "SELECT ?", payload).Scan(&actual))
	assert.Equal(s.T(), payload, actual)
}

func (s *ExternalActionSuite) TestMySQLFailedActionReleasesClaimForRetry() {
	var calls atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer backend.Close()
	s.startGovernance(externalactionHTTPConfig(backend.URL, dynamic.Middleware{ExternalAction: &dynamic.ExternalAction{
		Rules: []dynamic.ExternalActionRule{{
			Name: "retry'; --", Phase: "before", Mode: "sync", FailurePolicy: "fail",
			Body: dynamic.ExternalActionBody{ReadMode: "eager", Format: "json"},
			HTTP: &dynamic.ExternalActionHTTP{URL: target.URL},
			Deduplication: &dynamic.ExternalActionDeduplication{
				KeyTemplate: `{{ jsonPath .Body "$.event" }}`,
				MySQL:       &dynamic.ExternalActionDedupMySQL{DSN: s.dsn, Table: "retry_dedup", AutoCreate: true},
			},
		}},
	}}))
	const body = `{"event":"retry'); DROP TABLE retry_dedup; --"}`
	request, err := http.NewRequestWithContext(s.T().Context(), http.MethodPost, "http://127.0.0.1:8000/events", strings.NewReader(body))
	require.NoError(s.T(), err)
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())
	require.Equal(s.T(), http.StatusBadGateway, response.StatusCode)
	var count int
	require.NoError(s.T(), s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM retry_dedup").Scan(&count))
	assert.Zero(s.T(), count)
	s.request("/events", body, nil)
	s.request("/events", body, nil)
	assert.EqualValues(s.T(), 2, calls.Load())
	require.NoError(s.T(), s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM retry_dedup WHERE state='completed' AND rule_name=?", "retry'; --").Scan(&count))
	assert.Equal(s.T(), 1, count)
}
