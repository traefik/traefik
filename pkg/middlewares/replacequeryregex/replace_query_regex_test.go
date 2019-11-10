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
		regex              string
		replacement        string
		expectedQuery      string
		expectedRequestURI string
	}{
		{
			desc:               "no query no match",
			request:            "/foo",
			regex:              `(.+)`,
			replacement:        "bar=baz",
			expectedQuery:      "",
			expectedRequestURI: "/foo",
		},
		{
			desc:               "no query but match",
			request:            "/foo",
			regex:              `(.*)`,
			replacement:        "bar=baz",
			expectedQuery:      "bar=baz",
			expectedRequestURI: "/foo?bar=baz",
		},
		{
			desc:               "remove query parameter",
			request:            "/foo?remove=yes",
			regex:              `.*(.*)`, // greedy leaves nothing
			replacement:        "$1",
			expectedQuery:      "",
			expectedRequestURI: "/foo",
		},
		{
			desc:               "overwrite query parameters",
			request:            "/foo?dropped=yes",
			regex:              `.*`,
			replacement:        "bar=baz",
			expectedQuery:      "bar=baz",
			expectedRequestURI: "/foo?bar=baz",
		},
		{
			desc:               "append query parameter",
			request:            "/foo?keep=yes",
			regex:              `(.*)`,
			replacement:        "$1&bar=baz",
			expectedQuery:      "keep=yes&bar=baz",
			expectedRequestURI: "/foo?keep=yes&bar=baz",
		},
		{
			desc:               "modify query parameter",
			request:            "/foo?@=a",
			regex:              `@=a`,
			replacement:        "a=A",
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

			config := dynamic.ReplaceQueryRegex{
				Regex:       test.regex,
				Replacement: test.replacement,
			}

			handler, err := New(context.Background(), next, config, "foo-replace-query-regexp")
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
		regex         string
		replacement   string
		expectedQuery string
	}{
		{
			desc:          "bad regex",
			target:        "/foo",
			regex:         `(?!`,
			replacement:   "",
			expectedQuery: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			config := dynamic.ReplaceQueryRegex{
				Regex:       test.regex,
				Replacement: test.replacement,
			}

			_, err := New(context.Background(), next, config, "foo-replace-query-regexp")
			require.Error(t, err)
		})
	}
}
