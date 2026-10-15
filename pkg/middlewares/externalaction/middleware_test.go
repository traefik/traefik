package externalaction

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

func TestMiddlewareTransformsRequestBody(t *testing.T) {
	var targetBody string
	var targetHeader string
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		value, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		targetBody = string(value)
		targetHeader = req.Header.Get("X-Action-Token")
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(targetServer.Close)

	var forwardedBody string
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		value, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		forwardedBody = string(value)
		rw.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	handler, err := New(ctx, next, dynamic.ExternalAction{
		Rules: []dynamic.ExternalActionRule{
			{
				Name:  "transform",
				Phase: "before",
				Mode:  "sync",
				Match: dynamic.ExternalActionMatch{
					Methods:   []string{http.MethodPost},
					PathRegex: `^/source$`,
				},
				Body: dynamic.ExternalActionBody{
					MaxBytes:    1024,
					Format:      "json",
					ContentType: "application/json",
					Template:    `{"id":{{ jsonPath .Body "$.items[0].id" | toJson }},"operator":{{ coalesce (header . "X-Authenticated-User") (cookie . "account") | toJson }}}`,
				},
				HTTP: &dynamic.ExternalActionHTTP{
					URL:    targetServer.URL,
					Method: http.MethodPost,
					Auth: &dynamic.ExternalActionAuth{
						Type:       "headers",
						Parameters: map[string]string{"X-Action-Token": "secret"},
					},
				},
			},
		},
	}, "external-action")
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "http://example.com/source", strings.NewReader(`{"items":[{"id":"deploy-1"}]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Cookie", "account=alice")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"id":"deploy-1","operator":"alice"}`, targetBody)
	assert.Equal(t, "secret", targetHeader)
	assert.JSONEq(t, `{"items":[{"id":"deploy-1"}]}`, forwardedBody)
}

