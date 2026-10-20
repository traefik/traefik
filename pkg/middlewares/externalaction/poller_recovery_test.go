package externalaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func recoveryPollerConfig(t *testing.T, source, target string, event bool) dynamic.ExternalAction {
	t.Helper()
	rule := dynamic.ExternalActionRule{
		Name: "report", Phase: "after", Mode: "sync", FailurePolicy: "fail",
		Match:         dynamic.ExternalActionMatch{PathRegex: "^/observed$"},
		Body:          dynamic.ExternalActionBody{ReadMode: "none", Format: "json", Template: `{"id":{{ index .Item "id" | toJson }}}`},
		ResponseBody:  dynamic.ExternalActionResponseBody{ReadMode: "capture", Format: "json"},
		ForEach:       &dynamic.ExternalActionForEach{Source: "responseBody", JSONPath: "$.items"},
		HTTP:          &dynamic.ExternalActionHTTP{URL: target, Method: http.MethodPost},
		Deduplication: memoryCoordination(t, `{{ index .Item "id" }}`),
	}
	poller := dynamic.ExternalActionPoller{
		Name: "recovery", Interval: durationValue(t, time.Minute),
		Request:     dynamic.ExternalActionPollerRequest{URL: source},
		Observation: dynamic.ExternalActionPollerObservation{Path: "/observed"},
		Pagination: &dynamic.ExternalActionPollerPagination{
			PageParameter: "page", PageSizeParameter: "size", PageSize: 2, MaxPages: 4, ItemsJSONPath: "$.items",
		},
		Coordination: memoryCoordination(t, "recovery-poller"),
	}
	if event {
		poller.Observation.Event = &dynamic.ExternalActionPollerEvent{Source: "example", Principal: "observer"}
		rule.Phase = "before"
		rule.Match.PathRegex = "^/_governance/v1/events$"
		rule.Body.ReadMode = "eager"
		rule.ForEach = &dynamic.ExternalActionForEach{Source: "requestBody", JSONPath: "$.data.response_body.items"}
	}
	return dynamic.ExternalAction{SelectionMode: "all", Pollers: []dynamic.ExternalActionPoller{poller}, Rules: []dynamic.ExternalActionRule{rule}}
}

func TestPollerContinuesAfterIndependentActionFailure(t *testing.T) {
	for _, event := range []bool{false, true} {
		for _, failure := range []string{"target", "enrichment"} {
			t.Run(fmt.Sprintf("event=%t/%s", event, failure), func(t *testing.T) {
				var recovered atomic.Bool
				var sourceCalls atomic.Int64
				source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sourceCalls.Add(1)
					w.Header().Set("Content-Type", "application/json")
					if r.URL.Query().Get("page") == "1" {
						_, _ = w.Write([]byte(`{"items":[{"id":1},{"id":2}]}`))
					} else {
						_, _ = w.Write([]byte(`{"items":[{"id":3}]}`))
					}
				}))
				t.Cleanup(source.Close)
				var mu sync.Mutex
				sent := map[int]int{}
				target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body struct {
						ID int `json:"id"`
					}
					if json.NewDecoder(r.Body).Decode(&body) != nil {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					if failure == "target" && body.ID == 1 && !recovered.Load() {
						w.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					mu.Lock()
					sent[body.ID]++
					mu.Unlock()
					w.WriteHeader(http.StatusNoContent)
				}))
				t.Cleanup(target.Close)
				config := recoveryPollerConfig(t, source.URL, target.URL, event)
				compiled, err := compileConfigWithContext(t.Context(), config)
				require.NoError(t, err)
				if failure == "enrichment" {
					lookup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Query().Get("id") == "1" && !recovered.Load() {
							w.WriteHeader(http.StatusServiceUnavailable)
							return
						}
						_, _ = w.Write([]byte(`{"region":"example"}`))
					}))
					t.Cleanup(lookup.Close)
					enrichment, err := compileEnrichment("lookup", enrichmentFixture(lookup.URL))
					require.NoError(t, err)
					compiled.rules[0].target.(*httpTarget).enrichments = []*compiledEnrichment{enrichment}
				}
				handler := &middleware{config: compiled, logger: zerolog.Nop()}
				for range 2 {
					require.Error(t, compiled.pollers[0].runCoordinated(t.Context(), handler))
				}
				assert.Equal(t, int64(4), sourceCalls.Load())
				mu.Lock()
				assert.Equal(t, map[int]int{2: 1, 3: 1}, sent)
				mu.Unlock()
				recovered.Store(true)
				require.NoError(t, compiled.pollers[0].runCoordinated(t.Context(), handler))
				require.NoError(t, compiled.pollers[0].runCoordinated(t.Context(), handler))
				assert.Equal(t, int64(6), sourceCalls.Load())
				mu.Lock()
				assert.Equal(t, map[int]int{1: 1, 2: 1, 3: 1}, sent)
				mu.Unlock()
			})
		}
	}
}

