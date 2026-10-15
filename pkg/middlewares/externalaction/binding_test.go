package externalaction

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestCredentialSetReloadsRotatableValuesAndRejectsIdentityChange(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	write := func(name, value string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(directory, name), []byte(value), 0o600))
	}
	write("token", "token-a-1")
	write("staffname", "caller-a-1")
	write("projectname", "project-a")
	write("projectid", "1001")

	requests := make(chan http.Header, 3)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = io.Copy(io.Discard, req.Body)
		requests <- req.Header.Clone()
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0}`))
	}))
	t.Cleanup(server.Close)

	credentialConfig := map[string]dynamic.ExternalActionCredentialSet{
		"project-a-1001": {
			ReloadInterval:      ptypes.Duration(time.Millisecond),
			ImmutableParameters: []string{"projectid", "projectname"},
			Auth: dynamic.ExternalActionAuth{
				Type: "headers",
				ParameterFiles: map[string]string{
					"token":       filepath.Join(directory, "token"),
					"staffname":   filepath.Join(directory, "staffname"),
					"projectname": filepath.Join(directory, "projectname"),
					"projectid":   filepath.Join(directory, "projectid"),
				},
			},
		},
	}
	credentials, err := compileCredentialSets(credentialConfig)
	require.NoError(t, err)

	rule := dynamic.ExternalActionRule{
		Body: dynamic.ExternalActionBody{
			Template:    `{}`,
			ContentType: "application/json",
		},
		HTTP: &dynamic.ExternalActionHTTP{
			URL: server.URL,
			Auth: &dynamic.ExternalActionAuth{
				CredentialSet: "project-a-1001",
			},
		},
	}
	actionTarget, err := newHTTPTarget(rule, credentials)
	require.NoError(t, err)

	run := func() http.Header {
		t.Helper()
		require.NoError(t, actionTarget.Run(t.Context(), &actionData{}))
		select {
		case headers := <-requests:
			return headers
		case <-time.After(time.Second):
			t.Fatal("external request was not received")
			return nil
		}
	}

	first := run()
	assert.Equal(t, "token-a-1", first.Get("token"))
	assert.Equal(t, "caller-a-1", first.Get("staffname"))
	assert.Equal(t, "1001", first.Get("projectid"))

	write("token", "token-a-2")
	write("staffname", "caller-a-2")
	time.Sleep(2 * time.Millisecond)
	second := run()
	assert.Equal(t, "token-a-2", second.Get("token"))
	assert.Equal(t, "caller-a-2", second.Get("staffname"))
	assert.Equal(t, "1001", second.Get("projectid"))

	write("token", "token-b")
	write("staffname", "caller-b")
	write("projectname", "project-b")
	write("projectid", "2002")
	time.Sleep(2 * time.Millisecond)

	// Kubernetes Secret updates can rebuild the dynamic Middleware even though
	// the CR is unchanged. Recompiling the same semantic credential set must
	// retain the last valid project-A material across that rebuild.
	rebuiltCredentials, err := compileCredentialSets(credentialConfig)
	require.NoError(t, err)
	actionTarget, err = newHTTPTarget(rule, rebuiltCredentials)
	require.NoError(t, err)

	third := run()
	assert.Equal(t, "token-a-2", third.Get("token"))
	assert.Equal(t, "caller-a-2", third.Get("staffname"))
	assert.Equal(t, "project-a", third.Get("projectname"))
	assert.Equal(t, "1001", third.Get("projectid"))
}

func TestCredentialFilesKeepLastGoodMaterialAfterUnsafeRotation(t *testing.T) {
	directory := t.TempDir()
	rotate := func(version, token string) {
		t.Helper()
		require.NoError(t, os.Mkdir(filepath.Join(directory, version), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(directory, version, "token"), []byte(token), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(directory, version, "project"), []byte("project-a"), 0o600))
		require.NoError(t, os.Symlink(version, filepath.Join(directory, "..next")))
		require.NoError(t, os.Rename(filepath.Join(directory, "..next"), filepath.Join(directory, "..data")))
	}
	for _, name := range []string{"token", "project"} {
		require.NoError(t, os.Symlink(filepath.Join("..data", name), filepath.Join(directory, name)))
	}
	rotate("one", "first-token")
	auth := &compiledAuth{
		kind: "bearer", tokenFile: filepath.Join(directory, "token"),
		parameterFiles:      map[string]string{"project": filepath.Join(directory, "project")},
		immutableParameters: []string{"project"}, reloadInterval: time.Nanosecond,
	}
	_, bound, identity, warning, err := auth.bind(&actionData{})
	require.NoError(t, err)
	require.NoError(t, warning)
	assert.Equal(t, "first-token", bound.token)
	rotate("invalid", strings.Repeat("x", (1<<20)+1))
	_, bound, nextIdentity, warning, err := auth.bind(&actionData{})
	require.NoError(t, err)
	require.Error(t, warning)
	assert.Equal(t, "first-token", bound.token)
	assert.Equal(t, identity, nextIdentity)
	rotate("two", "second-token")
	_, bound, nextIdentity, warning, err = auth.bind(&actionData{})
	require.NoError(t, err)
	require.NoError(t, warning)
	assert.Equal(t, "second-token", bound.token)
	assert.Equal(t, identity, nextIdentity)
}

func TestLifecycleBindingPinsCredentialSetAndRejectsReassignment(t *testing.T) {
	t.Parallel()

	credentials, err := compileCredentialSets(map[string]dynamic.ExternalActionCredentialSet{
		"project-a-1001": literalCredentialSet("project-a", "1001", "token-a"),
		"project-b-2002": literalCredentialSet("project-b", "2002", "token-b"),
	})
	require.NoError(t, err)

	backend := &dynamic.ExternalActionBindingMemory{MaxEntries: 100}
	createA, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
		Mode:          "create",
		KeyTemplate:   "{{ .Request.Path }}",
		CredentialSet: "project-a-1001",
		Memory:        backend,
	}, credentials)
	require.NoError(t, err)
	resolve, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
		Mode:                  "resolve",
		KeyTemplate:           "{{ .Request.Path }}",
		AllowedCredentialSets: []string{"project-a-1001", "project-b-2002"},
		Memory:                backend,
	}, credentials)
	require.NoError(t, err)

	data := &actionData{Request: actionRequest{Path: "workflow-42"}}
	pinned, err := createA.pin(data)
	require.NoError(t, err)
	_, err = createA.prepare(t.Context(), pinned)
	require.NoError(t, err)

	resolved, err := resolve.prepare(t.Context(), data)
	require.NoError(t, err)
	assert.Equal(t, "project-a-1001", resolved.credentialSet)
	assert.NotEqual(t, [32]byte{}, resolved.expectedCredentialIdentity)

	createB, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
		Mode:          "create",
		KeyTemplate:   "{{ .Request.Path }}",
		CredentialSet: "project-b-2002",
		Memory:        backend,
	}, credentials)
	require.NoError(t, err)
	pinnedB, err := createB.pin(data)
	require.NoError(t, err)
	_, err = createB.prepare(t.Context(), pinnedB)
	require.ErrorContains(t, err, "already bound")
}

func TestLifecycleBindingMatchesSourceProjectWithoutHardCodedRuleValues(t *testing.T) {
	t.Parallel()

	credentials, err := compileCredentialSets(map[string]dynamic.ExternalActionCredentialSet{
		"project-a-1001": literalCredentialSet("project-a", "1001", "token-a"),
		"project-b-2002": literalCredentialSet("project-b", "2002", "token-b"),
	})
	require.NoError(t, err)

	newBinding := func(name string) *compiledBinding {
		t.Helper()
		value, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
			Mode:          "create",
			KeyTemplate:   "{{ .Response.RawBody }}",
			CredentialSet: name,
			Match: dynamic.ExternalActionBindingMatch{
				ResponseBodyJSONPaths: map[string]string{
					"projectid":   "$.project.id",
					"projectname": "$.project.name",
				},
			},
			Memory: &dynamic.ExternalActionBindingMemory{MaxEntries: 100},
		}, credentials)
		require.NoError(t, err)
		return value
	}

	data := &actionData{
		Response: actionResponse{
			Body: map[string]any{
				"project": map[string]any{
					"id":   float64(2002),
					"name": "project-b",
				},
			},
			RawBody: `{"project":{"id":2002,"name":"project-b"}}`,
		},
	}
	matchesA, err := newBinding("project-a-1001").matches(data)
	require.NoError(t, err)
	assert.False(t, matchesA)

	matchesB, err := newBinding("project-b-2002").matches(data)
	require.NoError(t, err)
	assert.True(t, matchesB)
}

func TestLifecycleBindingSelectsOneAllowedCredentialSetPerExpandedItem(t *testing.T) {
	t.Parallel()

	credentials, err := compileCredentialSets(map[string]dynamic.ExternalActionCredentialSet{
		"project-a-1001": literalCredentialSet("project-a", "1001", "token-a"),
		"project-b-2002": literalCredentialSet("project-b", "2002", "token-b"),
	})
	require.NoError(t, err)

	binding, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
		Mode:                  "create",
		KeyTemplate:           `{{ jsonPath .Item "$.execution_id" }}`,
		AllowedCredentialSets: []string{"project-a-1001", "project-b-2002"},
		Match: dynamic.ExternalActionBindingMatch{Templates: map[string]string{
			"projectid":   `{{ scalarString (jsonPath .Item "$.change_project_id") }}`,
			"projectname": `{{ jsonPath .Item "$.change_project_name" }}`,
		}},
		Memory: &dynamic.ExternalActionBindingMemory{MaxEntries: 100},
	}, credentials)
	require.NoError(t, err)

	data := &actionData{Item: map[string]any{
		"execution_id":        "execution-42",
		"change_project_id":   "2002",
		"change_project_name": "project-b",
	}}
	matches, err := binding.matches(data)
	require.NoError(t, err)
	require.True(t, matches)
	assert.Equal(t, "project-b-2002", data.credentialSet)

	pinned, err := binding.pin(data)
	require.NoError(t, err)
	prepared, err := binding.prepare(t.Context(), pinned)
	require.NoError(t, err)
	assert.Equal(t, "project-b-2002", prepared.credentialSet)
}

func TestLifecycleBindingRejectsAmbiguousAllowedCredentialSets(t *testing.T) {
	t.Parallel()

	credentials, err := compileCredentialSets(map[string]dynamic.ExternalActionCredentialSet{
		"project-b-primary":  literalCredentialSet("project-b", "2002", "token-b1"),
		"project-b-parallel": literalCredentialSet("project-b", "2002", "token-b2"),
	})
	require.NoError(t, err)

	binding, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
		Mode:                  "create",
		KeyTemplate:           "execution-42",
		AllowedCredentialSets: []string{"project-b-primary", "project-b-parallel"},
		Match: dynamic.ExternalActionBindingMatch{Templates: map[string]string{
			"projectid":   `2002`,
			"projectname": `project-b`,
		}},
		Memory: &dynamic.ExternalActionBindingMemory{MaxEntries: 100},
	}, credentials)
	require.NoError(t, err)

	_, err = binding.matches(&actionData{})
	require.ErrorContains(t, err, "matches multiple allowed credential sets")
}

func TestLifecycleBindingResolveOrCreateReconcilesMissingBinding(t *testing.T) {
	t.Parallel()

	credentials, err := compileCredentialSets(map[string]dynamic.ExternalActionCredentialSet{
		"project-a-1001": literalCredentialSet("project-a", "1001", "token-a"),
		"project-b-2002": literalCredentialSet("project-b", "2002", "token-b"),
	})
	require.NoError(t, err)

	binding, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
		Mode:                  "resolveOrCreate",
		KeyTemplate:           `{{ jsonPath .Item "$.execution_id" }}`,
		AllowedCredentialSets: []string{"project-a-1001", "project-b-2002"},
		Match: dynamic.ExternalActionBindingMatch{Templates: map[string]string{
			"projectid": `{{ scalarString (jsonPath .Item "$.change_project_id") }}`,
		}},
		Memory: &dynamic.ExternalActionBindingMemory{MaxEntries: 100},
	}, credentials)
	require.NoError(t, err)

	newData := func() *actionData {
		return &actionData{Item: map[string]any{
			"execution_id":      "execution-42",
			"change_project_id": "2002",
		}}
	}

	first := newData()
	matches, err := binding.matches(first)
	require.NoError(t, err)
	require.True(t, matches)
	pinned, err := binding.pin(first)
	require.NoError(t, err)
	prepared, err := binding.prepare(t.Context(), pinned)
	require.NoError(t, err)
	assert.Equal(t, "project-b-2002", prepared.credentialSet)
	assert.True(t, prepared.LifecycleBinding.Created)

	second := newData()
	matches, err = binding.matches(second)
	require.NoError(t, err)
	require.True(t, matches)
	pinned, err = binding.pin(second)
	require.NoError(t, err)
	prepared, err = binding.prepare(t.Context(), pinned)
	require.NoError(t, err)
	assert.Equal(t, "project-b-2002", prepared.credentialSet)
	assert.False(t, prepared.LifecycleBinding.Created)

	mismatched := &actionData{Item: map[string]any{
		"execution_id":      "execution-42",
		"change_project_id": "1001",
	}}
	matches, err = binding.matches(mismatched)
	require.NoError(t, err)
	require.True(t, matches)
	pinned, err = binding.pin(mismatched)
	require.NoError(t, err)
	_, err = binding.prepare(t.Context(), pinned)
	require.ErrorContains(t, err, "does not match source identity credential set")
}

func TestLifecycleBindingSkipMissingReleasesDedupForLaterRetry(t *testing.T) {
	t.Parallel()

	credentials, err := compileCredentialSets(map[string]dynamic.ExternalActionCredentialSet{
		"project-a-1001": literalCredentialSet("project-a", "1001", "token-a"),
	})
	require.NoError(t, err)

	backend := &dynamic.ExternalActionBindingMemory{MaxEntries: 100}
	resolve, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
		Mode:                  "resolve",
		OnMissing:             "skip",
		KeyTemplate:           "{{ .Request.Path }}",
		AllowedCredentialSets: []string{"project-a-1001"},
		Memory:                backend,
	}, credentials)
	require.NoError(t, err)

	dedupConfig := dynamic.ExternalActionDeduplication{
		KeyTemplate: "workflow-42-end",
		TTL:         ptypes.Duration(time.Hour),
		Lease:       ptypes.Duration(time.Minute),
		Memory:      &dynamic.ExternalActionDedupMemory{MaxEntries: 100},
	}
	dedup, err := newDeduplicator(t.Context(), dedupConfig)
	require.NoError(t, err)

	calls := 0
	rule := &compiledRule{
		config:           dynamic.ExternalActionRule{Deduplication: &dedupConfig},
		binding:          resolve,
		deduplication:    dedup,
		deduplicationKey: newLiteralTemplate("workflow-42-end"),
		target: targetFunc(func(context.Context, *actionData) error {
			calls++
			return nil
		}),
	}
	data := &actionData{Request: actionRequest{Path: "workflow-42"}}

	// A historical terminal observation has no start binding. It is accepted
	// as a no-op, but its dedup lease is released instead of being completed.
	require.NoError(t, runRule(t.Context(), rule, data))
	assert.Zero(t, calls)

	create, err := newCompiledBinding(t.Context(), dynamic.ExternalActionBinding{
		Mode:          "create",
		KeyTemplate:   "{{ .Request.Path }}",
		CredentialSet: "project-a-1001",
		Memory:        backend,
	}, credentials)
	require.NoError(t, err)
	pinned, err := create.pin(data)
	require.NoError(t, err)
	_, err = create.prepare(t.Context(), pinned)
	require.NoError(t, err)

	// A later poll after the start binding exists performs the action exactly
	// once, proving the missing observation did not poison deduplication.
	require.NoError(t, runRule(t.Context(), rule, data))
	require.NoError(t, runRule(t.Context(), rule, data))
	assert.Equal(t, 1, calls)
}

func literalCredentialSet(projectName, projectID, token string) dynamic.ExternalActionCredentialSet {
	return dynamic.ExternalActionCredentialSet{
		ReloadInterval:      ptypes.Duration(time.Millisecond),
		ImmutableParameters: []string{"projectid", "projectname"},
		Auth: dynamic.ExternalActionAuth{
			Type: "headers",
			Parameters: map[string]string{
				"token":       token,
				"projectname": projectName,
				"projectid":   projectID,
			},
		},
	}
}
