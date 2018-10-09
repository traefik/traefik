package addprefix

import (
	"context"
	"net/http"
	"testing"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/testhelpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAddPrefix(t *testing.T) {
	testCases := []struct {
		desc         string
		prefix       config.AddPrefix
		expectsError bool
	}{
		{
			desc:   "Works with a non empty prefix",
			prefix: config.AddPrefix{Prefix: "/a"},
		},
		{
			desc:         "Fails if prefix is empty",
			prefix:       config.AddPrefix{Prefix: ""},
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
	logrus.SetLevel(logrus.DebugLevel)
	testCases := []struct {
		desc            string
		prefix          config.AddPrefix
		path            string
		expectedPath    string
		expectedRawPath string
	}{
		{
			desc:         "Works with a regular path",
			prefix:       config.AddPrefix{Prefix: "/a"},
			path:         "/b",
			expectedPath: "/a/b",
		},
		{
			desc:            "Works with a raw path",
			prefix:          config.AddPrefix{Prefix: "/a"},
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
