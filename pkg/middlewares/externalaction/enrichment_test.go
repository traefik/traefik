package externalaction

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func enrichmentFixture(endpoint string) dynamic.QueryEnrichment {
	return dynamic.QueryEnrichment{
		InputTemplate: `{"id":{{ index .Payload "id" | toJson }}}`,
		Queries: []dynamic.EnrichmentQuery{{
			Name: "metadata", AllowedOrigins: []string{endpoint},
			HTTP:           dynamic.ExternalActionHTTP{URL: endpoint + `?id={{ index .Input "id" | urlquery }}`, Method: "GET"},
			ResultTemplate: `{{ .Result | toJson }}`,
		}},
		OutputTemplate: `{{ $_ := set .Payload "metadata" (index .Enrichment "metadata") }}{{ .Payload | toJson }}`,
	}
}

func TestQueryEnrichmentChainAndPoller(t *testing.T) {
	var calls atomic.Int64
	lookup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		assert.Equal(t, "metadata-token", r.Header.Get("Authorization"))
		assert.Equal(t, "destination-project", r.Header.Get("X-Source-Project"))
		require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"id": r.URL.Query().Get("id"), "region": "region-one"}))
	}))
	t.Cleanup(lookup.Close)
	var mu sync.Mutex
	var reports []map[string]any
	report := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "destination-token", r.Header.Get("Authorization"))
		value, err := decodeJSON(func() []byte { b, _ := io.ReadAll(r.Body); return b }())
		require.NoError(t, err)
		mu.Lock()
		reports = append(reports, value.(map[string]any))
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(report.Close)
	definition := enrichmentFixture(lookup.URL)
	definition.Queries[0].HTTP.Auth = &dynamic.ExternalActionAuth{Type: "header", Header: "Authorization", Token: "metadata-token"}
	definition.Queries[0].HTTP.Headers = map[string]string{"X-Source-Project": `{{ index .SourceAuthParameters "project" }}`}
	rule := dynamic.ExternalActionRule{
		Name: "event", Phase: "after", Mode: "sync",
		Match:        dynamic.ExternalActionMatch{PathRegex: "^/events$"},
		HTTP:         &dynamic.ExternalActionHTTP{URL: report.URL, Auth: &dynamic.ExternalActionAuth{Type: "header", Header: "Authorization", Token: "destination-token", Parameters: map[string]string{"project": "destination-project"}}},
		Body:         dynamic.ExternalActionBody{ReadMode: "none", ContentType: "application/json", Template: `{"id":{{ index .Item "id" | toJson }},"existing":"preserved"}`},
		ResponseBody: dynamic.ExternalActionResponseBody{ReadMode: "capture", Format: "json"},
		ForEach:      &dynamic.ExternalActionForEach{Source: "responseBody", JSONPath: "$.items"},
	}
	config := dynamic.ExternalAction{Enrichment: []string{"lookup"}, Rules: []dynamic.ExternalActionRule{rule}}
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"items":[{"id":9007199254740993},{"id":2}]}`)
	})
	handler, err := NewWithEnrichment(t.Context(), backend, config, t.Name(), map[string]dynamic.QueryEnrichment{"lookup": definition})
	require.NoError(t, err)
	chain, err := NewQueryEnrichment(t.Context(), handler, definition, "lookup")
	require.NoError(t, err)
	rw := httptest.NewRecorder()
	chain.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/events", nil))
	assert.Equal(t, 200, rw.Code)
	assert.JSONEq(t, `{"items":[{"id":9007199254740993},{"id":2}]}`, rw.Body.String())
	assert.Equal(t, int64(2), calls.Load())
	mu.Lock()
	require.Len(t, reports, 2)
	assert.Equal(t, "preserved", reports[0]["existing"])
	mu.Unlock()
	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/unmatched", nil))
	assert.Equal(t, int64(2), calls.Load())
	// A poller uses explicit references without ever entering the chain.
	source := httptest.NewServer(backend)
	t.Cleanup(source.Close)
	poller, err := newCompiledPoller(t.Context(), dynamic.ExternalActionPoller{
		Name: "observer", Interval: durationValue(t, time.Minute),
		Request:      dynamic.ExternalActionPollerRequest{URL: source.URL},
		Observation:  dynamic.ExternalActionPollerObservation{Method: "GET", Host: "example.test", Path: "/events"},
		Coordination: memoryCoordination(t, "enrichment-poller"),
	}, nil)
	require.NoError(t, err)
	data, err := poller.observationData(t.Context())
	require.NoError(t, err)
	_, err = poller.pollPage(t.Context(), handler.(*middleware), data, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(4), calls.Load())
}

func TestQueryEnrichmentValidationAndFailure(t *testing.T) {
	for _, scenario := range []string{"status", "invalid-json", "too-large", "redirect", "business-error", "timeout", "origin", "mapping"} {
		t.Run(scenario, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch scenario {
				case "status":
					w.WriteHeader(http.StatusServiceUnavailable)
				case "invalid-json":
					_, _ = io.WriteString(w, "not-json")
				case "too-large":
					_, _ = io.WriteString(w, strings.Repeat("x", 101))
				case "redirect":
					w.Header().Set("Location", "http://forbidden.example")
					w.WriteHeader(http.StatusFound)
				case "timeout":
					<-r.Context().Done()
				case "business-error":
					_, _ = io.WriteString(w, `{"code":12}`)
				default:
					_, _ = io.WriteString(w, `{"code":0}`)
				}
			}))
			t.Cleanup(server.Close)
			config := enrichmentFixture(server.URL)
			q := &config.Queries[0]
			q.Timeout = durationValue(t, 50*time.Millisecond)
			q.HTTP.MaxResponseBytes = 100
			q.CacheTTL = durationValue(t, time.Minute)
			q.HTTP.ResponseValidation = &dynamic.ExternalActionHTTPResponse{SuccessJSONPath: "$.code", SuccessValues: []string{"0"}}
			if scenario == "origin" {
				q.AllowedOrigins = []string{"https://allowed.example"}
			}
			if scenario == "mapping" {
				q.ResultTemplate = `{{ required "field missing" (jsonPath .Result "$.missing") | toJson }}`
			}
			compiled, err := compileEnrichment(t.Name(), config)
			require.NoError(t, err)
			_, _, err = enrichPayload(t.Context(), &actionData{Context: t.Context()}, `{"id":1}`, []*compiledEnrichment{compiled})
			require.Error(t, err)
			if scenario != "mapping" {
				assert.Empty(t, compiled.queries[0].cache)
			}
		})
	}
	for _, change := range []func(*dynamic.EnrichmentQuery){
		func(q *dynamic.EnrichmentQuery) { q.AllowedOrigins = nil },
		func(q *dynamic.EnrichmentQuery) { q.HTTP.Method = "DELETE" },
		func(q *dynamic.EnrichmentQuery) { q.MaxConcurrent = 65 },
		func(q *dynamic.EnrichmentQuery) { q.CacheEntries = 5000 },
		func(q *dynamic.EnrichmentQuery) { q.AllowedOrigins = []string{"https://user:secret@example.test"} },
	} {
		config := enrichmentFixture("http://example.test")
		change(&config.Queries[0])
		_, err := compileEnrichment(t.Name(), config)
		require.Error(t, err)
	}
}

func TestQueryEnrichmentSecretReloadAndBoundedCache(t *testing.T) {
	var calls atomic.Int64
	var token atomic.Value
	token.Store("one")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		assert.Equal(t, token.Load(), r.Header.Get("X-Token"))
		_, _ = io.WriteString(w, `{"id":9007199254740993}`)
	}))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(path, []byte("one"), 0o600))
	config := enrichmentFixture(server.URL)
	config.CredentialSets = map[string]dynamic.ExternalActionCredentialSet{"lookup": {
		Auth: dynamic.ExternalActionAuth{Type: "header", Header: "X-Token", TokenFile: path}, ReloadInterval: durationValue(t, time.Millisecond),
	}}
	query := &config.Queries[0]
	query.HTTP.Auth = &dynamic.ExternalActionAuth{CredentialSet: "lookup"}
	query.CacheTTL = durationValue(t, time.Minute)
	query.CacheEntries = 2
	query.CacheBytes = 80
	compiled, err := compileEnrichment(t.Name(), config)
	require.NoError(t, err)
	run := func(id int) string {
		_, body, err := enrichPayload(t.Context(), &actionData{Context: t.Context()}, fmt.Sprintf(`{"id":%d}`, id), []*compiledEnrichment{compiled})
		require.NoError(t, err)
		return body
	}
	assert.Contains(t, run(1), "9007199254740993")
	run(1)
	assert.Equal(t, int64(1), calls.Load())
	token.Store("two")
	require.NoError(t, os.WriteFile(path, []byte("two"), 0o600))
	time.Sleep(5 * time.Millisecond)
	run(1)
	assert.Equal(t, int64(2), calls.Load())
	for i := 2; i < 10; i++ {
		run(i)
	}
	assert.LessOrEqual(t, len(compiled.queries[0].cache), 2)
	assert.LessOrEqual(t, compiled.queries[0].bytes, int64(80))
}

func TestQueryEnrichmentConcurrencyAndIsolation(t *testing.T) {
	var active, peak atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := active.Add(1)
		for previous := peak.Load(); count > previous; previous = peak.Load() {
			if peak.CompareAndSwap(previous, count) {
				break
			}
		}
		defer active.Add(-1)
		time.Sleep(5 * time.Millisecond)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(server.Close)
	config := enrichmentFixture(server.URL)
	config.Queries[0].MaxConcurrent = 2
	config.OutputTemplate = `{{ $_ := set .Item "changed" true }}{{ .Payload | toJson }}`
	compiled, err := compileEnrichment(t.Name(), config)
	require.NoError(t, err)
	data := &actionData{Context: t.Context(), Item: map[string]any{"id": 1}}
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			_, _, err := enrichPayload(t.Context(), data, `{"id":1}`, []*compiledEnrichment{compiled})
			assert.NoError(t, err)
		})
	}
	wg.Wait()
	assert.Equal(t, int64(2), peak.Load())
	assert.NotContains(t, data.Item, "changed")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, err = enrichPayload(ctx, data, `{"id":1}`, []*compiledEnrichment{compiled})
	require.Error(t, err)
}

func TestQueryEnrichmentRuntimeSecretRebuild(t *testing.T) {
	var tokens []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens = append(tokens, r.Header.Get("X-Token"))
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(server.Close)
	for _, token := range []string{"first-secret", "second-secret"} {
		config := enrichmentFixture(server.URL)
		config.RuntimeValues = map[string]dynamic.RuntimeValue{"credential": {
			Source: "secret:example/lookup:arbitrary-key", Revision: token, Value: token,
		}}
		config.CredentialSets = map[string]dynamic.ExternalActionCredentialSet{"query": {
			Auth: dynamic.ExternalActionAuth{Type: "header", Header: "X-Token", Token: "runtime://credential"},
		}}
		config.Queries[0].HTTP.Auth = &dynamic.ExternalActionAuth{CredentialSet: "query"}
		compiled, err := getEnrichment(t.Context(), t.Name(), config)
		require.NoError(t, err)
		_, _, err = enrichPayload(t.Context(), &actionData{Context: t.Context()}, `{"id":1}`, []*compiledEnrichment{compiled})
		require.NoError(t, err)
		assert.Equal(t, "runtime://credential", config.CredentialSets["query"].Auth.Token)
		encoded, err := json.Marshal(config)
		require.NoError(t, err)
		assert.NotContains(t, string(encoded), `"Value"`)
	}
	assert.Equal(t, []string{"first-secret", "second-secret"}, tokens)
}

func TestQueryEnrichmentFailureReleasesDedupAndDoesNotChangeResponse(t *testing.T) {
	var queries, reports int
	lookup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries++
		if queries == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, `{"suffix":"enriched"}`)
	}))
	t.Cleanup(lookup.Close)
	report := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reports++
		assert.Equal(t, "/enriched", r.URL.Path)
		assert.Equal(t, "enriched", r.Header.Get("X-Mapped"))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(report.Close)
	config := dynamic.ExternalAction{Enrichment: []string{"lookup"}, Rules: []dynamic.ExternalActionRule{{
		Name: "retry", Phase: "after", Mode: "sync", FailurePolicy: "ignore",
		Body: dynamic.ExternalActionBody{ReadMode: "none", Template: `{"id":1}`},
		HTTP: &dynamic.ExternalActionHTTP{
			URL:     report.URL + `/{{ index .Enrichment "metadata" "suffix" }}`,
			Headers: map[string]string{"X-Mapped": `{{ index .Enrichment "metadata" "suffix" }}`},
		},
		Deduplication: &dynamic.ExternalActionDeduplication{KeyTemplate: "one-event", Memory: &dynamic.ExternalActionDedupMemory{MaxEntries: 10}},
	}}}
	handler, err := NewWithEnrichment(t.Context(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, "business-body")
	}), config, t.Name(), map[string]dynamic.QueryEnrichment{"lookup": enrichmentFixture(lookup.URL)})
	require.NoError(t, err)
	for range 3 {
		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/observe", nil))
		assert.Equal(t, 202, rw.Code)
		assert.Equal(t, "business-body", rw.Body.String())
	}
	assert.Equal(t, 2, queries)
	assert.Equal(t, 1, reports)
}

func TestQueryEnrichmentExec(t *testing.T) {
	lookup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"resource":"resource;false"}`)
	}))
	t.Cleanup(lookup.Close)
	definition := enrichmentFixture(lookup.URL)
	definition.OutputTemplate = `{{ index .Enrichment "metadata" | toJson }}`
	compiled, err := compileEnrichment(t.Name(), definition)
	require.NoError(t, err)
	executable, err := newExecTarget(dynamic.ExternalActionRule{Exec: &dynamic.ExternalActionExec{
		Command: os.Args[0],
		Args:    []string{"-test.run=^TestExternalActionExecHelperProcess$", "--", `{{ index .Payload "resource" }}`, "alice", "verify"},
		Env:     map[string]string{"EXTERNAL_ACTION_HELPER": "1", "EXPECTED_RESOURCE": `{{ index .Payload "resource" }}`},
		Stdin:   `{"id":1}`,
	}})
	require.NoError(t, err)
	executable.(*execTarget).enrichments = []*compiledEnrichment{compiled}
	require.NoError(t, executable.Run(t.Context(), &actionData{Context: t.Context()}))
}

func TestQueryEnrichmentPaginationRejectsIncompleteResults(t *testing.T) {
	for _, response := range []string{
		`{"items":[],"total":3}`,
		`{"items":[{"id":1}],"total":3}`,
		`{"items":[{"id":1}],"total":0}`,
		`{"items":{},"total":1}`,
		`{"items":[],"total":"invalid"}`,
	} {
		t.Run(response, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, response)
			}))
			t.Cleanup(server.Close)
			config := enrichmentFixture(server.URL)
			config.Queries[0].Pagination = &dynamic.ExternalActionPollerPagination{
				PageParameter: "page", PageSizeParameter: "size", PageSize: 1, MaxPages: 1,
				ItemsJSONPath: "$.items", TotalJSONPath: "$.total",
			}
			compiled, err := compileEnrichment(t.Name(), config)
			require.NoError(t, err)
			_, _, err = enrichPayload(t.Context(), &actionData{Context: t.Context()}, `{"id":1}`, []*compiledEnrichment{compiled})
			require.Error(t, err)
		})
	}
}
