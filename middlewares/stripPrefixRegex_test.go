package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestStripPrefixRegex(t *testing.T) {
	testPrefixRegex := []string{"/a/api/", "/b/{regex}/", "/c/{category}/{id:[0-9]+}/"}

	tests := []struct {
		path               string
		expectedStatusCode int
		expectedPath       string
		expectedRawPath    string
		expectedHeader     string
	}{
		{
			path:               "/a/test",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			path:               "/a/api/test",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "test",
			expectedHeader:     "/a/api/",
		},
		{
			path:               "/b/api/",
			expectedStatusCode: http.StatusOK,
			expectedHeader:     "/b/api/",
		},
		{
			path:               "/b/api/test1",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "test1",
			expectedHeader:     "/b/api/",
		},
		{
			path:               "/b/api2/test2",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "test2",
			expectedHeader:     "/b/api2/",
		},
		{
			path:               "/c/api/123/",
			expectedStatusCode: http.StatusOK,
			expectedHeader:     "/c/api/123/",
		},
		{
			path:               "/c/api/123/test3",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "test3",
			expectedHeader:     "/c/api/123/",
		},
		{
			path:               "/c/api/abc/test4",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			path:               "/a/api/a%2Fb",
			expectedStatusCode: http.StatusOK,
			expectedPath:       "a/b",
			expectedRawPath:    "a%2Fb",
			expectedHeader:     "/a/api/",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.path, func(t *testing.T) {
			t.Parallel()

			var actualPath, actualRawPath, actualHeader string
			handlerPath := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualRawPath = r.URL.RawPath
				actualHeader = r.Header.Get(ForwardedPrefixHeader)
			})
			handler := NewStripPrefixRegex(handlerPath, testPrefixRegex)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)
			resp := &httptest.ResponseRecorder{Code: http.StatusOK}

			handler.ServeHTTP(resp, req)

			assert.Equal(t, test.expectedStatusCode, resp.Code, "Unexpected status code.")
			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedRawPath, actualRawPath, "Unexpected raw path.")
			assert.Equal(t, test.expectedHeader, actualHeader, "Unexpected '%s' header.", ForwardedPrefixHeader)
		})
	}
}