func TestExecTargetTemplatesArgsEnvAndStdinWithoutShell(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "http://example.com/source?mode=verify", strings.NewReader(`{"resource":"resource;false"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Actor", "alice")
	data := newActionData(request, []byte(`{"resource":"resource;false"}`), http.StatusOK, http.Header{}, time.Now(), time.Now())

	rule := dynamic.ExternalActionRule{Exec: &dynamic.ExternalActionExec{
		Command: os.Args[0],
		Args: []string{
			"-test.run=^TestExternalActionExecHelperProcess$",
			"--",
			`{{ jsonPath .Body "$.resource" }}`,
			`{{ header . "X-Actor" }}`,
			`{{ query . "mode" }}`,
		},
		Env: map[string]string{
			"EXTERNAL_ACTION_HELPER": "1",
			"EXPECTED_RESOURCE":      `{{ jsonPath .Body "$.resource" }}`,
		},
		Stdin: `{{ .RawBody }}`,
	}}
	target, err := newExecTarget(rule)
	require.NoError(t, err)
	require.NoError(t, target.Run(t.Context(), data))
}

func TestExternalActionExecHelperProcess(t *testing.T) {
	if os.Getenv("EXTERNAL_ACTION_HELPER") != "1" {
		return
	}
	separator := -1
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index
			break
		}
	}
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil || separator < 0 || len(os.Args) != separator+4 ||
		os.Args[separator+1] != "resource;false" ||
		os.Args[separator+2] != "alice" ||
		os.Args[separator+3] != "verify" ||
		os.Getenv("EXPECTED_RESOURCE") != "resource;false" ||
		string(stdin) != `{"resource":"resource;false"}` {
		os.Exit(2)
	}
	os.Exit(0)
}

func TestMiddlewareRoutesRequestBodyToDifferentProjects(t *testing.T) {
	targetProjects := make(chan string, 3)
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		targetProjects <- req.Header.Get("projectid") + ":" + req.Header.Get("token")
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(targetServer.Close)

	backendCalls := 0
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	handler, err := New(ctx, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, readErr := io.ReadAll(req.Body)
		require.NoError(t, readErr)
		backendCalls++
		rw.WriteHeader(http.StatusOK)
	}), dynamic.ExternalAction{
		SelectionMode: "first",
		Rules: []dynamic.ExternalActionRule{
			projectRule("capacity-project", `^CAPACITY_REDUCTION$`, "1001", targetServer.URL),
			projectRule("migration-project", `^RESOURCE_MIGRATION$`, "2002", targetServer.URL),
			{
				Name:  "default-project",
				Phase: "before",
				Mode:  "sync",
				Match: dynamic.ExternalActionMatch{PathRegex: `^/tasks/create$`},
				Body: dynamic.ExternalActionBody{
					ReadMode: "eager",
					MaxBytes: 1024,
					Format:   "json",
					Template: `{}`,
				},
				HTTP: &dynamic.ExternalActionHTTP{
					URL:    targetServer.URL,
					Method: http.MethodPost,
					Auth: &dynamic.ExternalActionAuth{
						Type:       "headers",
						Parameters: map[string]string{"projectid": "default", "token": "token-default"},
					},
				},
			},
		},
	}, "external-action")
	require.NoError(t, err)

	for _, test := range []struct {
		taskType string
		project  string
	}{
		{taskType: "CAPACITY_REDUCTION", project: "1001:token-1001"},
		{taskType: "RESOURCE_MIGRATION", project: "2002:token-2002"},
		{taskType: "OTHER", project: "default:token-default"},
	} {
		request := httptest.NewRequest(http.MethodPost, "http://example.com/tasks/create",
			strings.NewReader(`{"task_type":"`+test.taskType+`"}`))
		request.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(httptest.NewRecorder(), request)

		select {
		case project := <-targetProjects:
			assert.Equal(t, test.project, project)
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for project %s", test.project)
		}
	}
	assert.Equal(t, 3, backendCalls)
	assert.Empty(t, targetProjects, "selectionMode=first must dispatch only one matching project")
}

func projectRule(name, taskType, projectID, targetURL string) dynamic.ExternalActionRule {
	return dynamic.ExternalActionRule{
		Name:  name,
		Phase: "before",
		Mode:  "sync",
		Match: dynamic.ExternalActionMatch{
			PathRegex:     `^/tasks/create$`,
			BodyJSONPaths: map[string]string{"$.task_type": taskType},
		},
		Body: dynamic.ExternalActionBody{
			ReadMode: "eager",
			MaxBytes: 1024,
			Format:   "json",
			Template: `{}`,
		},
		HTTP: &dynamic.ExternalActionHTTP{
			URL:    targetURL,
			Method: http.MethodPost,
			Auth: &dynamic.ExternalActionAuth{
				Type: "headers",
				Parameters: map[string]string{
					"projectid": projectID,
					"token":     "token-" + projectID,
				},
			},
		},
	}
}

func TestMiddlewareAfterStatusFilter(t *testing.T) {
	called := make(chan struct{}, 1)
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		called <- struct{}{}
		rw.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(targetServer.Close)

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	handler, err := New(ctx, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusCreated)
	}), dynamic.ExternalAction{
		Queue: dynamic.ExternalActionQueue{Capacity: 4, Workers: 1},
		Rules: []dynamic.ExternalActionRule{
			{
				Name:  "after",
				Phase: "after",
				Mode:  "async",
				Match: dynamic.ExternalActionMatch{
					ResponseStatuses: []string{"200-299"},
				},
				Body: dynamic.ExternalActionBody{ReadMode: "none"},
				HTTP: &dynamic.ExternalActionHTTP{URL: targetServer.URL, Method: http.MethodPost},
			},
		},
	}, "external-action")
	require.NoError(t, err)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com/", nil))

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("external action was not called")
	}
}

func TestMiddlewareExpandsAndDeduplicatesResponseItems(t *testing.T) {
	var targetBodies []string
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		value, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		targetBodies = append(targetBodies, string(value))
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(targetServer.Close)

	backendBody := `{"success":true,"items":[{"id":"task-1","status":"running"},{"id":"task-2","status":"success"},{"id":"task-3","status":"failed"}]}`
	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(backendBody))
	})
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	handler, err := New(ctx, next, dynamic.ExternalAction{
		Rules: []dynamic.ExternalActionRule{
			{
				Name:  "terminal-items",
				Phase: "after",
				Mode:  "sync",
				Match: dynamic.ExternalActionMatch{
					PathRegex:             `^/tasks$`,
					ResponseStatuses:      []string{"200-299"},
					ResponseBodyJSONPaths: map[string]string{"$.success": `^true$`},
				},
				Body: dynamic.ExternalActionBody{
					ReadMode:    "none",
					ContentType: "application/json",
					Template:    `{"id":{{ jsonPath .Item "$.id" | toJson }},"status":{{ jsonPath .Item "$.status" | toJson }},"count":{{ len (jsonPath .Response.Body "$.items") }}}`,
				},
				ResponseBody: dynamic.ExternalActionResponseBody{
					ReadMode: "capture",
					MaxBytes: 1024,
					Format:   "json",
					Overflow: "skip",
				},
				ForEach: &dynamic.ExternalActionForEach{
					Source:         "responseBody",
					JSONPath:       "$.items",
					MatchJSONPaths: map[string]string{"$.status": `^(success|failed)$`},
				},
				Deduplication: &dynamic.ExternalActionDeduplication{
					KeyTemplate: `terminal:{{ jsonPath .Item "$.id" }}:{{ jsonPath .Item "$.status" }}`,
					Memory:      &dynamic.ExternalActionDedupMemory{MaxEntries: 100},
				},
				HTTP: &dynamic.ExternalActionHTTP{URL: targetServer.URL, Method: http.MethodPost},
			},
		},
	}, "external-action")
	require.NoError(t, err)

	for range 2 {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/tasks", nil))
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.JSONEq(t, backendBody, recorder.Body.String())
	}

	require.Len(t, targetBodies, 2, "the second poll must not report the same terminal items again")
	assert.JSONEq(t, `{"id":"task-2","status":"success","count":3}`, targetBodies[0])
	assert.JSONEq(t, `{"id":"task-3","status":"failed","count":3}`, targetBodies[1])
}

func TestMiddlewareDecodesCompressedResponseBody(t *testing.T) {
	encoders := map[string]func(*testing.T, []byte) []byte{
		"gzip": func(t *testing.T, value []byte) []byte {
			t.Helper()
			var result bytes.Buffer
			writer := gzip.NewWriter(&result)
			_, err := writer.Write(value)
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			return result.Bytes()
		},
		"br": func(t *testing.T, value []byte) []byte {
			t.Helper()
			var result bytes.Buffer
			writer := brotli.NewWriter(&result)
			_, err := writer.Write(value)
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			return result.Bytes()
		},
		"zstd": func(t *testing.T, value []byte) []byte {
			t.Helper()
			var result bytes.Buffer
			writer, err := zstd.NewWriter(&result)
			require.NoError(t, err)
			_, err = writer.Write(value)
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			return result.Bytes()
		},
	}

	backendBody := []byte(`{"code":0,"data":{"execution_id":"exec-1","status":"success"}}`)
	for encoding, encode := range encoders {
		t.Run(encoding, func(t *testing.T) {
			encodedBody := encode(t, backendBody)
			var targetBody string
			targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				value, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				targetBody = string(value)
				rw.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(targetServer.Close)

			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.Header().Set("Content-Type", "application/json")
				rw.Header().Set("Content-Encoding", encoding)
				_, _ = rw.Write(encodedBody)
			})
			ctx, cancel := context.WithCancel(t.Context())
			t.Cleanup(cancel)
			handler, err := New(ctx, next, dynamic.ExternalAction{
				Rules: []dynamic.ExternalActionRule{{
					Name: "compressed-terminal-response", Phase: "after", Mode: "sync",
					Match: dynamic.ExternalActionMatch{
						ResponseStatuses:      []string{"200-299"},
						ResponseBodyJSONPaths: map[string]string{"$.code": `^0$`, "$.data.status": `^success$`},
					},
					Body:         dynamic.ExternalActionBody{ReadMode: "none", ContentType: "application/json", Template: `{"executionID":{{ jsonPath .Response.Body "$.data.execution_id" | toJson }}}`},
					ResponseBody: dynamic.ExternalActionResponseBody{ReadMode: "capture", MaxBytes: 4096, Format: "json", Overflow: "skip"},
					HTTP:         &dynamic.ExternalActionHTTP{URL: targetServer.URL, Method: http.MethodPost},
				}},
			}, "external-action")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/jobs/exec-1", nil))

			assert.Equal(t, encoding, recorder.Header().Get("Content-Encoding"))
			assert.Equal(t, encodedBody, recorder.Body.Bytes(), "capturing must not alter the bytes delivered to the client")
			assert.JSONEq(t, `{"executionID":"exec-1"}`, targetBody)
		})
	}
}

func TestDecodeCapturedResponseBodyEnforcesExpandedLimit(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, err := writer.Write([]byte(strings.Repeat("x", 4096)))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	require.Less(t, compressed.Len(), 128)

	decoded, overflow, err := decodeCapturedResponseBody("gzip", compressed.Bytes(), 128)
	require.NoError(t, err)
	assert.True(t, overflow)
	assert.Len(t, decoded, 128)
}

func TestMiddlewareRunsStatusOnlyAfterRuleWhenCapturedResponseCannotBeDecoded(t *testing.T) {
	targetCalls := 0
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		targetCalls++
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(targetServer.Close)

	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Encoding", "unsupported")
		_, _ = rw.Write([]byte(`not-json`))
	})
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	handler, err := New(ctx, next, dynamic.ExternalAction{
		Rules: []dynamic.ExternalActionRule{
			{
				Name: "capture-body", Phase: "after", Mode: "sync",
				ResponseBody: dynamic.ExternalActionResponseBody{ReadMode: "capture", MaxBytes: 4096, Format: "json", Overflow: "skip"},
				Match:        dynamic.ExternalActionMatch{ResponseBodyJSONPaths: map[string]string{"$.code": `^0$`}},
				HTTP:         &dynamic.ExternalActionHTTP{URL: targetServer.URL, Method: http.MethodPost},
			},
			{
				Name: "status-only", Phase: "after", Mode: "sync",
				Match: dynamic.ExternalActionMatch{ResponseStatuses: []string{"200-299"}},
				HTTP:  &dynamic.ExternalActionHTTP{URL: targetServer.URL, Method: http.MethodPost},
			},
		},
	}, "external-action")
	require.NoError(t, err)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com/status-only", nil))
	assert.Equal(t, 1, targetCalls)
}

func TestMiddlewareReleasesDeduplicationClaimAfterTargetFailure(t *testing.T) {
	targetCalls := 0
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		targetCalls++
		if targetCalls == 1 {
			http.Error(rw, "temporary failure", http.StatusServiceUnavailable)
			return
		}
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(targetServer.Close)

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	handler, err := New(ctx, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"id":"task-1","status":"success"}`))
	}), dynamic.ExternalAction{
		Rules: []dynamic.ExternalActionRule{
			{
				Name:          "retry-report",
				Phase:         "after",
				Mode:          "sync",
				FailurePolicy: "ignore",
				Match: dynamic.ExternalActionMatch{
					ResponseBodyJSONPaths: map[string]string{"$.status": `^success$`},
				},
				Body: dynamic.ExternalActionBody{ReadMode: "none", Template: `{}`},
				ResponseBody: dynamic.ExternalActionResponseBody{
					ReadMode: "capture",
					Format:   "json",
				},
				Deduplication: &dynamic.ExternalActionDeduplication{
					KeyTemplate: `{{ jsonPath .Response.Body "$.id" }}`,
					Memory:      &dynamic.ExternalActionDedupMemory{MaxEntries: 10},
				},
				HTTP: &dynamic.ExternalActionHTTP{URL: targetServer.URL},
			},
		},
	}, "external-action")
	require.NoError(t, err)

	for range 2 {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com/task", nil))
	}
	assert.Equal(t, 2, targetCalls, "a failed external call must release its claim for a later poll")
}

