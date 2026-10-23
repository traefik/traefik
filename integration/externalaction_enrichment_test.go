package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func (s *ExternalActionSuite) TestEnrichmentPollerRecoveryAndSecretRotation() {
	var recovered atomic.Bool
	var rotated atomic.Int64
	lookup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") == "rotated-integration-token" {
			rotated.Add(1)
		}
		if r.URL.Query().Get("id") == "1" && !recovered.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"region":"region-one"}`))
	}))
	defer lookup.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "1" {
			_, _ = w.Write([]byte(`{"items":[{"id":1},{"id":2}]}`))
		} else {
			_, _ = w.Write([]byte(`{"items":[{"id":3}]}`))
		}
	}))
	defer source.Close()
	var mu sync.Mutex
	reports := map[int]int{}
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ID     int    `json:"id"`
			Region string `json:"region"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			s.T().Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		assert.Equal(s.T(), "region-one", body.Region)
		mu.Lock()
		reports[body.ID]++
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()
	tokenFile := filepath.Join(s.T().TempDir(), "token")
	require.NoError(s.T(), os.WriteFile(tokenFile, []byte("initial-integration-token"), 0o600))
	config := externalactionHTTPConfig(source.URL, dynamic.Middleware{ExternalAction: &dynamic.ExternalAction{
		Enrichment: []string{"metadata"},
		Rules: []dynamic.ExternalActionRule{{
			Name: "report", Phase: "after", Mode: "sync", FailurePolicy: "fail",
			Match:        dynamic.ExternalActionMatch{PathRegex: "^/observed$"},
			Body:         dynamic.ExternalActionBody{ReadMode: "none", Template: `{"id":{{ jsonPath .Item "$.id" | toJson }}}`},
			ResponseBody: dynamic.ExternalActionResponseBody{ReadMode: "capture", Format: "json"},
			ForEach:      &dynamic.ExternalActionForEach{Source: "responseBody", JSONPath: "$.items"},
			HTTP:         &dynamic.ExternalActionHTTP{URL: target.URL},
			Deduplication: &dynamic.ExternalActionDeduplication{
				KeyTemplate: `event:{{ jsonPath .Item "$.id" }}`,
				MySQL:       &dynamic.ExternalActionDedupMySQL{DSN: s.dsn, Table: "enrichment_events", AutoCreate: true},
			},
		}},
		Pollers: []dynamic.ExternalActionPoller{{
			Name: "observe", Interval: ptypes.Duration(time.Second),
			Request:     dynamic.ExternalActionPollerRequest{URL: source.URL},
			Observation: dynamic.ExternalActionPollerObservation{Path: "/observed"},
			Pagination: &dynamic.ExternalActionPollerPagination{
				PageParameter: "page", PageSizeParameter: "size", PageSize: 2, MaxPages: 4, ItemsJSONPath: "$.items",
			},
			Coordination: &dynamic.ExternalActionDeduplication{
				KeyTemplate: "observe", TTL: ptypes.Duration(5 * time.Second), Lease: ptypes.Duration(5 * time.Second),
				MySQL: &dynamic.ExternalActionDedupMySQL{DSN: s.dsn, Table: "enrichment_coordination", AutoCreate: true},
			},
		}},
	}})
	config.HTTP.Middlewares["metadata"] = &dynamic.Middleware{QueryEnrichment: &dynamic.QueryEnrichment{
		CredentialSets: map[string]dynamic.ExternalActionCredentialSet{"reader": {
			Auth:           dynamic.ExternalActionAuth{Type: "header", Header: "X-Token", TokenFile: tokenFile},
			ReloadInterval: ptypes.Duration(50 * time.Millisecond),
		}},
		Queries: []dynamic.EnrichmentQuery{{
			Name: "metadata", AllowedOrigins: []string{lookup.URL},
			HTTP: dynamic.ExternalActionHTTP{
				URL: lookup.URL + `?id={{ jsonPath .Payload "$.id" }}`, Auth: &dynamic.ExternalActionAuth{CredentialSet: "reader"},
			},
		}},
		OutputTemplate: `{{ $_ := set .Payload "region" (index .Enrichment "metadata" "region") }}{{ .Payload | toJson }}`,
	}}
	cmd := s.startGovernance(config)
	require.Eventually(s.T(), func() bool {
		mu.Lock()
		defer mu.Unlock()
		return reports[2] == 1 && reports[3] == 1 && reports[1] == 0
	}, 10*time.Second, 50*time.Millisecond)
	require.NoError(s.T(), os.WriteFile(tokenFile, []byte("rotated-integration-token"), 0o600))
	recovered.Store(true)
	require.Eventually(s.T(), func() bool {
		mu.Lock()
		defer mu.Unlock()
		return reports[1] == 1 && rotated.Load() > 0
	}, 10*time.Second, 50*time.Millisecond)
	var completed int
	require.NoError(s.T(), s.db.QueryRowContext(s.T().Context(), "SELECT COUNT(*) FROM enrichment_events WHERE state='completed'").Scan(&completed))
	assert.Equal(s.T(), 3, completed)
	s.killCmd(cmd)
	_ = cmd.Wait()
	s.startGovernance(config)
	assert.Never(s.T(), func() bool {
		mu.Lock()
		defer mu.Unlock()
		return reports[1] != 1 || reports[2] != 1 || reports[3] != 1
	}, 6500*time.Millisecond, 50*time.Millisecond)
}