func TestPollerStopsOnSourceFailureAfterActionFailure(t *testing.T) {
	var pages atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages.Add(1)
		if r.URL.Query().Get("page") != "1" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"id":1},{"id":2}]}`))
	}))
	t.Cleanup(source.Close)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(target.Close)
	compiled, err := compileConfigWithContext(t.Context(), recoveryPollerConfig(t, source.URL, target.URL, true))
	require.NoError(t, err)
	handler := &middleware{config: compiled, logger: zerolog.Nop()}
	err = compiled.pollers[0].runCoordinated(t.Context(), handler)
	require.ErrorContains(t, err, "source returned status 503")
	assert.ErrorContains(t, err, "observation failures")
	assert.Equal(t, int64(2), pages.Load())
}

func TestPollerCancellationStopsItemsAndPages(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var pages, actions atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		pages.Add(1)
		_, _ = w.Write([]byte(`{"items":[{"id":1},{"id":2}]}`))
	}))
	t.Cleanup(source.Close)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		actions.Add(1)
		cancel()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(target.Close)
	compiled, err := compileConfigWithContext(ctx, recoveryPollerConfig(t, source.URL, target.URL, true))
	require.NoError(t, err)
	handler := &middleware{config: compiled, logger: zerolog.Nop()}
	require.ErrorIs(t, compiled.pollers[0].runCoordinated(ctx, handler), context.Canceled)
	assert.Equal(t, int64(1), pages.Load())
	assert.Equal(t, int64(1), actions.Load())
}

func TestObservationFailuresAreBounded(t *testing.T) {
	first := errors.New("first failure")
	var failures observationFailures
	require.NoError(t, failures.err())
	failures.add(first)
	for i := range 10000 {
		failures.add(errors.New(strconv.Itoa(i)))
	}
	require.ErrorIs(t, failures.err(), first)
	assert.ErrorContains(t, failures.err(), "10001 observation failures")
	assert.Same(t, first, failures.first)
}

func TestPollerNonPaginatedFailurePolicy(t *testing.T) {
	for _, event := range []bool{false, true} {
		for _, policy := range []string{"fail", "ignore"} {
			t.Run(fmt.Sprintf("event=%t/%s", event, policy), func(t *testing.T) {
				source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte(`{"items":[{"id":1},{"id":2}]}`))
				}))
				t.Cleanup(source.Close)
				var calls atomic.Int64
				target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					calls.Add(1)
					w.WriteHeader(http.StatusServiceUnavailable)
				}))
				t.Cleanup(target.Close)
				config := recoveryPollerConfig(t, source.URL, target.URL, event)
				config.Pollers[0].Pagination = nil
				config.Rules[0].FailurePolicy = policy
				compiled, err := compileConfigWithContext(t.Context(), config)
				require.NoError(t, err)
				err = compiled.pollers[0].runCoordinated(t.Context(), &middleware{config: compiled, logger: zerolog.Nop()})
				if policy == "ignore" {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
				assert.Equal(t, int64(2), calls.Load())
			})
		}
	}
}

func TestPollerFailurePreservesRuleSelection(t *testing.T) {
	for _, selection := range []string{"first", "all"} {
		t.Run(selection, func(t *testing.T) {
			source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"items":[{"id":1}]}`))
			}))
			t.Cleanup(source.Close)
			var firstCalls, secondCalls atomic.Int64
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/first" {
					firstCalls.Add(1)
					w.WriteHeader(http.StatusServiceUnavailable)
				} else {
					secondCalls.Add(1)
					w.WriteHeader(http.StatusNoContent)
				}
			}))
			t.Cleanup(target.Close)
			config := recoveryPollerConfig(t, source.URL, target.URL+"/first", true)
			config.SelectionMode = selection
			second := config.Rules[0]
			second.Name = "second"
			second.Deduplication = nil
			second.HTTP = &dynamic.ExternalActionHTTP{URL: target.URL + "/second"}
			config.Rules = append(config.Rules, second)
			compiled, err := compileConfigWithContext(t.Context(), config)
			require.NoError(t, err)
			require.Error(t, compiled.pollers[0].runCoordinated(t.Context(), &middleware{config: compiled, logger: zerolog.Nop()}))
			assert.Equal(t, int64(1), firstCalls.Load())
			if selection == "first" {
				assert.Zero(t, secondCalls.Load())
			} else {
				assert.Equal(t, int64(1), secondCalls.Load())
			}
		})
	}
}