func TestMiddlewareSkipsResponseBodyThatExceedsRuleLimit(t *testing.T) {
	targetCalls := 0
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		targetCalls++
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(targetServer.Close)

	responseBody := `{"value":"` + strings.Repeat("x", 128) + `"}`
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	handler, err := New(ctx, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(responseBody))
	}), dynamic.ExternalAction{
		Rules: []dynamic.ExternalActionRule{
			{
				Name:  "bounded-response",
				Phase: "after",
				Mode:  "sync",
				Match: dynamic.ExternalActionMatch{
					ResponseBodyJSONPaths: map[string]string{"$.value": `.+`},
				},
				Body: dynamic.ExternalActionBody{ReadMode: "none", Template: `{}`},
				ResponseBody: dynamic.ExternalActionResponseBody{
					ReadMode: "capture",
					MaxBytes: 32,
					Format:   "json",
					Overflow: "skip",
				},
				HTTP: &dynamic.ExternalActionHTTP{URL: targetServer.URL},
			},
		},
	}, "external-action")
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/task", nil))

	assert.Equal(t, responseBody, recorder.Body.String(), "capture must not alter the backend response")
	assert.Zero(t, targetCalls)
}

func TestTemplateRequired(t *testing.T) {
	value, err := newTemplate("required", `{{ required "missing flow ID" (jsonPath .Body "$.flow.id") }}`)
	require.NoError(t, err)

	_, err = value.render(&actionData{Body: map[string]any{}})
	require.ErrorContains(t, err, "missing flow ID")
}

