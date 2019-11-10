package replacepathqueryregex

import (
	"context"
	"net/http"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplacePathQueryRegex(t *testing.T) {
	testCases := []struct {
		desc               string
		request            string
		config             dynamic.ReplacePathQueryRegex
		expectedQuery      string
		expectedRequestURI string
	}{
		{
			desc:    "add query parameter",
			request: "/foo",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `.*`,
				Replacement: "bar=baz",
			},
			expectedQuery:      "bar=baz",
			expectedRequestURI: "/foo?bar=baz",
		},
		{
			desc:    "add query parameter, no path",
			request: "",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `.*`,
				Replacement: "bar=baz",
			},
			expectedQuery:      "bar=baz",
			expectedRequestURI: "/?bar=baz",
		},
		{
			desc:    "remove query parameter",
			request: "/foo?remove=yes",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `.*(.*)`, // greedy leaves nothing
				Replacement: "$1",
			},
			expectedQuery:      "",
			expectedRequestURI: "/foo",
		},
		{
			desc:    "overwrite query parameters",
			request: "/foo?dropped=yes",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `.*`,
				Replacement: "bar=baz",
			},
			expectedQuery:      "bar=baz",
			expectedRequestURI: "/foo?bar=baz",
		},
		{
			desc:    "append query parameter",
			request: "/foo?keep=yes",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `^/foo\?(.*)$`,
				Replacement: "$1&bar=baz",
			},
			expectedQuery:      "keep=yes&bar=baz",
			expectedRequestURI: "/foo?keep=yes&bar=baz",
		},
		{
			desc:    "modify query parameter",
			request: "/foo?a=a",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `^/foo\?a=a$`,
				Replacement: "a=A",
			},
			expectedQuery:      "a=A",
			expectedRequestURI: "/foo?a=A",
		},
		{
			desc:    "use path component as new query parameters",
			request: "/foo/animal/CAT/food/FISH?keep=no",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `^/foo/animal/([^/]+)/food/([^?]+)(\?.*)?$`,
				Replacement: "animal=$1&food=$2",
			},
			expectedQuery:      "animal=CAT&food=FISH",
			expectedRequestURI: "/foo/animal/CAT/food/FISH?animal=CAT&food=FISH",
		},
		{
			desc:    "use path component as new query parameters, keep existing query params",
			request: "/foo/animal/CAT/food/FISH?keep=yes",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `^/foo/animal/([^/]+)/food/([^/]+)\?(.*)$`,
				Replacement: "$3&animal=$1&food=$2",
			},
			expectedQuery:      "keep=yes&animal=CAT&food=FISH",
			expectedRequestURI: "/foo/animal/CAT/food/FISH?keep=yes&animal=CAT&food=FISH",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var actualQuery, requestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestURI = r.RequestURI
				actualQuery = r.URL.RawQuery
			})

			handler, err := New(context.Background(), next, test.config, "foo-replace-path-query-regexp")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.request, nil)
			req.RequestURI = test.request

			handler.ServeHTTP(nil, req)

			assert.Equal(t, test.expectedQuery, actualQuery)
			assert.Equal(t, test.expectedRequestURI, requestURI)
		})
	}
}

func TestReplacePathQueryRegexError(t *testing.T) {
	testCases := []struct {
		desc          string
		request       string
		config        dynamic.ReplacePathQueryRegex
		expectedQuery string
	}{
		{
			desc:    "bad regex",
			request: "/foo",
			config: dynamic.ReplacePathQueryRegex{
				Regex:       `(?!`,
				Replacement: "",
			},
			expectedQuery: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			_, err := New(context.Background(), next, test.config, "foo-replace-path-query-regexp")
			require.Error(t, err)
		})
	}
}
