package addprefix

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

func TestNewAddPrefix(t *testing.T) {
	testCases := []struct {
		desc         string
		prefix       dynamic.AddPrefix
		expectsError bool
	}{
		{
			desc:   "Works with a non empty prefix",
			prefix: dynamic.AddPrefix{Prefix: "/a"},
		},
		{
			desc:         "Fails if prefix is empty",
			prefix:       dynamic.AddPrefix{Prefix: ""},
			expectsError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			_, err := New(context.Background(), next, test.prefix, "foo-add-prefix")
			if test.expectsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddPrefix(t *testing.T) {
	testCases := []struct {
		desc            string
		prefix          dynamic.AddPrefix
		path            string
		expectedPath    string
		expectedRawPath string
	}{
		{
			desc:         "Works with a regular path",
			prefix:       dynamic.AddPrefix{Prefix: "/a"},
			path:         "/b",
			expectedPath: "/a/b",
		},
		{
			desc:         "Works with missing leading slash",
			prefix:       dynamic.AddPrefix{Prefix: "a"},
			path:         "/",
			expectedPath: "/a/",
		},
		{
			desc:            "Works with a raw path",
			prefix:          dynamic.AddPrefix{Prefix: "/a"},
			path:            "/b%2Fc",
			expectedPath:    "/a/b/c",
			expectedRawPath: "/a/b%2Fc",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var actualPath, actualRawPath, requestURI string

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualRawPath = r.URL.RawPath
				requestURI = r.RequestURI
			})

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)

			handler, err := New(context.Background(), next, test.prefix, "foo-add-prefix")
			require.NoError(t, err)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, test.expectedPath, actualPath)
			assert.Equal(t, test.expectedRawPath, actualRawPath)

			expectedURI := test.expectedPath
			if test.expectedRawPath != "" {
				// go HTTP uses the raw path when existent in the RequestURI
				expectedURI = test.expectedRawPath
			}
			assert.Equal(t, expectedURI, requestURI)
		})
	}
}
