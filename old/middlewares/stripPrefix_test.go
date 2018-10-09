package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		desc               string
		prefixes           []string
		path               string
		expectedStatusCode int
		expectedPath       string
		expectedRawPath    string
		expectedHeader     string
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
			expectedHeader:     "/",
		},
		{
			desc:               "prefix and path matching",
			prefixes:           []string{"/stat"},
			path:               "/stat",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
			expectedHeader:     "/stat",
		},
		{
			desc:               "path prefix on exactly matching path",
			prefixes:           []string{"/stat/"},
			path:               "/stat/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
			expectedHeader:     "/stat/",
		},
		{
			desc:               "path prefix on matching longer path",
			prefixes:           []string{"/stat/"},
			path:               "/stat/us",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/us",
			expectedHeader:     "/stat/",
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
			expectedHeader:     "/stat",
		},
		{
			desc:               "earlier prefix matching",
			prefixes:           []string{"/stat", "/stat/us"},
			path:               "/stat/us",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/us",
			expectedHeader:     "/stat",
		},
		{
			desc:               "later prefix matching",
			prefixes:           []string{"/mismatch", "/stat"},
			path:               "/stat",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
			expectedHeader:     "/stat",
		},
		{
			desc:               "prefix matching within slash boundaries",
			prefixes:           []string{"/stat"},
			path:               "/status",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/us",
			expectedHeader:     "/stat",
		},
		{
			desc:               "raw path is also stripped",
			prefixes:           []string{"/stat"},
			path:               "/stat/a%2Fb",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/a/b",
			expectedRawPath:    "/a%2Fb",
			expectedHeader:     "/stat",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var actualPath, actualRawPath, actualHeader, requestURI string
			handler := &StripPrefix{
				Prefixes: test.prefixes,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					actualPath = r.URL.Path
					actualRawPath = r.URL.RawPath
					actualHeader = r.Header.Get(ForwardedPrefixHeader)
					requestURI = r.RequestURI
				}),
			}

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)
			resp := &httptest.ResponseRecorder{Code: http.StatusOK}

			handler.ServeHTTP(resp, req)

			assert.Equal(t, test.expectedStatusCode, resp.Code, "Unexpected status code.")
			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedRawPath, actualRawPath, "Unexpected raw path.")
			assert.Equal(t, test.expectedHeader, actualHeader, "Unexpected '%s' header.", ForwardedPrefixHeader)

			expectedURI := test.expectedPath
			if test.expectedRawPath != "" {
				// go HTTP uses the raw path when existent in the RequestURI
				expectedURI = test.expectedRawPath
			}
			assert.Equal(t, expectedURI, requestURI, "Unexpected request URI.")
		})
	}
}
