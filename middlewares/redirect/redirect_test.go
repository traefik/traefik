package redirect

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegexHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		config         config.Redirect
		url            string
		expectedURL    string
		expectedStatus int
		errorExpected  bool
		secured        bool
	}{
		{
			desc: "simple redirection",
			config: config.Redirect{
				Regex:       `^(?:http?:\/\/)(foo)(\.com)(:\d+)(.*)$`,
				Replacement: "https://${1}bar$2:443$4",
			},
			url:            "http://foo.com:80",
			expectedURL:    "https://foobar.com:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "use request header",
			config: config.Redirect{
				Regex:       `^(?:http?:\/\/)(foo)(\.com)(:\d+)(.*)$`,
				Replacement: `https://${1}{{ .Request.Header.Get "X-Foo" }}$2:443$4`,
			},
			url:            "http://foo.com:80",
			expectedURL:    "https://foobar.com:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "URL doesn't match regex",
			config: config.Redirect{
				Regex:       `^(?:http?:\/\/)(foo)(\.com)(:\d+)(.*)$`,
				Replacement: "https://${1}bar$2:443$4",
			},
			url:            "http://bar.com:80",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "invalid rewritten URL",
			config: config.Redirect{
				Regex:       `^(.*)$`,
				Replacement: "http://192.168.0.%31/",
			},
			url:            "http://foo.com:80",
			expectedStatus: http.StatusBadGateway,
		},
		{
			desc: "invalid regex",
			config: config.Redirect{
				Regex:       `^(.*`,
				Replacement: "$1",
			},
			url:           "http://foo.com:80",
			errorExpected: true,
		},
		{
			desc: "HTTP to HTTPS permanent",
			config: config.Redirect{
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
			config: config.Redirect{
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
			config: config.Redirect{
				Regex:       `http://foo:80`,
				Replacement: "https://foo:443",
			},
			url:            "http://foo:80",
			expectedURL:    "https://foo:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc: "HTTPS to HTTP",
			config: config.Redirect{
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
			config: config.Redirect{
				Regex:       `http://foo:80`,
				Replacement: "http://foo:88",
			},
			url:            "http://foo:80",
			expectedURL:    "http://foo:88",
			expectedStatus: http.StatusFound,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := New(context.Background(), next, test.config, "traefikTest")

			if test.errorExpected {
				require.Error(t, err)
				require.Nil(t, handler)
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)

				recorder := httptest.NewRecorder()
				r := testhelpers.MustNewRequest(http.MethodGet, test.url, nil)
				if test.secured {
					r.TLS = &tls.ConnectionState{}
				}
				r.Header.Set("X-Foo", "bar")
				handler.ServeHTTP(recorder, r)

				if test.expectedStatus == http.StatusMovedPermanently || test.expectedStatus == http.StatusFound {
					assert.Equal(t, test.expectedStatus, recorder.Code)

					location, err := recorder.Result().Location()
					require.NoError(t, err)

					assert.Equal(t, test.expectedURL, location.String())
				} else {
					assert.Equal(t, test.expectedStatus, recorder.Code)

					location, err := recorder.Result().Location()
					require.Errorf(t, err, "Location %v", location)
				}
			}
		})
	}
}
