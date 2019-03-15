package redirect

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectSchemeHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		config         config.RedirectScheme
		method         string
		url            string
		secured        bool
		expectedURL    string
		expectedStatus int
		errorExpected  bool
	}{
		{
			desc:          "Without scheme",
			config:        config.RedirectScheme{},
			url:           "http://foo",
			errorExpected: true,
		},
		{
			desc: "HTTP to HTTPS",
			config: config.RedirectScheme{
				Scheme: "https",
			},
			url:            "http://foo",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP with port to HTTPS without port",
			config: config.RedirectScheme{
				Scheme: "https",
			},
			url:            "http://foo:8080",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP without port to HTTPS with port",
			config: config.RedirectScheme{
				Scheme: "https",
				Port:   "8443",
			},
			url:            "http://foo",
			expectedURL:    "https://foo:8443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP with port to HTTPS with port",
			config: config.RedirectScheme{
				Scheme: "https",
				Port:   "8443",
			},
			url:            "http://foo:8000",
			expectedURL:    "https://foo:8443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTPS with port to HTTPS with port",
			config: config.RedirectScheme{
				Scheme: "https",
				Port:   "8443",
			},
			url:            "https://foo:8000",
			expectedURL:    "https://foo:8443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTPS with port to HTTPS without port",
			config: config.RedirectScheme{
				Scheme: "https",
			},
			url:            "https://foo:8000",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "redirection to HTTPS without port from an URL already in https",
			config: config.RedirectScheme{
				Scheme: "https",
			},
			url:            "https://foo:8000/theother",
			expectedURL:    "https://foo/theother",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS permanent",
			config: config.RedirectScheme{
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
			config: config.RedirectScheme{
				Scheme: "http",
				Port:   "80",
			},
			url:            "http://foo:80",
			expectedURL:    "http://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to wss",
			config: config.RedirectScheme{
				Scheme: "wss",
				Port:   "9443",
			},
			url:            "http://foo",
			expectedURL:    "wss://foo:9443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to wss without port",
			config: config.RedirectScheme{
				Scheme: "wss",
			},
			url:            "http://foo",
			expectedURL:    "wss://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP with port to wss without port",
			config: config.RedirectScheme{
				Scheme: "wss",
			},
			url:            "http://foo:5678",
			expectedURL:    "wss://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS without port",
			config: config.RedirectScheme{
				Scheme: "https",
			},
			url:            "http://foo:443",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP port redirection",
			config: config.RedirectScheme{
				Scheme: "http",
				Port:   "8181",
			},
			url:            "http://foo:8080",
			expectedURL:    "http://foo:8181",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTPS with port 80 to HTTPS without port",
			config: config.RedirectScheme{
				Scheme: "https",
			},
			url:            "https://foo:80",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusFound,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := NewRedirectScheme(context.Background(), next, test.config, "traefikTest")

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
				r := httptest.NewRequest(method, test.url, nil)

				if test.secured {
					r.TLS = &tls.ConnectionState{}
				}
				r.Header.Set("X-Foo", "bar")
				handler.ServeHTTP(recorder, r)

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

				schemeRegex := `^(https?):\/\/([\w\._-]+)(:\d+)?(.*)$`
				re, _ := regexp.Compile(schemeRegex)

				if re.Match([]byte(test.url)) {
					match := re.FindStringSubmatch(test.url)
					r.RequestURI = match[4]

					handler.ServeHTTP(recorder, r)

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
