package redirect

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

func TestRedirectRegexHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		config         dynamic.RedirectRegex
		method         string
		url            string
		headers        map[string]string
		secured        bool
		expectedURL    string
		expectedStatus int
		errorExpected  bool
	}{
		{
			desc: "simple redirection",
			config: dynamic.RedirectRegex{
				Regex:       `^(?:http?:\/\/)(foo)(\.com)(:\d+)(.*)$`,
				Replacement: "https://${1}bar$2:443$4",
			},
			url:            "http://foo.com:80",
			expectedURL:    "https://foobar.com:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "URL doesn't match regex",
			config: dynamic.RedirectRegex{
				Regex:       `^(?:http?:\/\/)(foo)(\.com)(:\d+)(.*)$`,
				Replacement: "https://${1}bar$2:443$4",
			},
			url:            "http://bar.com:80",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "invalid rewritten URL",
			config: dynamic.RedirectRegex{
				Regex:       `^(.*)$`,
				Replacement: "http://192.168.0.%31/",
			},
			url:            "http://foo.com:80",
			expectedStatus: http.StatusBadGateway,
		},
		{
			desc: "invalid regex",
			config: dynamic.RedirectRegex{
				Regex:       `^(.*`,
				Replacement: "$1",
			},
			url:           "http://foo.com:80",
			errorExpected: true,
		},
		{
			desc: "HTTP to HTTPS permanent",
			config: dynamic.RedirectRegex{
				Regex:       `^http://`,
				Replacement: "https://$1",
				Permanent:   true,
			},
			url:            "http://foo",
			expectedURL:    "https://foo",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc: "HTTPS to HTTP permanent",
			config: dynamic.RedirectRegex{
				Regex:       `https://foo`,
				Replacement: "http://foo",
				Permanent:   true,
			},
			secured:        true,
			url:            "https://foo",
			expectedURL:    "http://foo",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc: "HTTP to HTTPS",
			config: dynamic.RedirectRegex{
				Regex:       `http://foo:80`,
				Replacement: "https://foo:443",
			},
			url:            "http://foo:80",
			expectedURL:    "https://foo:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTPS, with X-Forwarded-Proto",
			config: dynamic.RedirectRegex{
				Regex:       `http://foo:80`,
				Replacement: "https://foo:443",
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
			config: dynamic.RedirectRegex{
				Regex:       `https://foo:443`,
				Replacement: "http://foo:80",
			},
			secured:        true,
			url:            "https://foo:443",
			expectedURL:    "http://foo:80",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTP",
			config: dynamic.RedirectRegex{
				Regex:       `http://foo:80`,
				Replacement: "http://foo:88",
			},
			url:            "http://foo:80",
			expectedURL:    "http://foo:88",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTP to HTTP POST",
			config: dynamic.RedirectRegex{
				Regex:       `^http://`,
				Replacement: "https://$1",
			},
			url:            "http://foo",
			method:         http.MethodPost,
			expectedURL:    "https://foo",
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			desc: "HTTP to HTTP POST permanent",
			config: dynamic.RedirectRegex{
				Regex:       `^http://`,
				Replacement: "https://$1",
				Permanent:   true,
			},
			url:            "http://foo",
			method:         http.MethodPost,
			expectedURL:    "https://foo",
			expectedStatus: http.StatusPermanentRedirect,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := NewRedirectRegex(context.Background(), next, test.config, "traefikTest")

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
				case http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
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
