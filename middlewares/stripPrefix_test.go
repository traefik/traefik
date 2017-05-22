package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		desc               string
		prefixes           []string
		path               string
		expectedStatusCode int
		expectedPath       string
	}{
		{
			desc:               "no prefixes configured",
			prefixes:           []string{},
			path:               "/noprefixes",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			desc:               "wildcard (.*) requests",
			prefixes:           []string{"/"},
			path:               "/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
		},
		{
			desc:               "prefix and path matching",
			prefixes:           []string{"/stat"},
			path:               "/stat",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
		},
		{
			desc:               "path prefix on exactly matching path",
			prefixes:           []string{"/stat/"},
			path:               "/stat/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
		},
		{
			desc:               "path prefix on matching longer path",
			prefixes:           []string{"/stat/"},
			path:               "/stat/us",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/us",
		},
		{
			desc:               "path prefix on mismatching path",
			prefixes:           []string{"/stat/"},
			path:               "/status",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			desc:               "general prefix on matching path",
			prefixes:           []string{"/stat"},
			path:               "/stat/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
		},
		{
			desc:               "earlier prefix matching",
			prefixes:           []string{"/stat", "/stat/us"},
			path:               "/stat/us",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/us",
		},
		{
			desc:               "later prefix matching",
			prefixes:           []string{"/mismatch", "/stat"},
			path:               "/stat",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			server := httptest.NewServer(&StripPrefix{
				Prefixes: test.prefixes,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotPath = r.URL.Path
				}),
			})
			defer server.Close()

			resp, err := http.Get(server.URL + test.path)
			require.NoError(t, err, "Failed to send GET request")
			assert.Equal(t, test.expectedStatusCode, resp.StatusCode, "Unexpected status code")

			assert.Equal(t, test.expectedPath, gotPath, "Unexpected path")
		})
	}
}
