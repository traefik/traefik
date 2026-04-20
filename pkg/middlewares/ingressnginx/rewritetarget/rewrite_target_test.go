package rewritetarget

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestRewriteTarget(t *testing.T) {
	testCases := []struct {
		desc                     string
		path                     string
		config                   dynamic.RewriteTarget
		expectedPath             string
		expectedRawPath          string
		expectedXForwardedPrefix string
		expectedStatusCode       int
		expectedRedirectURL      string
		expectsError             bool
	}{
		{
			desc: "empty replacement",
			config: dynamic.RewriteTarget{
				Replacement: "",
			},
			expectsError: true,
		},
		{
			desc: "plain replacement",
			path: "/foo/bar",
			config: dynamic.RewriteTarget{
				Replacement: "/replacement",
			},
			expectedPath:    "/replacement",
			expectedRawPath: "/replacement",
		},
		{
			desc: "plain replacement with escaped char in replacement",
			path: "/foo",
			config: dynamic.RewriteTarget{
				Replacement: "/foo%2Fbar",
			},
			expectedPath:    "/foo/bar",
			expectedRawPath: "/foo%2Fbar",
		},
		{
			desc: "plain replacement with x-forwarded-prefix",
			path: "/foo/bar",
			config: dynamic.RewriteTarget{
				Replacement:      "/replacement",
				XForwardedPrefix: "/foo",
			},
			expectedPath:             "/replacement",
			expectedRawPath:          "/replacement",
			expectedXForwardedPrefix: "/foo",
		},
		{
			desc: "regex with capture group",
			path: "/foo/bar",
			config: dynamic.RewriteTarget{
				Regex:       `^/foo/(.*)`,
				Replacement: "/new/$1",
			},
			expectedPath:    "/new/bar",
			expectedRawPath: "/new/bar",
		},
		{
			desc: "regex with multiple capture groups",
			path: "/downloads/src/source.go",
			config: dynamic.RewriteTarget{
				Regex:       `^(?i)/downloads/([^/]+)/([^/]+)$`,
				Replacement: "/downloads/$1-$2",
			},
			expectedPath:    "/downloads/src-source.go",
			expectedRawPath: "/downloads/src-source.go",
		},
		{
			desc: "regex with escaped char in replacement",
			path: "/aaa/bbb",
			config: dynamic.RewriteTarget{
				Regex:       `/aaa/bbb`,
				Replacement: "/foo%2Fbar",
			},
			expectedPath:    "/foo/bar",
			expectedRawPath: "/foo%2Fbar",
		},
		{
			desc: "regex - no match passthrough",
			path: "/foo/bar",
			config: dynamic.RewriteTarget{
				Regex:       `^/baz/(.*)`,
				Replacement: "/new/$1",
			},
			expectedPath:    "/foo/bar",
			expectedRawPath: "",
		},
		{
			desc: "invalid regex",
			config: dynamic.RewriteTarget{
				Regex:       `^(?err)/invalid/regexp/([^/]+)$`,
				Replacement: "/valid/$1",
			},
			expectsError: true,
		},
		{
			desc: "regex with x-forwarded-prefix capture group",
			path: "/foo/bar",
			config: dynamic.RewriteTarget{
				Regex:            `^(/foo)/(.*)`,
				Replacement:      "/$2",
				XForwardedPrefix: "$1",
			},
			expectedPath:             "/bar",
			expectedRawPath:          "/bar",
			expectedXForwardedPrefix: "/foo",
		},
		{
			desc: "regex with x-forwarded-prefix using third capture group",
			path: "/prefix/sub/endpoint",
			config: dynamic.RewriteTarget{
				Regex:            `^/(prefix)/(sub)/(.*)`,
				Replacement:      "/$3",
				XForwardedPrefix: "/$1/$2",
			},
			expectedPath:             "/endpoint",
			expectedRawPath:          "/endpoint",
			expectedXForwardedPrefix: "/prefix/sub",
		},
		{
			desc: "x-forwarded-prefix not set when regex does not match",
			path: "/foo/bar",
			config: dynamic.RewriteTarget{
				Regex:            `^/baz/(.*)`,
				Replacement:      "/$1",
				XForwardedPrefix: "/baz",
			},
			expectedPath:             "/foo/bar",
			expectedRawPath:          "",
			expectedXForwardedPrefix: "",
		},
		{
			desc: "full URL replacement issues 302 redirect",
			path: "/some/path",
			config: dynamic.RewriteTarget{
				Replacement: "https://bar.example.org/some/path",
			},
			expectedStatusCode:  http.StatusFound,
			expectedRedirectURL: "https://bar.example.org/some/path",
		},
		{
			desc: "regex with full URL replacement issues 302 redirect",
			path: "/prefix/foo",
			config: dynamic.RewriteTarget{
				Regex:       `(?i)/prefix(/|$)(.*)`,
				Replacement: "https://bar.example.org/$2",
			},
			expectedStatusCode:  http.StatusFound,
			expectedRedirectURL: "https://bar.example.org/foo",
		},
		{
			desc: "regex with full URL replacement - multiple paths, no regex",
			path: "/foo/a/b/c",
			config: dynamic.RewriteTarget{
				Regex:       "",
				Replacement: "https://bar.example.org/$1",
			},
			expectedStatusCode:  http.StatusFound,
			expectedRedirectURL: "https://bar.example.org/",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var actualPath, actualRawPath, actualXForwardedPrefix, requestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualRawPath = r.URL.RawPath
				actualXForwardedPrefix = r.Header.Get(xForwardedPrefixHeader)
				requestURI = r.RequestURI
			})

			handler, err := New(t.Context(), next, test.config, "test-rewrite-target")
			if test.expectsError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			server := httptest.NewServer(handler)
			defer server.Close()

			client := &http.Client{
				CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			resp, err := client.Get(server.URL + test.path)
			require.NoError(t, err)

			expectedStatus := test.expectedStatusCode
			if expectedStatus == 0 {
				expectedStatus = http.StatusOK
			}
			require.Equal(t, expectedStatus, resp.StatusCode)

			if test.expectedRedirectURL != "" {
				assert.Equal(t, test.expectedRedirectURL, resp.Header.Get("Location"), "Unexpected redirect location.")
				return
			}

			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedRawPath, actualRawPath, "Unexpected raw path.")
			assert.Equal(t, test.expectedXForwardedPrefix, actualXForwardedPrefix, "Unexpected %s header.", xForwardedPrefixHeader)

			if actualRawPath == "" {
				assert.Equal(t, actualPath, requestURI, "Unexpected request URI.")
			} else {
				assert.Equal(t, actualRawPath, requestURI, "Unexpected request URI.")
			}
		})
	}
}
