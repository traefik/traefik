package queryparameters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestQueryParameters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc               string
		target             string
		config             dynamic.QueryParameters
		expectedRawQuery   string
		expectedRequestURI string
	}{
		{
			desc:               "set replaces an existing parameter in place",
			target:             "/foo?a=1&c=3&b=2",
			config:             dynamic.QueryParameters{Set: map[string]string{"c": "6"}},
			expectedRawQuery:   "a=1&c=6&b=2",
			expectedRequestURI: "/foo?a=1&c=6&b=2",
		},
		{
			desc:               "set replaces every value of a repeated parameter",
			target:             "/foo?a=1&b=2&a=10",
			config:             dynamic.QueryParameters{Set: map[string]string{"a": "5"}},
			expectedRawQuery:   "a=5&b=2",
			expectedRequestURI: "/foo?a=5&b=2",
		},
		{
			desc:               "set appends a missing parameter",
			target:             "/foo?a=1",
			config:             dynamic.QueryParameters{Set: map[string]string{"c": "6"}},
			expectedRawQuery:   "a=1&c=6",
			expectedRequestURI: "/foo?a=1&c=6",
		},
		{
			desc:               "add keeps the existing values",
			target:             "/foo?a=1",
			config:             dynamic.QueryParameters{Add: map[string]string{"a": "10"}},
			expectedRawQuery:   "a=1&a=10",
			expectedRequestURI: "/foo?a=1&a=10",
		},
		{
			desc:               "add to a request without a query",
			target:             "/foo",
			config:             dynamic.QueryParameters{Add: map[string]string{"d": "4"}},
			expectedRawQuery:   "d=4",
			expectedRequestURI: "/foo?d=4",
		},
		{
			desc:               "delete removes every value of a parameter",
			target:             "/foo?a=1&fbclid=x&b=2&fbclid=y",
			config:             dynamic.QueryParameters{Delete: []string{"fbclid"}},
			expectedRawQuery:   "a=1&b=2",
			expectedRequestURI: "/foo?a=1&b=2",
		},
		{
			desc:               "delete the only parameter",
			target:             "/foo?fbclid=x",
			config:             dynamic.QueryParameters{Delete: []string{"fbclid"}},
			expectedRawQuery:   "",
			expectedRequestURI: "/foo",
		},
		{
			desc:               "delete matches an encoded parameter name",
			target:             "/foo?f%62clid=x&a=1",
			config:             dynamic.QueryParameters{Delete: []string{"fbclid"}},
			expectedRawQuery:   "a=1",
			expectedRequestURI: "/foo?a=1",
		},
		{
			desc:               "untouched parameters keep their order and encoding",
			target:             "/foo?q=a%20b&sig=AbC%2B%2F&z=1",
			config:             dynamic.QueryParameters{Set: map[string]string{"c": "1"}},
			expectedRawQuery:   "q=a%20b&sig=AbC%2B%2F&z=1&c=1",
			expectedRequestURI: "/foo?q=a%20b&sig=AbC%2B%2F&z=1&c=1",
		},
		{
			desc:               "values are encoded",
			target:             "/foo",
			config:             dynamic.QueryParameters{Set: map[string]string{"next": "https://example.com/?a=b&c"}},
			expectedRawQuery:   "next=https%3A%2F%2Fexample.com%2F%3Fa%3Db%26c",
			expectedRequestURI: "/foo?next=https%3A%2F%2Fexample.com%2F%3Fa%3Db%26c",
		},
		{
			desc:               "appended parameters are sorted by name",
			target:             "/foo",
			config:             dynamic.QueryParameters{Set: map[string]string{"b": "2", "a": "1"}, Add: map[string]string{"d": "4", "c": "3"}},
			expectedRawQuery:   "a=1&b=2&c=3&d=4",
			expectedRequestURI: "/foo?a=1&b=2&c=3&d=4",
		},
		{
			desc:               "set and add on the same parameter",
			target:             "/foo?a=1&a=2",
			config:             dynamic.QueryParameters{Set: map[string]string{"a": "5"}, Add: map[string]string{"a": "10"}},
			expectedRawQuery:   "a=5&a=10",
			expectedRequestURI: "/foo?a=5&a=10",
		},
		{
			desc:   "delete, set and add together",
			target: "/foo?a=1&b=2&fbclid=x&c=3",
			config: dynamic.QueryParameters{
				Set:    map[string]string{"c": "6"},
				Add:    map[string]string{"d": "4"},
				Delete: []string{"fbclid"},
			},
			expectedRawQuery:   "a=1&b=2&c=6&d=4",
			expectedRequestURI: "/foo?a=1&b=2&c=6&d=4",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var actualRawQuery, actualRequestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualRawQuery = r.URL.RawQuery
				actualRequestURI = r.RequestURI
			})

			handler, err := New(t.Context(), next, test.config, "foo-query-parameters")
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://localhost"+test.target, nil)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, test.expectedRawQuery, actualRawQuery)
			assert.Equal(t, test.expectedRequestURI, actualRequestURI)
		})
	}
}

func TestNewInvalidConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc   string
		config dynamic.QueryParameters
	}{
		{
			desc:   "empty name in set",
			config: dynamic.QueryParameters{Set: map[string]string{"": "1"}},
		},
		{
			desc:   "empty name in add",
			config: dynamic.QueryParameters{Add: map[string]string{"": "1"}},
		},
		{
			desc:   "empty name in delete",
			config: dynamic.QueryParameters{Delete: []string{""}},
		},
		{
			desc:   "parameter in both set and delete",
			config: dynamic.QueryParameters{Set: map[string]string{"a": "1"}, Delete: []string{"a"}},
		},
		{
			desc:   "parameter in both add and delete",
			config: dynamic.QueryParameters{Add: map[string]string{"a": "1"}, Delete: []string{"a"}},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			_, err := New(t.Context(), next, test.config, "foo-query-parameters")
			assert.Error(t, err)
		})
	}
}