func TestJSONNumbersRetainPrecision(t *testing.T) {
	for _, number := range []string{"9007199254740993", "18446744073709551615", "-9007199254740993", "1.234567890123456789", "1e30"} {
		t.Run(number, func(t *testing.T) {
			raw := []byte(`{"id":` + number + `}`)
			body := parseBody("json", "application/json", raw)
			assert.Equal(t, json.Number(number), selectJSON(body, "$.id"))
			rule := &compiledRule{bodyJSONPaths: map[string]*regexp.Regexp{"$.id": regexp.MustCompile("^" + regexp.QuoteMeta(number) + "$")}}
			assert.True(t, rule.matchesBody("application/json", raw))
			value, err := newTemplate("number", `{{ jsonPath .Body "$.id" | toJson }}`)
			require.NoError(t, err)
			actual, err := value.render(&actionData{Body: body})
			require.NoError(t, err)
			assert.Equal(t, number, actual)
		})
	}
	for _, raw := range []string{`{"id":1} {"id":2}`, `{"id":1} trailing`, `{"id":`} {
		assert.Equal(t, raw, parseBody("json", "application/json", []byte(raw)))
	}
}

func TestJSONPathSupportsArrayIndexesAndQuotedMapKeys(t *testing.T) {
	payload := map[string]any{
		"items": []any{
			map[string]any{
				"labels": map[string]any{
					"automation.example.com/change-id": "12345",
					"single'quote":                     "supported",
				},
			},
		},
	}

	assert.Equal(t, "12345", selectJSON(payload, `$.items[0].labels["automation.example.com/change-id"]`))
	assert.Equal(t, "supported", selectJSON(payload, `$.items[0].labels['single\'quote']`))
	assert.Nil(t, selectJSON(payload, `$.items[*].labels`))
	assert.Nil(t, selectJSON(payload, `$.items[0].labels["unterminated]`))
}

