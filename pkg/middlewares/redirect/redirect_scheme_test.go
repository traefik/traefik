package redirect

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestRedirectSchemeHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		config         dynamic.RedirectScheme
		method         string
		url            string
		headers        map[string]string
		secured        bool
		expectedURL    string
		expectedStatus int
		errorExpected  bool
	}{
		{
			desc:          "Without scheme",
			config:        dynamic.RedirectScheme{},
			url:           "http://foo",
			errorExpected: true,
		},
		{
			desc: "HTTP to HTTPS",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url:            "http://foo",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS, with X-Forwarded-Proto",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url: "http://foo",
			headers: map[string]string{
				"X-Forwarded-Proto": "http",
			},
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS, with X-Forwarded-Proto to HTTPS",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url: "http://foo",
			headers: map[string]string{
				"X-Forwarded-Proto": "https",
			},
			expectedStatus: http.StatusOK,
		},
		{
			desc: "HTTP to HTTPS, with X-Forwarded-Proto to unknown value",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url: "http://foo",
			headers: map[string]string{
				"X-Forwarded-Proto": "bar",
			},
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS, with X-Forwarded-Proto to ws",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url: "http://foo",
			headers: map[string]string{
				"X-Forwarded-Proto": "ws",
			},
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS, with X-Forwarded-Proto to wss",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url: "http://foo",
			headers: map[string]string{
				"X-Forwarded-Proto": "wss",
			},
			expectedStatus: http.StatusOK,
		},
		{
			desc: "HTTP with port to HTTPS without port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url:            "http://foo:8080",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP without port to HTTPS with port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
				Port:   "8443",
			},
			url:            "http://foo",
			expectedURL:    "https://foo:8443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP with port to HTTPS with port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
				Port:   "8443",
			},
			url:            "http://foo:8000",
			expectedURL:    "https://foo:8443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTPS with port to HTTPS with port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
				Port:   "8443",
			},
			url:            "https://foo:8000",
			expectedURL:    "https://foo:8443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTPS with port to HTTPS without port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url:            "https://foo:8000",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "redirection to HTTPS without port from an URL already in https",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url:            "https://foo:8000/theother",
			expectedURL:    "https://foo/theother",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS permanent",
			config: dynamic.RedirectScheme{
				Scheme:    "https",
				Port:      "8443",
				Permanent: true,
			},
			url:            "http://foo",
			expectedURL:    "https://foo:8443",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc: "to HTTP 80",
			config: dynamic.RedirectScheme{
				Scheme: "http",
				Port:   "80",
			},
			url:            "http://foo:80",
			expectedURL:    "http://foo:80",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "to HTTPS 443",
			config: dynamic.RedirectScheme{
				Scheme: "https",
				Port:   "443",
			},
			url:            "https://foo:443",
			expectedURL:    "https://foo:443",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "HTTP to wss",
			config: dynamic.RedirectScheme{
				Scheme: "wss",
				Port:   "9443",
			},
			url:            "http://foo",
			expectedURL:    "wss://foo:9443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to wss without port",
			config: dynamic.RedirectScheme{
				Scheme: "wss",
			},
			url:            "http://foo",
			expectedURL:    "wss://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP with port to wss without port",
			config: dynamic.RedirectScheme{
				Scheme: "wss",
			},
			url:            "http://foo:5678",
			expectedURL:    "wss://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS without port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url:            "http://foo:443",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP port redirection",
			config: dynamic.RedirectScheme{
				Scheme: "http",
				Port:   "8181",
			},
			url:            "http://foo:8080",
			expectedURL:    "http://foo:8181",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTPS with port 80 to HTTPS without port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url:            "https://foo:80",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "IPV6 HTTP to HTTPS redirection without port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url:            "http://[::1]",
			expectedURL:    "https://[::1]",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "IPV6 HTTP to HTTPS redirection with port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
				Port:   "8443",
			},
			url:            "http://[::1]",
			expectedURL:    "https://[::1]:8443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "IPV6 HTTP with port 80 to HTTPS redirection without port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
			},
			url:            "http://[::1]:80",
			expectedURL:    "https://[::1]",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "IPV6 HTTP with port 80 to HTTPS redirection with port",
			config: dynamic.RedirectScheme{
				Scheme: "https",
				Port:   "8443",
			},
			url:            "http://[::1]:80",
			expectedURL:    "https://[::1]:8443",
			expectedStatus: http.StatusFound,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := NewRedirectScheme(t.Context(), next, test.config, "traefikTest")

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

				if test.secured {
					req.TLS = &tls.ConnectionState{}
				}
				req.Header.Set("X-Foo", "bar")
				handler.ServeHTTP(recorder, req)

				assert.Equal(t, test.expectedStatus, recorder.Code)

				switch test.expectedStatus {
				case http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
					location, err := recorder.Result().Location()
					require.NoError(t, err)

					assert.Equal(t, test.expectedURL, location.String())
				default:
					location, err := recorder.Result().Location()
					require.Errorf(t, err, "Location %v", location)
				}

				schemeRegex := `^(https?):\/\/(\[[\w:.]+\]|[\w\._-]+)?(:\d+)?(.*)$`
				re, _ := regexp.Compile(schemeRegex)

				if re.MatchString(test.url) {
					match := re.FindStringSubmatch(test.url)
					req.RequestURI = match[4]

					handler.ServeHTTP(recorder, req)

					assert.Equal(t, test.expectedStatus, recorder.Code)
					if test.expectedStatus == http.StatusMovedPermanently ||
						test.expectedStatus == http.StatusFound ||
						test.expectedStatus == http.StatusTemporaryRedirect ||
						test.expectedStatus == http.StatusPermanentRedirect {
						location, err := recorder.Result().Location()
						require.NoError(t, err)

						assert.Equal(t, test.expectedURL, location.String())
					} else {
						location, err := recorder.Result().Location()
						require.Errorf(t, err, "Location %v", location)
					}
				}
			}
		})
	}
}
