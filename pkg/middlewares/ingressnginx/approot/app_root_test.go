package approot

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestAppRoot(t *testing.T) {
	testCases := []struct {
		desc             string
		config           dynamic.AppRoot
		requestURL       string
		expectedStatus   int
		expectedLocation string
		expectsError     bool
	}{
		{
			desc:         "empty path",
			config:       dynamic.AppRoot{Path: ""},
			expectsError: true,
		},
		{
			desc:         "path does not start with /",
			config:       dynamic.AppRoot{Path: "http://foo.com"},
			expectsError: true,
		},
		{
			desc:             "root path redirects to app root",
			config:           dynamic.AppRoot{Path: "/login"},
			requestURL:       "http://example.com/",
			expectedStatus:   http.StatusFound,
			expectedLocation: "http://example.com/login",
		},
		{
			desc:             "root path with query params and $is_args$args preserves them",
			config:           dynamic.AppRoot{Path: "/login$is_args$args"},
			requestURL:       "http://example.com/?foo=bar",
			expectedStatus:   http.StatusFound,
			expectedLocation: "http://example.com/login?foo=bar",
		},
		{
			desc:             "root path with query params without $is_args$args drops them",
			config:           dynamic.AppRoot{Path: "/login"},
			requestURL:       "http://example.com/?foo=bar",
			expectedStatus:   http.StatusFound,
			expectedLocation: "http://example.com/login",
		},
		{
			desc:           "non-root path passes through",
			config:         dynamic.AppRoot{Path: "/login"},
			requestURL:     "http://example.com/something",
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {})

			mw, err := New(t.Context(), next, test.config, "app-root")
			if test.expectsError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, test.requestURL, nil)
			rw := httptest.NewRecorder()

			mw.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatus, rw.Code)
			if test.expectedLocation != "" {
				assert.Equal(t, test.expectedLocation, rw.Header().Get("Location"))
			}
		})
	}
}
