package replacequeryregex

import (
	"context"
	"net/http"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceQueryRegex(t *testing.T) {
	testCases := []struct {
		desc               string
		request            string
		config             dynamic.ReplaceQueryRegex
		expectedQuery      string
		expectedRequestURI string
	}{
		{
			desc:    "no query no match",
			request: "/foo",
			config: dynamic.ReplaceQueryRegex{
				Regex:       `(.+)`,
				Replacement: "bar=baz",
			},
			expectedQuery:      "",
			expectedRequestURI: "/foo",
		},
		{
			desc:    "no query but match",
			request: "/foo",
			config: dynamic.ReplaceQueryRegex{
				Regex:       `(.*)`,
				Replacement: "bar=baz",
			},
			expectedQuery:      "bar=baz",
			expectedRequestURI: "/foo?bar=baz",
		},
		{
			desc:    "no query but match, no path",
			request: "",
			config: dynamic.ReplaceQueryRegex{
				Regex:       `(.*)`,
				Replacement: "bar=baz",
			},
			expectedQuery:      "bar=baz",
			expectedRequestURI: "/?bar=baz",
		},
		{
			desc:    "remove query parameter",
			request: "/foo?remove=yes",
			config: dynamic.ReplaceQueryRegex{
				Regex:       `.*(.*)`, // greedy leaves nothing
				Replacement: "$1",
			},
			expectedQuery:      "",
			expectedRequestURI: "/foo",
		},
		{
			desc:    "overwrite query parameters",
			request: "/foo?dropped=yes",
			config: dynamic.ReplaceQueryRegex{
				Regex:       `.*`,
				Replacement: "bar=baz",
			},
			expectedQuery:      "bar=baz",
			expectedRequestURI: "/foo?bar=baz",
		},
		{
			desc:    "append query parameter",
			request: "/foo?keep=yes",
			config: dynamic.ReplaceQueryRegex{
				Regex:       `(.*)`,
				Replacement: "$1&bar=baz",
			},
			expectedQuery:      "keep=yes&bar=baz",
			expectedRequestURI: "/foo?keep=yes&bar=baz",
		},
		{
			desc:    "modify query parameter",
			request: "/foo?@=a",
			config: dynamic.ReplaceQueryRegex{
				Regex:       `@=a`,
				Replacement: "a=A",
			},
			expectedQuery:      "a=A",
			expectedRequestURI: "/foo?a=A",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var actualQuery, requestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestURI = r.RequestURI
				actualQuery = r.URL.RawQuery
			})

			handler, err := New(context.Background(), next, test.config, "foo-replace-query-regexp")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.request, nil)
			req.RequestURI = test.request

			handler.ServeHTTP(nil, req)

			assert.Equal(t, test.expectedQuery, actualQuery)
			assert.Equal(t, test.expectedRequestURI, requestURI)
		})
	}
}

func TestReplaceQueryRegexError(t *testing.T) {
	testCases := []struct {
		desc          string
		target        string
		config        dynamic.ReplaceQueryRegex
		expectedQuery string
	}{
		{
			desc:   "bad regex",
			target: "/foo",
			config: dynamic.ReplaceQueryRegex{
				Regex:       `(?!`,
				Replacement: "",
			},
			expectedQuery: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			_, err := New(context.Background(), next, test.config, "foo-replace-query-regexp")
			require.Error(t, err)
		})
	}
}