func TestHTTPTargetValidatesApplicationResponse(t *testing.T) {
	testCases := []struct {
		desc          string
		response      string
		expectedError string
	}{
		{
			desc:     "success",
			response: `{"code":0,"msg":null,"data":{"status_error":false}}`,
		},
		{
			desc:          "business error carried by HTTP 200",
			response:      `{"code":5000,"msg":"OperationType is not registered","data":null}`,
			expectedError: `external action response validation failed at $.code (value "5000"): OperationType is not registered`,
		},
		{
			desc:          "missing success field",
			response:      `{"msg":"missing code"}`,
			expectedError: `external action response validation failed at $.code (value "<missing>"): missing code`,
		},
		{
			desc:          "invalid JSON",
			response:      `not-json`,
			expectedError: "external action response is not valid JSON",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.Header().Set("Content-Type", "application/json")
				_, _ = rw.Write([]byte(test.response))
			}))
			t.Cleanup(targetServer.Close)

			actionTarget, err := newHTTPTarget(dynamic.ExternalActionRule{
				Body: dynamic.ExternalActionBody{Template: `{}`},
				HTTP: &dynamic.ExternalActionHTTP{
					URL: targetServer.URL,
					ResponseValidation: &dynamic.ExternalActionHTTPResponse{
						Statuses:        []string{"200-299"},
						SuccessJSONPath: "$.code",
						SuccessValues:   []string{"0"},
						MessageJSONPath: "$.msg",
					},
				},
			})
			require.NoError(t, err)

			err = actionTarget.Run(t.Context(), &actionData{})
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.expectedError)
			}
		})
	}
}

