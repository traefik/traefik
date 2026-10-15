package accessgovernance

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestCompiledConfigDatabaseFor(t *testing.T) {
	config, err := compileConfig(dynamic.AccessGovernance{
		Databases: map[string]dynamic.AccessGovernanceDatabase{
			"default":   {},
			"sensitive": {},
			"capacity":  {},
		},
		DefaultDatabase: "default",
		Routes: []dynamic.AccessGovernanceRoute{
			{
				Name:          "capacity-reduction",
				Database:      "capacity",
				Protocols:     []string{"http"},
				Methods:       []string{"POST"},
				PathRegex:     `^/tasks/create$`,
				BodyJSONPaths: map[string]string{"$.task_type": `^CAPACITY_REDUCTION$`},
			},
			{
				Name:      "sensitive",
				Database:  "sensitive",
				Protocols: []string{"http"},
				Methods:   []string{"POST"},
				PathRegex: `^/sensitive(?:/|$)`,
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "sensitive", config.databaseFor(eventRoute{
		Protocol: "http2",
		Method:   "POST",
		Path:     "/sensitive/one",
	}))
	assert.Equal(t, "default", config.databaseFor(eventRoute{
		Protocol: "http",
		Method:   "GET",
		Path:     "/sensitive/one",
	}))
	assert.Equal(t, "capacity", config.databaseFor(eventRoute{
		Protocol: "http",
		Method:   "POST",
		Path:     "/tasks/create",
		Body:     map[string]any{"task_type": "CAPACITY_REDUCTION"},
	}))
	assert.Equal(t, "default", config.databaseFor(eventRoute{
		Protocol: "http",
		Method:   "POST",
		Path:     "/tasks/create",
		Body:     map[string]any{"task_type": "OTHER"},
	}))
	assert.True(t, config.needsBodyRoute(eventRoute{
		Protocol: "http",
		Method:   "POST",
		Path:     "/tasks/create",
	}))
	assert.False(t, config.needsBodyRoute(eventRoute{
		Protocol: "http",
		Method:   "POST",
		Path:     "/unrelated",
	}))
	database, routeName := config.destinationFor(eventRoute{
		Protocol: "http",
		Method:   "POST",
		Path:     "/sensitive/one",
	})
	assert.Equal(t, "sensitive", database)
	assert.Equal(t, "sensitive", routeName)
}

func TestCompiledConfigIdentity(t *testing.T) {
	config, err := compileConfig(dynamic.AccessGovernance{
		Databases: map[string]dynamic.AccessGovernanceDatabase{"default": {}},
		Identity: []dynamic.AccessGovernanceIdentitySource{
			{Kind: "header", Name: "X-Authenticated-User"},
			{Kind: "cookie", Name: "account", Decode: "url"},
			{Kind: "json", JSONPath: "$.actor.name"},
		},
	})
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
	request.Header.Set("Cookie", "account=alice%2Eexample")
	assert.Equal(t, "alice.example", config.identity(request, nil))

	request.Header.Del("Cookie")
	body := newLazyBodyCapture(io.NopCloser(strings.NewReader(`{"actor":{"name":"bob"}}`)), 1024)
	_, err = io.ReadAll(body)
	require.NoError(t, err)
	assert.Equal(t, "bob", config.identity(request, body))
}

func TestJSONIdentityAndRoutingPreserveLargeNumbers(t *testing.T) {
	config, err := compileConfig(dynamic.AccessGovernance{
		Databases: map[string]dynamic.AccessGovernanceDatabase{"selected": {}},
		Identity:  []dynamic.AccessGovernanceIdentitySource{{Kind: "json", JSONPath: "$.actor"}},
		Routes: []dynamic.AccessGovernanceRoute{{
			Database: "selected", BodyJSONPaths: map[string]string{"$.actor": "^9007199254740993$"},
		}},
	})
	require.NoError(t, err)
	raw := `{"actor":9007199254740993}`
	body, err := newEagerBodyCapture(io.NopCloser(strings.NewReader(raw)), 1024)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
	assert.Equal(t, "9007199254740993", config.identity(request, body))
	assert.Equal(t, "selected", config.databaseFor(eventRoute{Body: parseJSONBody([]byte(raw))}))
	assert.Nil(t, parseJSONBody([]byte(raw+` {"actor":1}`)))
	assert.Nil(t, parseJSONBody([]byte(raw+` trailing`)))
}

func TestEagerBodyCaptureRestoresBody(t *testing.T) {
	body, err := newEagerBodyCapture(io.NopCloser(strings.NewReader(`{"actor":{"name":"alice"}}`)), 1024)
	require.NoError(t, err)

	assert.JSONEq(t, `{"actor":{"name":"alice"}}`, string(body.Bytes()))
	forwarded, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"actor":{"name":"alice"}}`, string(forwarded))
	assert.False(t, body.overflow)
}

func TestCompileConfigRejectsUnsafeTableName(t *testing.T) {
	_, err := compileConfig(dynamic.AccessGovernance{
		Databases: map[string]dynamic.AccessGovernanceDatabase{
			"default": {HTTPTable: "events; DROP TABLE users"},
		},
	})
	require.Error(t, err)
}
