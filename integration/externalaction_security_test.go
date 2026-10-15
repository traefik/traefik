package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func (s *ExternalActionSuite) TestProjectedCredentialsAndSQLPayloads() {
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
		CredentialSets: map[string]dynamic.ExternalActionCredentialSet{"example": {
			ReloadInterval:      ptypes.Duration(time.Nanosecond),
			ImmutableParameters: []string{"X-Project"},
			Auth: dynamic.ExternalActionAuth{Type: "headers", ParameterFiles: map[string]string{
				"X-Token": filepath.Join(directory, "token"), "X-Project": filepath.Join(directory, "project"),
			}},
		}},
		Rules: []dynamic.ExternalActionRule{{
			Name: "events", Phase: "before", Mode: "sync", FailurePolicy: "fail",
			Body: dynamic.ExternalActionBody{ReadMode: "eager", Format: "json"},
			HTTP: &dynamic.ExternalActionHTTP{URL: target.URL, Auth: &dynamic.ExternalActionAuth{CredentialSet: "example"}},
			Deduplication: &dynamic.ExternalActionDeduplication{
				KeyTemplate: `{{ jsonPath .Body "$.event" }}`,
				MySQL:       &dynamic.ExternalActionDedupMySQL{DSNFile: filepath.Join(directory, "dsn"), Table: "security_dedup", AutoCreate: true},
			},
			Binding: &dynamic.ExternalActionBinding{
				Mode: "resolveOrCreate", KeyTemplate: `{{ jsonPath .Body "$.flow" }}`,
				AllowedCredentialSets: []string{"example"},
				Match:                 dynamic.ExternalActionBindingMatch{BodyJSONPaths: map[string]string{"X-Project": "$.project"}},
				MySQL:                 &dynamic.ExternalActionBindingMySQL{DSNFile: filepath.Join(directory, "dsn"), Table: "security_binding", AutoCreate: true},
			},
		}},
	}
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
	request("three'; --")
	mu.Lock()
	actual := append([]string(nil), tokens...)
	mu.Unlock()
	assert.Equal(s.T(), []string{"first-token", "first-token", "second-token"}, actual)
	var count int
	require.NoError(s.T(), s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM security_binding WHERE binding_kind='lifecycle' AND binding_key=?", payload).Scan(&count))
	assert.Equal(s.T(), 1, count)
	require.NoError(s.T(), s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM security_dedup WHERE state='completed'").Scan(&count))
	assert.Equal(s.T(), 3, count)
}