func TestResponseValidationConfiguration(t *testing.T) {
	testCases := []struct {
		desc     string
		response *dynamic.ExternalActionHTTPResponse
		expected string
	}{
		{
			desc:     "missing success values",
			response: &dynamic.ExternalActionHTTPResponse{SuccessJSONPath: "$.code"},
			expected: "successValues is required",
		},
		{
			desc:     "missing success JSON path",
			response: &dynamic.ExternalActionHTTPResponse{SuccessValues: []string{"0"}},
			expected: "successJSONPath is required",
		},
		{
			desc: "ambiguous body validators",
			response: &dynamic.ExternalActionHTTPResponse{
				SuccessJSONPath: "$.code",
				SuccessValues:   []string{"0"},
				BodyRegex:       `"code":0`,
			},
			expected: "mutually exclusive",
		},
		{
			desc:     "invalid response status",
			response: &dynamic.ExternalActionHTTPResponse{Statuses: []string{"invalid"}},
			expected: "parsing HTTP response validation statuses",
		},
		{
			desc:     "invalid body regex",
			response: &dynamic.ExternalActionHTTPResponse{BodyRegex: `(`},
			expected: "compiling HTTP response validation bodyRegex",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			_, err := compileConfig(dynamic.ExternalAction{
				Rules: []dynamic.ExternalActionRule{{
					Body: dynamic.ExternalActionBody{Template: `{}`},
					HTTP: &dynamic.ExternalActionHTTP{
						URL:                "http://example.com",
						ResponseValidation: test.response,
					},
				}},
			})
			require.ErrorContains(t, err, test.expected)
		})
	}
}

func TestSelectionModeConfiguration(t *testing.T) {
	_, err := compileConfig(dynamic.ExternalAction{
		SelectionMode: "unsupported",
		Rules: []dynamic.ExternalActionRule{{
			Body: dynamic.ExternalActionBody{Template: `{}`},
			HTTP: &dynamic.ExternalActionHTTP{URL: "http://example.com"},
		}},
	})
	require.ErrorContains(t, err, `unsupported external action selection mode "unsupported"`)
}

func TestHTTPTargetUsesAuthParametersInBodyWithoutHardCodingProject(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, "dynamic-project", req.Header.Get("projectname"))
		require.Equal(t, "10086", req.Header.Get("projectid"))

		var body map[string]any
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.InDelta(t, float64(10086), body["ChangeProjectId"], 0)
		require.Equal(t, "dynamic-project", body["ChangeProjectName"])
		require.Equal(t, "20020001", body["FlowId"])
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(targetServer.Close)

	actionTarget, err := newHTTPTarget(dynamic.ExternalActionRule{
		Body: dynamic.ExternalActionBody{
			ContentType: "application/json",
			Template: `{
				"ChangeProjectId": {{ parseInt (required "projectid required" (authParameter . "projectid")) }},
				"ChangeProjectName": {{ required "projectname required" (authParameter . "projectname") | toJson }},
				"FlowId": {{ scalarString (jsonPath .Response.Body "$.flow_id") | toJson }}
			}`,
		},
		HTTP: &dynamic.ExternalActionHTTP{
			URL: targetServer.URL,
			Auth: &dynamic.ExternalActionAuth{
				Type: "headers",
				Parameters: map[string]string{
					"projectid":   "10086",
					"projectname": "dynamic-project",
				},
			},
		},
	})
	require.NoError(t, err)
	require.NoError(t, actionTarget.Run(t.Context(), &actionData{
		Response: actionResponse{
			Body: map[string]any{"flow_id": float64(20020001)},
		},
	}))
}

