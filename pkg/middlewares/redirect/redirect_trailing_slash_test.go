package redirect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestRedirectTrailingSlash(t *testing.T) {
	testCases := []struct {
		desc           string
		mode           dynamic.TrailingSlashMode
		permanent      bool
		method         string
		requestURL     string
		expectedURL    string
		expectedStatus int
	}{
		// Encoded characters
		{
			desc:           "ADD: encoded path /caf%C3%A9 -> /caf%C3%A9/",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			requestURL:     "http://example.com/caf%C3%A9",
			expectedURL:    "http://example.com/caf%C3%A9/",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "REMOVE: encoded path /caf%C3%A9/ -> /caf%C3%A9",
			mode:           dynamic.TrailingSlashRemove,
			permanent:      true,
			requestURL:     "http://example.com/caf%C3%A9/",
			expectedURL:    "http://example.com/caf%C3%A9",
			expectedStatus: http.StatusMovedPermanently,
		},

		// Port in URL
		{
			desc:           "ADD: URL with port /about -> /about/",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			requestURL:     "http://example.com:8080/about",
			expectedURL:    "http://example.com:8080/about/",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "REMOVE: URL with port /about/ -> /about",
			mode:           dynamic.TrailingSlashRemove,
			permanent:      true,
			requestURL:     "http://example.com:8080/about/",
			expectedURL:    "http://example.com:8080/about",
			expectedStatus: http.StatusMovedPermanently,
		},

		// No path (root)
		{
			desc:           "ADD: root / -> no redirect",
			mode:           dynamic.TrailingSlashAdd,
			requestURL:     "http://example.com/",
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "REMOVE: root / -> no redirect",
			mode:           dynamic.TrailingSlashRemove,
			requestURL:     "http://example.com/",
			expectedStatus: http.StatusOK,
		},

		// ADD mode - should redirect
		{
			desc:           "ADD: /about -> /about/",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			requestURL:     "http://example.com/about",
			expectedURL:    "http://example.com/about/",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "ADD: /docs/index -> /docs/index/",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			requestURL:     "http://example.com/docs/index",
			expectedURL:    "http://example.com/docs/index/",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "ADD: /search?q=traefik -> /search/?q=traefik",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			requestURL:     "http://example.com/search?q=traefik",
			expectedURL:    "http://example.com/search/?q=traefik",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "ADD: temporary redirect",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      false,
			requestURL:     "http://example.com/about",
			expectedURL:    "http://example.com/about/",
			expectedStatus: http.StatusFound,
		},
		{
			desc:           "ADD: multiple query params",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			requestURL:     "http://example.com/search?q=traefik&lang=en",
			expectedURL:    "http://example.com/search/?q=traefik&lang=en",
			expectedStatus: http.StatusMovedPermanently,
		},

		// ADD mode - should NOT redirect
		{
			desc:           "ADD: /about/ -> no redirect",
			mode:           dynamic.TrailingSlashAdd,
			requestURL:     "http://example.com/about/",
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "ADD: /?mode=dev -> no redirect",
			mode:           dynamic.TrailingSlashAdd,
			requestURL:     "http://example.com/?mode=dev",
			expectedStatus: http.StatusOK,
		},

		// REMOVE mode - should redirect
		{
			desc:           "REMOVE: /contact/ -> /contact",
			mode:           dynamic.TrailingSlashRemove,
			permanent:      true,
			requestURL:     "http://example.com/contact/",
			expectedURL:    "http://example.com/contact",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "REMOVE: /docs/index/ -> /docs/index",
			mode:           dynamic.TrailingSlashRemove,
			permanent:      true,
			requestURL:     "http://example.com/docs/index/",
			expectedURL:    "http://example.com/docs/index",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "REMOVE: /search/?q=traefik -> /search?q=traefik",
			mode:           dynamic.TrailingSlashRemove,
			permanent:      true,
			requestURL:     "http://example.com/search/?q=traefik",
			expectedURL:    "http://example.com/search?q=traefik",
			expectedStatus: http.StatusMovedPermanently,
		},

		// REMOVE mode - should NOT redirect
		{
			desc:           "REMOVE: /contact -> no redirect",
			mode:           dynamic.TrailingSlashRemove,
			requestURL:     "http://example.com/contact",
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "REMOVE: /?mode=dev -> no redirect",
			mode:           dynamic.TrailingSlashRemove,
			requestURL:     "http://example.com/?mode=dev",
			expectedStatus: http.StatusOK,
		},

		// File extension paths
		{
			desc:           "ADD: path with extension /file.html -> /file.html/",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			requestURL:     "http://example.com/file.html",
			expectedURL:    "http://example.com/file.html/",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "REMOVE: path with extension /file.html/ -> /file.html",
			mode:           dynamic.TrailingSlashRemove,
			permanent:      true,
			requestURL:     "http://example.com/file.html/",
			expectedURL:    "http://example.com/file.html",
			expectedStatus: http.StatusMovedPermanently,
		},

		// HTTPS scheme
		{
			desc:           "ADD: HTTPS scheme",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			requestURL:     "https://example.com/about",
			expectedURL:    "https://example.com/about/",
			expectedStatus: http.StatusMovedPermanently,
		},

		// Non-GET methods: permanent=false -> 307, permanent=true -> 308
		{
			desc:           "ADD: POST temporary redirect -> 307",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      false,
			method:         http.MethodPost,
			requestURL:     "http://example.com/about",
			expectedURL:    "http://example.com/about/",
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			desc:           "ADD: POST permanent redirect -> 308",
			mode:           dynamic.TrailingSlashAdd,
			permanent:      true,
			method:         http.MethodPost,
			requestURL:     "http://example.com/about",
			expectedURL:    "http://example.com/about/",
			expectedStatus: http.StatusPermanentRedirect,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			})

			conf := dynamic.RedirectTrailingSlash{
				Mode:      tc.mode,
				Permanent: tc.permanent,
			}

			handler, err := NewRedirectTrailingSlash(context.Background(), next, conf, "test")
			require.NoError(t, err)

			method := http.MethodGet
			if tc.method != "" {
				method = tc.method
			}

			req := httptest.NewRequest(method, tc.requestURL, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedURL != "" {
				location := recorder.Header().Get("Location")
				assert.Equal(t, tc.expectedURL, location)
			}
		})
	}
}

func TestRedirectTrailingSlashInvalidMode(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	conf := dynamic.RedirectTrailingSlash{
		Mode: "invalid",
	}

	_, err := NewRedirectTrailingSlash(context.Background(), next, conf, "test")
	require.Error(t, err)
}
