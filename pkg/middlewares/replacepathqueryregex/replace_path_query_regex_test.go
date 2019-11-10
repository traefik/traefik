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
			desc:               "add query parameter",
			request:            "/foo",
			regex:              `.*`,
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
			regex:              `^/foo\?(.*)$`,
			replacement:        "$1&bar=baz",
			expectedQuery:      "keep=yes&bar=baz",
			expectedRequestURI: "/foo?keep=yes&bar=baz",
		},
		{
			desc:               "modify query parameter",
			request:            "/foo?a=a",
			regex:              `^/foo\?a=a$`,
			replacement:        "a=A",
			expectedQuery:      "a=A",
			expectedRequestURI: "/foo?a=A",
		},
		{
			desc:               "use path component as new query parameters",
			request:            "/foo/animal/CAT/food/FISH?keep=no",
			regex:              `^/foo/animal/([^/]+)/food/([^?]+)(\?.*)?$`,
			replacement:        "animal=$1&food=$2",
			expectedQuery:      "animal=CAT&food=FISH",
			expectedRequestURI: "/foo/animal/CAT/food/FISH?animal=CAT&food=FISH",
		},
		{
			desc:               "use path component as new query parameters, keep existing query params",
			request:            "/foo/animal/CAT/food/FISH?keep=yes",
			regex:              `^/foo/animal/([^/]+)/food/([^/]+)\?(.*)$`,
			replacement:        "$3&animal=$1&food=$2",
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

			handler, err := New(context.Background(), next, dynamic.ReplacePathQueryRegex{
				Regex:       test.regex,
				Replacement: test.replacement,
			}, "foo-replace-path-query-regexp")
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
		regex         string
		replacement   string
		expectedQuery string
	}{
		{
			desc:          "bad regex",
			request:       "/foo",
			regex:         `(?!`,
			replacement:   "",
			expectedQuery: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			config := dynamic.ReplacePathQueryRegex{
				Regex:       test.regex,
				Replacement: test.replacement,
			}

			_, err := New(context.Background(), next, config, "foo-replace-path-query-regexp")
			require.Error(t, err)
		})
	}
}
