package stripprefix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

func TestStripPrefix(t *testing.T) {
	testCases := []struct {
		desc               string
		config             dynamic.StripPrefix
		path               string
		expectedStatusCode int
		expectedPath       string
		expectedRawPath    string
		expectedHeader     string
	}{
		{
			desc: "no prefixes configured",
			config: dynamic.StripPrefix{
				Prefixes: []string{},
			},
			path:               "/noprefixes",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/noprefixes",
		},
		{
			desc: "wildcard (.*) requests (ForceSlash)",
			config: dynamic.StripPrefix{
				Prefixes:   []string{"/"},
				ForceSlash: true,
			},
			path:               "/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
			expectedHeader:     "/",
		},
		{
			desc: "wildcard (.*) requests",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/"},
			},
			path:               "/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "",
			expectedHeader:     "/",
		},
		{
			desc: "prefix and path matching (ForceSlash)",
			config: dynamic.StripPrefix{
				Prefixes:   []string{"/stat"},
				ForceSlash: true,
			},
			path:               "/stat",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
			expectedHeader:     "/stat",
		},
		{
			desc: "prefix and path matching",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/stat"},
			},
			path:               "/stat",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "",
			expectedHeader:     "/stat",
		},
		{
			desc: "path prefix on exactly matching path (ForceSlash)",
			config: dynamic.StripPrefix{
				Prefixes:   []string{"/stat/"},
				ForceSlash: true,
			},
			path:               "/stat/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
			expectedHeader:     "/stat/",
		},
		{
			desc: "path prefix on exactly matching path",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/stat/"},
			},
			path:               "/stat/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "",
			expectedHeader:     "/stat/",
		},
		{
			desc: "path prefix on matching longer path",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/stat/"},
			},
			path:               "/stat/us",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/us",
			expectedHeader:     "/stat/",
		},
		{
			desc: "path prefix on mismatching path",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/stat/"},
			},
			path:               "/status",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/status",
		},
		{
			desc: "general prefix on matching path",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/stat"},
			},
			path:               "/stat/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
			expectedHeader:     "/stat",
		},
		{
			desc: "earlier prefix matching",
			config: dynamic.StripPrefix{

				Prefixes: []string{"/stat", "/stat/us"},
			},
			path:               "/stat/us",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/us",
			expectedHeader:     "/stat",
		},
		{
			desc: "later prefix matching (ForceSlash)",
			config: dynamic.StripPrefix{
				Prefixes:   []string{"/mismatch", "/stat"},
				ForceSlash: true,
			},
			path:               "/stat",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/",
			expectedHeader:     "/stat",
		},
		{
			desc: "later prefix matching",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/mismatch", "/stat"},
			},
			path:               "/stat",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "",
			expectedHeader:     "/stat",
		},
		{
			desc: "prefix matching within slash boundaries",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/stat"},
			},
			path:               "/status",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/us",
			expectedHeader:     "/stat",
		},
		{
			desc: "raw path is also stripped",
			config: dynamic.StripPrefix{
				Prefixes: []string{"/stat"},
			},
			path:               "/stat/a%2Fb",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/a/b",
			expectedRawPath:    "/a%2Fb",
			expectedHeader:     "/stat",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var actualPath, actualRawPath, actualHeader, requestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualRawPath = r.URL.RawPath
				actualHeader = r.Header.Get(ForwardedPrefixHeader)
				requestURI = r.RequestURI
			})

			handler, err := New(context.Background(), next, test.config, "foo-strip-prefix")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)
			req.RequestURI = req.URL.RequestURI()

			resp := &httptest.ResponseRecorder{Code: http.StatusOK}

			handler.ServeHTTP(resp, req)

			assert.Equal(t, test.expectedStatusCode, resp.Code, "Unexpected status code.")
			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedRawPath, actualRawPath, "Unexpected raw path.")
			assert.Equal(t, test.expectedHeader, actualHeader, "Unexpected '%s' header.", ForwardedPrefixHeader)

			expectedRequestURI := test.expectedPath
			if test.expectedRawPath != "" {
				// go HTTP uses the raw path when existent in the RequestURI
				expectedRequestURI = test.expectedRawPath
			}
			if test.expectedPath == "" {
				expectedRequestURI = "/"
			}
			assert.Equal(t, expectedRequestURI, requestURI, "Unexpected request URI.")
		})
	}
}
