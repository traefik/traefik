package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripPrefixRegex(t *testing.T) {

	testPrefixRegex := []string{"/a/api/", "/b/{regex}/", "/c/{category}/{id:[0-9]+}/"}

	tests := []struct {
		path               string
		expectedStatusCode int
		expectedPath       string
		expectedHeader     string
	}{
		{
			path:               "/a/test",
			expectedStatusCode: 404,
		},
		{
			path:               "/a/api/test",
			expectedStatusCode: 200,
			expectedPath:       "test",
			expectedHeader:     "/a/api/",
		},
		{
			path:               "/b/api/",
			expectedStatusCode: 200,
			expectedHeader:     "/b/api/",
		},
		{
			path:               "/b/api/test1",
			expectedStatusCode: 200,
			expectedPath:       "test1",
			expectedHeader:     "/b/api/",
		},
		{
			path:               "/b/api2/test2",
			expectedStatusCode: 200,
			expectedPath:       "test2",
			expectedHeader:     "/b/api2/",
		},
		{
			path:               "/c/api/123/",
			expectedStatusCode: 200,
			expectedHeader:     "/c/api/123/",
		},
		{
			path:               "/c/api/123/test3",
			expectedStatusCode: 200,
			expectedPath:       "test3",
			expectedHeader:     "/c/api/123/",
		},
		{
			path:               "/c/api/abc/test4",
			expectedStatusCode: 404,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.path, func(t *testing.T) {
			t.Parallel()

			var actualPath, actualHeader string
			handlerPath := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualHeader = r.Header.Get(ForwardedPrefixHeader)
			})

			handler := NewStripPrefixRegex(handlerPath, testPrefixRegex)
			server := httptest.NewServer(handler)
			defer server.Close()

			resp, err := http.Get(server.URL + test.path)
			require.NoError(t, err, "%s: unexpected error.", test.path)

			assert.Equal(t, test.expectedStatusCode, resp.StatusCode, "%s: unexpected status code.", test.path)
			assert.Equal(t, test.expectedPath, actualPath, "%s: unexpected path.", test.path)
			assert.Equal(t, test.expectedHeader, actualHeader, "%s: unexpected '%s' header.", test.path, ForwardedPrefixHeader)
		})
	}

}
