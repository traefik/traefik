package urlrewrite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"k8s.io/utils/ptr"
)

func TestURLRewriteHandler(t *testing.T) {
	testCases := []struct {
		desc          string
		config        dynamic.URLRewrite
		method        string
		url           string
		headers       map[string]string
		expectedURL   string
		expectedHost  string
		errorExpected bool
	}{
		{
			desc: "replace path",
			config: dynamic.URLRewrite{
				Path: ptr.To("/baz"),
			},
			url:         "http://foo.com:80/foo/bar",
			expectedURL: "http://foo.com:80/baz",
		},
		{
			desc: "replace path without trailing slash",
			config: dynamic.URLRewrite{
				Path: ptr.To("/baz"),
			},
			url:         "http://foo.com:80/foo/bar/",
			expectedURL: "http://foo.com:80/baz",
		},
		{
			desc: "replace path with trailing slash",
			config: dynamic.URLRewrite{
				Path: ptr.To("/baz/"),
			},
			url:         "http://foo.com:80/foo/bar",
			expectedURL: "http://foo.com:80/baz/",
		},
		{
			desc: "only host",
			config: dynamic.URLRewrite{
				Hostname: ptr.To("bar.com"),
			},
			url:          "http://foo.com:8080/foo/",
			expectedHost: "bar.com",
		},
		{
			desc: "host and path",
			config: dynamic.URLRewrite{
				Hostname: ptr.To("bar.com"),
				Path:     ptr.To("/baz/"),
			},
			url:          "http://foo.com:8080/foo/",
			expectedURL:  "http://foo.com:8080/baz/",
			expectedHost: "bar.com",
		},
		{
			desc: "replace prefix path",
			config: dynamic.URLRewrite{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:         "http://foo.com:80/foo/bar",
			expectedURL: "http://foo.com:80/baz/bar",
		},
		{
			desc: "replace prefix path with trailing slash",
			config: dynamic.URLRewrite{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:         "http://foo.com:80/foo/bar/",
			expectedURL: "http://foo.com:80/baz/bar/",
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.URLRewrite{
				Path:       ptr.To("baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:         "http://foo.com:80/foo/bar",
			expectedURL: "http://foo.com:80/baz/bar",
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.URLRewrite{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo/"),
			},
			url:         "http://foo.com:80/foo/bar",
			expectedURL: "http://foo.com:80/baz/bar",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.expectedURL != "" {
					assert.Equal(t, test.expectedURL, r.URL.String())
				}
				if test.expectedHost != "" {
					assert.Equal(t, test.expectedHost, r.Host)
				}
			})
			handler, err := NewURLRewrite(context.Background(), next, test.config, "traefikTest")

			if test.errorExpected {
				require.Error(t, err)
				require.Nil(t, handler)
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)

				recorder := httptest.NewRecorder()

				method := http.MethodGet
				if test.method != "" {
					method = test.method
				}

				req := httptest.NewRequest(method, test.url, nil)

				for k, v := range test.headers {
					req.Header.Set(k, v)
				}

				req.Header.Set("X-Foo", "bar")
				handler.ServeHTTP(recorder, req)
			}
		})
	}
}
