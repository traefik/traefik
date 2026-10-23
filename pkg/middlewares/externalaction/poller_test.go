package externalaction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestPollerPaginatesAndCoordinatesDirectObservations(t *testing.T) {
	var sourceCalls atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		require.True(t, ok)
		assert.Equal(t, "observer", username)
		assert.Equal(t, "secret", password)
		assert.Equal(t, "team-a", req.URL.Query().Get("tenant"))
		page, err := strconv.Atoi(req.URL.Query().Get("page"))
		require.NoError(t, err)
		assert.Equal(t, "2", req.URL.Query().Get("page_size"))
		sourceCalls.Add(1)
		rw.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			_, _ = rw.Write([]byte(`{"code":0,"data":{"total":3,"items":[{"id":1,"status":"success"},{"id":2,"status":"failed"}]}}`))
		case 2:
			_, _ = rw.Write([]byte(`{"code":0,"data":{"total":3,"items":[{"id":3,"status":"success"}]}}`))
		default:
			t.Fatalf("unexpected page %d", page)
		}
	}))
	t.Cleanup(source.Close)

	var targetMu sync.Mutex
	var targetIDs []int
	target := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var payload struct {
			ID int `json:"id"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
		targetMu.Lock()
		targetIDs = append(targetIDs, payload.ID)
		targetMu.Unlock()
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(target.Close)

	config := dynamic.ExternalAction{
		SelectionMode: "all",
		Pollers: []dynamic.ExternalActionPoller{{
			Name:     "sop-history",
			Interval: durationValue(t, time.Minute),
			Timeout:  durationValue(t, time.Second),
			Request: dynamic.ExternalActionPollerRequest{
				URL:              source.URL + `?tenant={{ authParameter . "tenant" }}`,
				Method:           http.MethodGet,
				MaxResponseBytes: 1024 * 1024,
				Auth: &dynamic.ExternalActionAuth{
					Type:     "basic",
					Username: "observer",
					Password: "secret",
					Parameters: map[string]string{
						"tenant": "team-a",
					},
				},
			},
			Observation: dynamic.ExternalActionPollerObservation{
				Method: http.MethodGet,
				Host:   "automation.example.com",
				Path:   "/api/v1/jobs",
				Headers: map[string]string{
					"X-Governance-Principal": "automatic-reconciler",
				},
			},
			Pagination: &dynamic.ExternalActionPollerPagination{
				PageParameter:     "page",
				PageSizeParameter: "page_size",
				PageSize:          2,
				MaxPages:          4,
				ItemsJSONPath:     "$.data.items",
				TotalJSONPath:     "$.data.total",
			},
			Coordination: memoryCoordination(t, "poller:sop-history"),
		}},
		Rules: []dynamic.ExternalActionRule{{
			Name: "sop-terminal",
			Match: dynamic.ExternalActionMatch{
				Methods:          []string{http.MethodGet},
				Hosts:            []string{"automation.example.com"},
				PathRegex:        `^/api/v1/jobs$`,
				ResponseStatuses: []string{"200-299"},
			},
			Phase: "after",
			Mode:  "sync",
			Body: dynamic.ExternalActionBody{
				ReadMode:    "none",
				Format:      "json",
				ContentType: "application/json",
				Template:    `{"id":{{ jsonPath .Item "$.id" }}}`,
			},
			ResponseBody: dynamic.ExternalActionResponseBody{ReadMode: "capture", MaxBytes: 1024 * 1024, Format: "json"},
			ForEach: &dynamic.ExternalActionForEach{
				Source:         "responseBody",
				JSONPath:       "$.data.items",
				MatchJSONPaths: map[string]string{"$.status": `^(success|failed)$`},
			},
			HTTP: &dynamic.ExternalActionHTTP{URL: target.URL, Method: http.MethodPost},
		}},
	}

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	compiled, err := compileConfigWithContext(ctx, config)
	require.NoError(t, err)
	handler := &middleware{config: compiled, logger: zerolog.Nop()}
	require.NoError(t, compiled.pollers[0].runCoordinated(ctx, handler))
	require.NoError(t, compiled.pollers[0].runCoordinated(ctx, handler))

	assert.Equal(t, int64(2), sourceCalls.Load())
	targetMu.Lock()
	assert.ElementsMatch(t, []int{1, 2, 3}, targetIDs)
	targetMu.Unlock()
}

func TestPollerBuildsNeutralResponseEvent(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"success":true,"data":[{"ID":42,"Status":"succeed"}]}`))
	}))
	t.Cleanup(source.Close)

	var targetCalls atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		targetCalls.Add(1)
		var payload struct {
			ID int `json:"id"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
		assert.Equal(t, 42, payload.ID)
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(target.Close)

	config := dynamic.ExternalAction{
		Pollers: []dynamic.ExternalActionPoller{{
			Name:     "migration-terminal",
			Interval: durationValue(t, time.Minute),
			Request: dynamic.ExternalActionPollerRequest{
				URL:              source.URL,
				MaxResponseBytes: 1024 * 1024,
			},
			Observation: dynamic.ExternalActionPollerObservation{
				Method: http.MethodGet,
				Host:   "automation.example.com",
				Path:   "/api/v1/migrations",
				Event: &dynamic.ExternalActionPollerEvent{
					Source:    "job-service",
					Principal: "automatic-reconciler",
				},
			},
			Coordination: memoryCoordination(t, "poller:migration-terminal"),
		}},
		Rules: []dynamic.ExternalActionRule{{
			Name: "migration-event",
			Match: dynamic.ExternalActionMatch{
				Methods:   []string{http.MethodPost},
				PathRegex: `^/_governance/v1/events$`,
				BodyJSONPaths: map[string]string{
					"$.specversion":                `^1\.0$`,
					"$.source":                     `^job-service$`,
					"$.data.request.path":          `^/api/v1/migrations$`,
					"$.data.response_body.success": `^true$`,
				},
			},
			Phase:         "before",
			Mode:          "sync",
			FailurePolicy: "fail",
			Body: dynamic.ExternalActionBody{
				ReadMode:    "eager",
				Format:      "json",
				ContentType: "application/json",
				Template:    `{"id":{{ jsonPath .Item "$.ID" }}}`,
			},
			ForEach: &dynamic.ExternalActionForEach{
				Source:         "requestBody",
				JSONPath:       "$.data.response_body.data",
				MatchJSONPaths: map[string]string{"$.Status": `^succeed$`},
			},
			HTTP: &dynamic.ExternalActionHTTP{URL: target.URL, Method: http.MethodPost},
		}},
	}

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	compiled, err := compileConfigWithContext(ctx, config)
	require.NoError(t, err)
	handler := &middleware{config: compiled, logger: zerolog.Nop()}
	require.NoError(t, compiled.pollers[0].runCoordinated(ctx, handler))
	assert.Equal(t, int64(1), targetCalls.Load())
}

func TestPollerReleasesCoordinationAfterSourceFailure(t *testing.T) {
	var sourceCalls atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		if sourceCalls.Add(1) == 1 {
			http.Error(rw, "temporary", http.StatusBadGateway)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0,"data":{"items":[]}}`))
	}))
	t.Cleanup(source.Close)

	config := dynamic.ExternalAction{
		Pollers: []dynamic.ExternalActionPoller{{
			Name:     "retry",
			Interval: durationValue(t, time.Minute),
			Request: dynamic.ExternalActionPollerRequest{
				URL:              source.URL,
				MaxResponseBytes: 1024,
			},
			Observation:  dynamic.ExternalActionPollerObservation{Path: "/observed"},
			Coordination: memoryCoordination(t, "poller:retry"),
		}},
		Rules: []dynamic.ExternalActionRule{{
			Name:  "unused",
			Match: dynamic.ExternalActionMatch{PathRegex: `^/unused$`},
			HTTP:  &dynamic.ExternalActionHTTP{URL: source.URL},
		}},
	}

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	compiled, err := compileConfigWithContext(ctx, config)
	require.NoError(t, err)
	handler := &middleware{config: compiled, logger: zerolog.Nop()}
	require.Error(t, compiled.pollers[0].runCoordinated(ctx, handler))
	require.NoError(t, compiled.pollers[0].runCoordinated(ctx, handler))
	assert.Equal(t, int64(2), sourceCalls.Load())
}

