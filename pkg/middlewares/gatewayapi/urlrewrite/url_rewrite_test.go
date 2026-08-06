package urlrewrite

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestURLRewriteHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		config         dynamic.URLRewrite
		url            string
		wantURL        string
		wantHost       string
		wantStatusCode int
	}{
		{
			desc: "replace path",
			config: dynamic.URLRewrite{
				Path: new("/baz"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz",
			wantHost: "foo.com",
		},
		{
			desc: "replace path without trailing slash",
			config: dynamic.URLRewrite{
				Path: new("/baz"),
			},
			url:      "http://foo.com/foo/bar/",
			wantURL:  "http://foo.com/baz",
			wantHost: "foo.com",
		},
		{
			desc: "replace path with trailing slash",
			config: dynamic.URLRewrite{
				Path: new("/baz/"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz/",
			wantHost: "foo.com",
		},
		{
			desc: "only host",
			config: dynamic.URLRewrite{
				Hostname: new("bar.com"),
			},
			url:      "http://foo.com/foo/",
			wantURL:  "http://foo.com/foo/",
			wantHost: "bar.com",
		},
		{
			desc: "host and path",
			config: dynamic.URLRewrite{
				Hostname: new("bar.com"),
				Path:     new("/baz/"),
			},
			url:      "http://foo.com/foo/",
			wantURL:  "http://foo.com/baz/",
			wantHost: "bar.com",
		},
		{
			desc: "replace prefix path",
			config: dynamic.URLRewrite{
				Path:       new("/baz"),
				PathPrefix: new("/foo"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz/bar",
			wantHost: "foo.com",
		},
		{
			desc: "replace prefix path with trailing slash",
			config: dynamic.URLRewrite{
				Path:       new("/baz"),
				PathPrefix: new("/foo"),
			},
			url:      "http://foo.com/foo/bar/",
			wantURL:  "http://foo.com/baz/bar/",
			wantHost: "foo.com",
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.URLRewrite{
				Path:       new("baz"),
				PathPrefix: new("/foo"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz/bar",
			wantHost: "foo.com",
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.URLRewrite{
				Path:       new("/baz"),
				PathPrefix: new("/foo/"),
			},
			url:      "http://foo.com/foo/bar",
			wantURL:  "http://foo.com/baz/bar",
			wantHost: "foo.com",
		},
		{
			desc: "replace prefix path with trailing slash prefix",
			config: dynamic.URLRewrite{
				Path:       new("/baz"),
				PathPrefix: new("/foo/"),
			},
			url:      "http://foo.com/foo",
			wantURL:  "http://foo.com/baz",
			wantHost: "foo.com",
		},
		{
			desc: "replace prefix path with empty replacement with slash",
			config: dynamic.URLRewrite{
				Path:       new(""),
				PathPrefix: new("/foo"),
			},
			url:      "http://foo.com/foo",
			wantURL:  "http://foo.com/",
			wantHost: "foo.com",
		},
		{
			desc: "dot-segment traversal fused with an encoded slash is rejected",
			config: dynamic.URLRewrite{
				Path:       new("/baz"),
				PathPrefix: new("/foo"),
			},
			url:            "http://foo.com/foo..%2Fadmin",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			desc: "encoded slash in the trimmed tail succeeds",
			config: dynamic.URLRewrite{
				Path:       new("/baz"),
				PathPrefix: new("/foo"),
			},
			url:      "http://foo.com/foo/a%2Fb",
			wantURL:  "http://foo.com/baz/a/b",
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

			wantStatusCode := test.wantStatusCode
			if wantStatusCode == 0 {
				wantStatusCode = http.StatusOK
			}
			assert.Equal(t, wantStatusCode, recorder.Code)

			if wantStatusCode != http.StatusOK {
				return
			}

			assert.Equal(t, test.wantURL, req.URL.String())
			assert.Equal(t, test.wantHost, req.Host)
		})
	}
}
