package middlewares

import (
	"net/http"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestAddPrefix(t *testing.T) {
	tests := []struct {
		desc            string
		prefix          string
		path            string
		expectedPath    string
		expectedRawPath string
	}{
		{
			desc:         "regular path",
			prefix:       "/a",
			path:         "/b",
			expectedPath: "/a/b",
		},
		{
			desc:            "raw path is supported",
			prefix:          "/a",
			path:            "/b%2Fc",
			expectedPath:    "/a/b/c",
			expectedRawPath: "/a/b%2Fc",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var actualPath, actualRawPath, requestURI string
			handler := &AddPrefix{
				Prefix: test.prefix,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					actualPath = r.URL.Path
					actualRawPath = r.URL.RawPath
					requestURI = r.RequestURI
				}),
			}

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedRawPath, actualRawPath, "Unexpected raw path.")

			expectedURI := test.expectedPath
			if test.expectedRawPath != "" {
				// go HTTP uses the raw path when existent in the RequestURI
				expectedURI = test.expectedRawPath
			}
			assert.Equal(t, expectedURI, requestURI, "Unexpected request URI.")
		})
	}
}