func TestPollerRedactsSourceAuthenticationErrors(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "source-secret", req.Header.Get("Authorization"))
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusForbidden)
		_, _ = rw.Write([]byte(`{"message":"token source-secret is forbidden"}`))
	}))
	t.Cleanup(source.Close)

	config := dynamic.ExternalAction{
		Pollers: []dynamic.ExternalActionPoller{{
			Name:     "redaction",
			Interval: durationValue(t, time.Minute),
			Request: dynamic.ExternalActionPollerRequest{
				URL: source.URL,
				Auth: &dynamic.ExternalActionAuth{
					Type:   "header",
					Header: "Authorization",
					Token:  "source-secret",
				},
				ResponseValidation: &dynamic.ExternalActionHTTPResponse{
					Statuses:        []string{"200-299"},
					MessageJSONPath: "$.message",
				},
			},
			Observation:  dynamic.ExternalActionPollerObservation{Path: "/observed"},
			Coordination: memoryCoordination(t, "poller:redaction"),
		}},
		Rules: []dynamic.ExternalActionRule{{
			Name:  "unused",
			Match: dynamic.ExternalActionMatch{PathRegex: `^/unused$`},
			HTTP:  &dynamic.ExternalActionHTTP{URL: source.URL},
		}},
	}

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	compiled, err := compileConfigWithContext(ctx, config)
	require.NoError(t, err)
	handler := &middleware{config: compiled, logger: zerolog.Nop()}
	err = compiled.pollers[0].runCoordinated(ctx, handler)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "source-secret")
	assert.Contains(t, err.Error(), "<redacted>")
}

func memoryCoordination(t *testing.T, key string) *dynamic.ExternalActionDeduplication {
	t.Helper()
	return &dynamic.ExternalActionDeduplication{
		KeyTemplate: key,
		TTL:         durationValue(t, time.Minute),
		Lease:       durationValue(t, time.Minute),
		Memory:      &dynamic.ExternalActionDedupMemory{MaxEntries: 100},
	}
}

func durationValue(t *testing.T, value time.Duration) ptypes.Duration {
	t.Helper()
	var result ptypes.Duration
	require.NoError(t, result.Set(value.String()))
	return result
}
