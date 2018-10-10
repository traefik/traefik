package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceQueryRegex(t *testing.T) {
	testCases := []struct {
		desc          string
		target        string
		regex         string
		replacement   string
		expectedQuery string
	}{
		{
			desc:          "add query parameter",
			target:        "/foo",
			regex:         `.*`,
			replacement:   "bar=baz",
			expectedQuery: "bar=baz",
		},
		{
			desc:          "remove query parameter",
			target:        "/foo?remove=yes",
			regex:         `.*(.*)`, // greedy leaves nothing
			replacement:   "$1",
			expectedQuery: "",
		},
		{
			desc:          "overwrite query parameters",
			target:        "/foo?dropped=yes",
			regex:         `.*`,
			replacement:   "bar=baz",
			expectedQuery: "bar=baz",
		},
		{
			desc:          "append query parameter",
			target:        "/foo?keep=yes",
			regex:         `^/foo\?(.*)$`,
			replacement:   "$1&bar=baz",
			expectedQuery: "keep=yes&bar=baz",
		},
		{
			desc:          "modify query parameter",
			target:        "/foo?a=a",
			regex:         `^/foo\?a=a$`,
			replacement:   "a=A",
			expectedQuery: "a=A",
		},
		{
			desc:          "use path component as new query parameters",
			target:        "/foo/animal/CAT/food/FISH?keep=no",
			regex:         `^/foo/animal/([^/]+)/food/([^?]+)(\?.*)?$`,
			replacement:   "animal=$1&food=$2",
			expectedQuery: "animal=CAT&food=FISH",
		},
		{
			desc:          "use path component as new query parameters, keep existing query params",
			target:        "/foo/animal/CAT/food/FISH?keep=yes",
			regex:         `^/foo/animal/([^/]+)/food/([^/]+)\?(.*)$`,
			replacement:   "$3&animal=$1&food=$2",
			expectedQuery: "keep=yes&animal=CAT&food=FISH",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var actualQuery string
			handler := NewReplaceQueryRegexHandler(
				test.regex,
				test.replacement,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					actualQuery = r.URL.RawQuery
				}),
			)

			req := httptest.NewRequest(http.MethodGet, test.target, nil)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, test.expectedQuery, actualQuery, "Unespected request query")
		})
	}
}
