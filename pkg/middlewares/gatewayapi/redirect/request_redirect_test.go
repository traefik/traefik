package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"k8s.io/utils/ptr"
)

func TestRequestRedirectHandler(t *testing.T) {
	testCases := []struct {
		desc       string
		config     dynamic.RequestRedirect
		url        string
		wantURL    string
		wantStatus int
		wantErr    bool
	}{
		{
			desc: "wrong status code",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("/baz"),
				StatusCode: http.StatusOK,
			},
			url:     "http://foo.com:80/foo/bar",
			wantErr: true,
		},
		{
			desc: "replace path",
			config: dynamic.RequestRedirect{
				Path: ptr.To("/baz"),
			},
			url:        "http://foo.com:80/foo/bar",
			wantURL:    "http://foo.com:80/baz",
			wantStatus: http.StatusFound,
		},
		{
			desc: "replace path without trailing slash",
			config: dynamic.RequestRedirect{
				Path: ptr.To("/baz"),
			},
			url:        "http://foo.com:80/foo/bar/",
			wantURL:    "http://foo.com:80/baz",
			wantStatus: http.StatusFound,
		},
		{
			desc: "replace path with trailing slash",
			config: dynamic.RequestRedirect{
				Path: ptr.To("/baz/"),
			},
			url:        "http://foo.com:80/foo/bar",
			wantURL:    "http://foo.com:80/baz/",
			wantStatus: http.StatusFound,
		},
		{
			desc: "only hostname",
			config: dynamic.RequestRedirect{
				Hostname: ptr.To("bar.com"),
			},
			url:        "http://foo.com:8080/foo/",
			wantURL:    "http://bar.com:8080/foo/",
			wantStatus: http.StatusFound,
		},
		{
			desc: "replace prefix path",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:        "http://foo.com:80/foo/bar",
			wantURL:    "http://foo.com:80/baz/bar",
			wantStatus: http.StatusFound,
		},
		{
			desc: "replace prefix path with trailing slash",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:        "http://foo.com:80/foo/bar/",
			wantURL:    "http://foo.com:80/baz/bar/",
			wantStatus: http.StatusFound,
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:        "http://foo.com:80/foo/bar",
			wantURL:    "http://foo.com:80/baz/bar",
			wantStatus: http.StatusFound,
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo/"),
			},
			url:        "http://foo.com:80/foo/bar",
			wantURL:    "http://foo.com:80/baz/bar",
			wantStatus: http.StatusFound,
		},
		{
			desc: "simple redirection",
			config: dynamic.RequestRedirect{
				Scheme:   ptr.To("https"),
				Hostname: ptr.To("foobar.com"),
				Port:     ptr.To("443"),
			},
			url:        "http://foo.com:80",
			wantURL:    "https://foobar.com:443",
			wantStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS permanent",
			config: dynamic.RequestRedirect{
				Scheme:     ptr.To("https"),
				StatusCode: http.StatusMovedPermanently,
			},
			url:        "http://foo",
			wantURL:    "https://foo",
			wantStatus: http.StatusMovedPermanently,
		},
		{
			desc: "HTTPS to HTTP permanent",
			config: dynamic.RequestRedirect{
				Scheme:     ptr.To("http"),
				StatusCode: http.StatusMovedPermanently,
			},
			url:        "https://foo",
			wantURL:    "http://foo",
			wantStatus: http.StatusMovedPermanently,
		},
		{
			desc: "HTTP to HTTPS",
			config: dynamic.RequestRedirect{
				Scheme: ptr.To("https"),
				Port:   ptr.To("443"),
			},
			url:        "http://foo:80",
			wantURL:    "https://foo:443",
			wantStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS, with X-Forwarded-Proto",
			config: dynamic.RequestRedirect{
				Scheme: ptr.To("https"),
				Port:   ptr.To("443"),
			},
			url:        "http://foo:80",
			wantURL:    "https://foo:443",
			wantStatus: http.StatusFound,
		},
		{
			desc: "HTTPS to HTTP",
			config: dynamic.RequestRedirect{
				Scheme: ptr.To("http"),
				Port:   ptr.To("80"),
			},
			url:        "https://foo:443",
			wantURL:    "http://foo:80",
			wantStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTP",
			config: dynamic.RequestRedirect{
				Scheme: ptr.To("http"),
				Port:   ptr.To("88"),
			},
			url:        "http://foo:80",
			wantURL:    "http://foo:88",
			wantStatus: http.StatusFound,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			handler, err := NewRequestRedirect(t.Context(), next, test.config, "traefikTest")
			if test.wantErr {
				require.Error(t, err)
				require.Nil(t, handler)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, handler)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, test.url, nil)

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, test.wantStatus, recorder.Code)
			switch test.wantStatus {
			case http.StatusMovedPermanently, http.StatusFound:
				location, err := recorder.Result().Location()
				require.NoError(t, err)

				assert.Equal(t, test.wantURL, location.String())
			default:
				location, err := recorder.Result().Location()
				require.Errorf(t, err, "Location %v", location)
			}
		})
	}
}
