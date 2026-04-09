package stripprefixregex

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/middlewares/stripprefix"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

func TestStripPrefixRegex(t *testing.T) {
	testCases := []struct {
		desc               string
		config             dynamic.StripPrefixRegex
		path               string
		expectedStatusCode int
		expectedPath       string
		expectedRawPath    string
		expectedRequestURI string
		expectedHeader     string
	}{
		{
			desc:               "/a/test",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/a/api/"}},
			path:               "/a/test",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/a/test",
		},
		{
			desc:               "/a/test/",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/a/api/"}},
			path:               "/a/test/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/a/test/",
		},
		{
			desc:               "/a/api/",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/a/api/"}},
			path:               "/a/api/",
			expectedStatusCode: http.StatusOK,
			// ensureLeadingSlash do not add a slash when the path is empty.
			expectedPath:       "",
			expectedRequestURI: "/",
			expectedHeader:     "/a/api/",
		},
		{
			desc:               "/a/api/test",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/a/api/"}},
			path:               "/a/api/test",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/test",
			expectedRequestURI: "/test",
			expectedHeader:     "/a/api/",
		},
		{
			desc:               "/a/api/test/",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/a/api/"}},
			path:               "/a/api/test/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/test/",
			expectedRequestURI: "/test/",
			expectedHeader:     "/a/api/",
		},
		{
			desc:               "/b/api/",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/b/([a-z0-9]+)/"}},
			path:               "/b/api/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "",
			expectedRequestURI: "/",
			expectedHeader:     "/b/api/",
		},
		{
			desc:               "/b/api",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/b/([a-z0-9]+)/"}},
			path:               "/b/api",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/b/api",
			// When the path do not match, the requestURI is not computed.
			expectedRequestURI: "",
		},
		{
			desc:               "/b/api/test1",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/b/([a-z0-9]+)/"}},
			path:               "/b/api/test1",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/test1",
			expectedRequestURI: "/test1",
			expectedHeader:     "/b/api/",
		},
		{
			desc:               "/b/api2/test2",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/b/([a-z0-9]+)/"}},
			path:               "/b/api2/test2",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/test2",
			expectedRequestURI: "/test2",
			expectedHeader:     "/b/api2/",
		},
		{
			desc:               "/c/api/123/",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/c/[a-z0-9]+/[0-9]+/"}},
			path:               "/c/api/123/",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "",
			expectedRequestURI: "/",
			expectedHeader:     "/c/api/123/",
		},
		{
			desc:               "/c/api/123",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/c/[a-z0-9]+/[0-9]+/"}},
			path:               "/c/api/123",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/c/api/123",
			// When the path do not match, the requestURI is not computed.
			expectedRequestURI: "",
		},
		{
			desc:               "/c/api/123/test3",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/c/[a-z0-9]+/[0-9]+/"}},
			path:               "/c/api/123/test3",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/test3",
			expectedRequestURI: "/test3",
			expectedHeader:     "/c/api/123/",
		},
		{
			desc:               "/c/api/abc/test4",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/c/[a-z0-9]+/[0-9]+/"}},
			path:               "/c/api/abc/test4",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/c/api/abc/test4",
			// When the path do not match, the requestURI is not computed.
			expectedRequestURI: "",
		},
		{
			desc:               "/a/api/a2Fb",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/a/api/"}},
			path:               "/a/api/a%2Fb",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/a/b",
			expectedRawPath:    "/a%2Fb",
			expectedRequestURI: "/a%2Fb",
			expectedHeader:     "/a/api/",
		},
		{
			desc:               "/b/ap69/test",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/b/([a-z0-9]+)/"}},
			path:               "/b/ap%69/test",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/test",
			expectedRawPath:    "/test",
			expectedRequestURI: "/test",
			expectedHeader:     "/b/api/",
		},
		{
			desc:               "/b/ap69/a2Fb",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/b/([a-z0-9]+)/"}},
			path:               "/b/ap%69/a%2Fb",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/a/b",
			expectedRawPath:    "/a%2Fb",
			expectedRequestURI: "/a%2Fb",
			expectedHeader:     "/b/api/",
		},
		{
			desc:               "/t2F/test/foo",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/t /test"}},
			path:               "/t%2F/test/foo",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/t//test/foo",
			expectedRawPath:    "/t%2F/test/foo",
			// When the path do not match, the requestURI is not computed.
			expectedRequestURI: "",
		},
		{
			desc:               "/t /test/a2Fb",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/t /test"}},
			path:               "/t /test/a%2Fb",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/a/b",
			expectedRawPath:    "/a%2Fb",
			expectedRequestURI: "/a%2Fb",
			expectedHeader:     "/t /test",
		},
		{
			desc:               "/t20/test/a2Fb",
			config:             dynamic.StripPrefixRegex{Regex: []string{"/t /test"}},
			path:               "/t%20/test/a%2Fb",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "/a/b",
			expectedRawPath:    "/a%2Fb",
			expectedRequestURI: "/a%2Fb",
			expectedHeader:     "/t /test",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var actualPath, actualRawPath, actualHeader, actualRequestURI string
			handlerPath := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualRawPath = r.URL.RawPath
				actualHeader = r.Header.Get(stripprefix.ForwardedPrefixHeader)
				actualRequestURI = r.RequestURI
			})
			handler, err := New(t.Context(), handlerPath, test.config, "foo-strip-prefix-regex")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)
			resp := &httptest.ResponseRecorder{Code: http.StatusOK}

			handler.ServeHTTP(resp, req)

			assert.Equal(t, test.expectedStatusCode, resp.Code, "Unexpected status code.")
			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedRawPath, actualRawPath, "Unexpected raw path.")
			assert.Equal(t, test.expectedHeader, actualHeader, "Unexpected '%s' header.", stripprefix.ForwardedPrefixHeader)
			assert.Equal(t, test.expectedRequestURI, actualRequestURI, "Unexpected request uri.")
		})
	}
}
