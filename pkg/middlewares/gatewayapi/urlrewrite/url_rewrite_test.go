package urlrewrite

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"k8s.io/utils/ptr"
)

func TestURLRewriteHandler(t *testing.T) {
	testCases := []struct {
		desc     string
		config   dynamic.URLRewrite
		url      string
		wantURL  string
		wantHost string
	}{
		{
			desc: "replace path",
			config: dynamic.URLRewrite{
				Path: ptr.To("/baz"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz",
			wantHost: "foo.com",
		},
		{
			desc: "replace path without trailing slash",
			config: dynamic.URLRewrite{
				Path: ptr.To("/baz"),
			},
			url:      "http://foo.com/foo/bar/",
			wantURL:  "http://foo.com/baz",
			wantHost: "foo.com",
		},
		{
			desc: "replace path with trailing slash",
			config: dynamic.URLRewrite{
				Path: ptr.To("/baz/"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz/",
			wantHost: "foo.com",
		},
		{
			desc: "only host",
			config: dynamic.URLRewrite{
				Hostname: ptr.To("bar.com"),
			},
			url:      "http://foo.com/foo/",
			wantURL:  "http://foo.com/foo/",
			wantHost: "bar.com",
		},
		{
			desc: "host and path",
			config: dynamic.URLRewrite{
				Hostname: ptr.To("bar.com"),
				Path:     ptr.To("/baz/"),
			},
			url:      "http://foo.com/foo/",
			wantURL:  "http://foo.com/baz/",
			wantHost: "bar.com",
		},
		{
			desc: "replace prefix path",
			config: dynamic.URLRewrite{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz/bar",
			wantHost: "foo.com",
		},
		{
			desc: "replace prefix path with trailing slash",
			config: dynamic.URLRewrite{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:      "http://foo.com/foo/bar/",
			wantURL:  "http://foo.com/baz/bar/",
			wantHost: "foo.com",
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.URLRewrite{
				Path:       ptr.To("baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz/bar",
			wantHost: "foo.com",
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.URLRewrite{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo/"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz/bar",
			wantHost: "foo.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			handler := NewURLRewrite(t.Context(), next, test.config, "traefikTest")

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, test.wantURL, req.URL.String())
			assert.Equal(t, test.wantHost, req.Host)
		})
	}
}