func TestHTTPTargetRejectsInvalidIntegerAuthParameterBeforeCallingTarget(t *testing.T) {
	called := false
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		called = true
		rw.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(targetServer.Close)

	actionTarget, err := newHTTPTarget(dynamic.ExternalActionRule{
		Body: dynamic.ExternalActionBody{
			ContentType: "application/json",
			Template:    `{"project":{{ parseInt (authParameter . "projectid") }}}`,
		},
		HTTP: &dynamic.ExternalActionHTTP{
			URL: targetServer.URL,
			Auth: &dynamic.ExternalActionAuth{
				Type:       "headers",
				Parameters: map[string]string{"projectid": "not-an-integer"},
			},
		},
	})
	require.NoError(t, err)

	err = actionTarget.Run(t.Context(), &actionData{})
	require.ErrorContains(t, err, "invalid syntax")
	require.False(t, called)
}

func TestHTTPTargetRedactsAuthenticationValuesFromResponseErrors(t *testing.T) {
	const secret = "sensitive-token-value"
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`{"code":190002,"msg":"token [` + secret + `] cannot access project"}`))
	}))
	t.Cleanup(targetServer.Close)

	actionTarget, err := newHTTPTarget(dynamic.ExternalActionRule{
		Body: dynamic.ExternalActionBody{Template: `{}`},
		HTTP: &dynamic.ExternalActionHTTP{
			URL: targetServer.URL,
			Auth: &dynamic.ExternalActionAuth{
				Type:       "headers",
				Parameters: map[string]string{"token": secret},
			},
			ResponseValidation: &dynamic.ExternalActionHTTPResponse{
				SuccessJSONPath: "$.code",
				SuccessValues:   []string{"0"},
				MessageJSONPath: "$.msg",
			},
		},
	})
	require.NoError(t, err)

	err = actionTarget.Run(t.Context(), &actionData{})
	require.Error(t, err)
	require.NotContains(t, err.Error(), secret)
	require.ErrorContains(t, err, "token [<redacted>] cannot access project")
}

func TestHTTPTargetLogsBoundedRedactedNonJSONError(t *testing.T) {
	const secret = "sensitive-token-value"
	targetServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write([]byte("invalid request for token " + secret))
	}))
	t.Cleanup(targetServer.Close)

	actionTarget, err := newHTTPTarget(dynamic.ExternalActionRule{
		Body: dynamic.ExternalActionBody{Template: `{}`},
		HTTP: &dynamic.ExternalActionHTTP{
			URL: targetServer.URL,
			Auth: &dynamic.ExternalActionAuth{
				Type:       "headers",
				Parameters: map[string]string{"token": secret},
			},
			ResponseValidation: &dynamic.ExternalActionHTTPResponse{
				Statuses:        []string{"200-299"},
				SuccessJSONPath: "$.code",
				SuccessValues:   []string{"0"},
				MessageJSONPath: "$.msg",
			},
		},
	})
	require.NoError(t, err)

	err = actionTarget.Run(t.Context(), &actionData{})
	require.EqualError(t, err, "external action returned status 400: invalid request for token <redacted>")
}

