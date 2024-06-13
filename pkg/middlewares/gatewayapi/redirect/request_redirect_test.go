package redirect

import (
	"context"
	"crypto/tls"
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
		desc           string
		config         dynamic.RequestRedirect
		method         string
		url            string
		headers        map[string]string
		secured        bool
		expectedURL    string
		expectedStatus int
		errorExpected  bool
	}{
		{
			desc: "wrong status code",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("/baz"),
				StatusCode: http.StatusOK,
			},
			url:           "http://foo.com:80/foo/bar",
			errorExpected: true,
		},
		{
			desc: "replace path",
			config: dynamic.RequestRedirect{
				Path: ptr.To("/baz"),
			},
			url:            "http://foo.com:80/foo/bar",
			expectedURL:    "http://foo.com:80/baz",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "replace path without trailing slash",
			config: dynamic.RequestRedirect{
				Path: ptr.To("/baz"),
			},
			url:            "http://foo.com:80/foo/bar/",
			expectedURL:    "http://foo.com:80/baz",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "replace path with trailing slash",
			config: dynamic.RequestRedirect{
				Path: ptr.To("/baz/"),
			},
			url:            "http://foo.com:80/foo/bar",
			expectedURL:    "http://foo.com:80/baz/",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "only hostname",
			config: dynamic.RequestRedirect{
				Hostname: ptr.To("bar.com"),
			},
			url:            "http://foo.com:8080/foo/",
			expectedURL:    "http://bar.com:8080/foo/",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "replace prefix path",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:            "http://foo.com:80/foo/bar",
			expectedURL:    "http://foo.com:80/baz/bar",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "replace prefix path with trailing slash",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:            "http://foo.com:80/foo/bar/",
			expectedURL:    "http://foo.com:80/baz/bar/",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("baz"),
				PathPrefix: ptr.To("/foo"),
			},
			url:            "http://foo.com:80/foo/bar",
			expectedURL:    "http://foo.com:80/baz/bar",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "replace prefix path without slash prefix",
			config: dynamic.RequestRedirect{
				Path:       ptr.To("/baz"),
				PathPrefix: ptr.To("/foo/"),
			},
			url:            "http://foo.com:80/foo/bar",
			expectedURL:    "http://foo.com:80/baz/bar",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "simple redirection",
			config: dynamic.RequestRedirect{
				Scheme:   ptr.To("https"),
				Hostname: ptr.To("foobar.com"),
				Port:     ptr.To("443"),
			},
			url:            "http://foo.com:80",
			expectedURL:    "https://foobar.com:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS permanent",
			config: dynamic.RequestRedirect{
				Scheme:     ptr.To("https"),
				StatusCode: http.StatusMovedPermanently,
			},
			url:            "http://foo",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc: "HTTPS to HTTP permanent",
			config: dynamic.RequestRedirect{
				Scheme:     ptr.To("http"),
				StatusCode: http.StatusMovedPermanently,
			},
			secured:        true,
			url:            "https://foo",
			expectedURL:    "http://foo",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc: "HTTP to HTTPS",
			config: dynamic.RequestRedirect{
				Scheme: ptr.To("https"),
				Port:   ptr.To("443"),
			},
			url:            "http://foo:80",
			expectedURL:    "https://foo:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS, with X-Forwarded-Proto",
			config: dynamic.RequestRedirect{
				Scheme: ptr.To("https"),
				Port:   ptr.To("443"),
			},
			url: "http://foo:80",
			headers: map[string]string{
				"X-Forwarded-Proto": "https",
			},
			expectedURL:    "https://foo:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTPS to HTTP",
			config: dynamic.RequestRedirect{
				Scheme: ptr.To("http"),
				Port:   ptr.To("80"),
			},
			secured:        true,
			url:            "https://foo:443",
			expectedURL:    "http://foo:80",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTP",
			config: dynamic.RequestRedirect{
				Scheme: ptr.To("http"),
				Port:   ptr.To("88"),
			},
			url:            "http://foo:80",
			expectedURL:    "http://foo:88",
			expectedStatus: http.StatusFound,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := NewRequestRedirect(context.Background(), next, test.config, "traefikTest")

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
				if test.secured {
					req.TLS = &tls.ConnectionState{}
				}

				for k, v := range test.headers {
					req.Header.Set(k, v)
				}

				req.Header.Set("X-Foo", "bar")
				handler.ServeHTTP(recorder, req)

				assert.Equal(t, test.expectedStatus, recorder.Code)
				switch test.expectedStatus {
				case http.StatusMovedPermanently, http.StatusFound:
					location, err := recorder.Result().Location()
					require.NoError(t, err)

					assert.Equal(t, test.expectedURL, location.String())
				default:
					location, err := recorder.Result().Location()
					require.Errorf(t, err, "Location %v", location)
				}
			}
		})
	}
}