func TestRuntimeCredentialRotationKeepsImmutableIdentity(t *testing.T) {
	newConfig := func(token, projectID, projectSource string) dynamic.ExternalAction {
		return dynamic.ExternalAction{
			ConfigurationID: t.Name(),
			RuntimeValues: map[string]dynamic.RuntimeValue{
				"outboundToken": {
					Source:   "secret:test/runtime:token",
					Revision: runtimeconfig.StableKey(token),
					Value:    token,
				},
				"destinationID": {
					Source:   projectSource,
					Revision: runtimeconfig.StableKey(projectID),
					Value:    projectID,
				},
				// An unrelated runtime value must not change the stable
				// credential-set cache identity.
				"unrelatedSDKSetting": {
					Source: "configmap:test/sdk:setting",
					Value:  "enabled",
				},
			},
			CredentialSets: map[string]dynamic.ExternalActionCredentialSet{
				"stable-destination": {
					ImmutableParameters: []string{"projectid"},
					Auth: dynamic.ExternalActionAuth{
						Type: "headers",
						Parameters: map[string]string{
							"token":     "runtime://outboundToken",
							"projectid": "runtime://destinationID",
						},
					},
				},
			},
			Rules: []dynamic.ExternalActionRule{{
				Name: "test",
				Body: dynamic.ExternalActionBody{Template: `{}`},
				HTTP: &dynamic.ExternalActionHTTP{
					URL:  "http://example.test",
					Auth: &dynamic.ExternalActionAuth{CredentialSet: "stable-destination"},
				},
			}},
		}
	}

	first, err := compileConfigWithContext(t.Context(), newConfig("token-a", "1001", "secret:test/runtime:project"))
	require.NoError(t, err)
	firstCredential := first.credentials["stable-destination"]
	_, firstBound, _, warning, err := firstCredential.bind(&actionData{})
	require.NoError(t, err)
	require.NoError(t, warning)
	require.Equal(t, "token-a", firstBound.parameters["token"])

	rotated, err := compileConfigWithContext(t.Context(), newConfig("token-b", "1001", "secret:test/runtime:project"))
	require.NoError(t, err)
	rotatedCredential := rotated.credentials["stable-destination"]
	require.Same(t, firstCredential, rotatedCredential)
	_, rotatedBound, _, warning, err := rotatedCredential.bind(&actionData{})
	require.NoError(t, err)
	require.NoError(t, warning)
	require.Equal(t, "token-b", rotatedBound.parameters["token"])
	require.Equal(t, "1001", rotatedBound.parameters["projectid"])

	// Even changing the referenced Secret/key cannot repurpose the same
	// credentialSet name for another immutable destination.
	rejected, err := compileConfigWithContext(t.Context(), newConfig("token-c", "2002", "secret:test/another-runtime:another-project"))
	require.NoError(t, err)
	rejectedCredential := rejected.credentials["stable-destination"]
	require.Same(t, firstCredential, rejectedCredential)
	_, rejectedBound, _, warning, err := rejectedCredential.bind(&actionData{})
	require.NoError(t, err)
	require.NoError(t, warning)
	require.Equal(t, "token-b", rejectedBound.parameters["token"])
	require.Equal(t, "1001", rejectedBound.parameters["projectid"])
}

func TestDispatcherEnforcesMemoryBudget(t *testing.T) {
	dispatch := &dispatcher{
		queue:          make(chan actionTask, 1),
		done:           make(chan struct{}),
		maxQueuedBytes: 4096,
	}
	task := actionTask{data: &actionData{RawBody: strings.Repeat("x", 512)}}

	require.True(t, dispatch.enqueue(task))
	queued := dispatch.queuedBytes.Load()
	require.Positive(t, queued)
	require.False(t, dispatch.enqueue(task))
	assert.Equal(t, queued, dispatch.queuedBytes.Load())
}

func TestDispatcherDrainsAcceptedTasksAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	started := make(chan struct{})
	release := make(chan struct{})
	completed := make(chan struct{}, 2)
	target := targetFunc(func(ctx context.Context, _ *actionData) error {
		if len(completed) == 0 {
			close(started)
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		completed <- struct{}{}
		return nil
	})
	rule := &compiledRule{
		config: dynamic.ExternalActionRule{
			Name:    "drain",
			Timeout: ptypes.Duration(time.Second),
		},
		target: target,
	}
	dispatch := &dispatcher{
		ctx:            ctx,
		queue:          make(chan actionTask, 2),
		done:           make(chan struct{}),
		maxQueuedBytes: 1 << 20,
		logger:         zerolog.Nop(),
	}
	dispatch.wg.Add(1)
	go dispatch.run()

	require.True(t, dispatch.enqueue(actionTask{rule: rule, data: &actionData{}}))
	<-started
	require.True(t, dispatch.enqueue(actionTask{rule: rule, data: &actionData{}}))

	cancel()
	dispatch.stopMu.Lock()
	dispatch.stopped = true
	close(dispatch.done)
	dispatch.stopMu.Unlock()
	close(release)
	dispatch.wg.Wait()

	assert.Len(t, completed, 2)
	assert.Zero(t, dispatch.queuedBytes.Load())
	require.False(t, dispatch.enqueue(actionTask{rule: rule, data: &actionData{}}))
}

type targetFunc func(context.Context, *actionData) error

func (f targetFunc) Run(ctx context.Context, data *actionData) error {
	return f(ctx, data)
}
