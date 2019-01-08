package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/tls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntryPointHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		entryPoint     *configuration.EntryPoint
		permanent      bool
		method         string
		url            string
		expectedURL    string
		expectedStatus int
		errorExpected  bool
	}{
		{
			desc:           "HTTP to HTTPS",
			entryPoint:     &configuration.EntryPoint{Address: ":443", TLS: &tls.TLS{}},
			url:            "http://foo:80",
			expectedURL:    "https://foo:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc:           "HTTPS to HTTP",
			entryPoint:     &configuration.EntryPoint{Address: ":80"},
			url:            "https://foo:443",
			expectedURL:    "http://foo:80",
			expectedStatus: http.StatusFound,
		},
		{
			desc:           "HTTP to HTTP",
			entryPoint:     &configuration.EntryPoint{Address: ":88"},
			url:            "http://foo:80",
			expectedURL:    "http://foo:88",
			expectedStatus: http.StatusFound,
		},
		{
			desc:           "HTTP to HTTPS permanent",
			entryPoint:     &configuration.EntryPoint{Address: ":443", TLS: &tls.TLS{}},
			permanent:      true,
			url:            "http://foo:80",
			expectedURL:    "https://foo:443",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "HTTPS to HTTP permanent",
			entryPoint:     &configuration.EntryPoint{Address: ":80"},
			permanent:      true,
			url:            "https://foo:443",
			expectedURL:    "http://foo:80",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			desc:           "HTTP to HTTP POST",
			entryPoint:     &configuration.EntryPoint{Address: ":80"},
			permanent:      false,
			url:            "http://foo:90",
			method:         http.MethodPost,
			expectedURL:    "http://foo:80",
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			desc:           "HTTP to HTTP POST permanent",
			entryPoint:     &configuration.EntryPoint{Address: ":80"},
			permanent:      true,
			url:            "http://foo:90",
			method:         http.MethodPost,
			expectedURL:    "http://foo:80",
			expectedStatus: http.StatusPermanentRedirect,
		},
		{
			desc:          "invalid address",
			entryPoint:    &configuration.EntryPoint{Address: ":foo", TLS: &tls.TLS{}},
			url:           "http://foo:80",
			errorExpected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler, err := NewEntryPointHandler(test.entryPoint, test.permanent)

			if test.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				recorder := httptest.NewRecorder()
				method := http.MethodGet
				if test.method != "" {
					method = test.method
				}
				r := testhelpers.MustNewRequest(method, test.url, nil)
				handler.ServeHTTP(recorder, r, nil)

				location, err := recorder.Result().Location()
				require.NoError(t, err)

				assert.Equal(t, test.expectedURL, location.String())
				assert.Equal(t, test.expectedStatus, recorder.Code)
			}
		})
	}
}

func TestNewRegexHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		regex          string
		replacement    string
		permanent      bool
		url            string
		expectedURL    string
		expectedStatus int
		errorExpected  bool
	}{
		{
			desc:           "simple redirection",
			regex:          `^(?:http?:\/\/)(foo)(\.com)(:\d+)(.*)$`,
			replacement:    "https://${1}bar$2:443$4",
			url:            "http://foo.com:80",
			expectedURL:    "https://foobar.com:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc:           "use request header",
			regex:          `^(?:http?:\/\/)(foo)(\.com)(:\d+)(.*)$`,
			replacement:    `https://${1}{{ .Request.Header.Get "X-Foo" }}$2:443$4`,
			url:            "http://foo.com:80",
			expectedURL:    "https://foobar.com:443",
			expectedStatus: http.StatusFound,
		},
		{
			desc:           "URL doesn't match regex",
			regex:          `^(?:http?:\/\/)(foo)(\.com)(:\d+)(.*)$`,
			replacement:    "https://${1}bar$2:443$4",
			url:            "http://bar.com:80",
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "invalid rewritten URL",
			regex:          `^(.*)$`,
			replacement:    "http://192.168.0.%31/",
			url:            "http://foo.com:80",
			expectedStatus: http.StatusBadGateway,
		},
		{
			desc:          "invalid regex",
			regex:         `^(.*`,
			replacement:   "$1",
			url:           "http://foo.com:80",
			errorExpected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler, err := NewRegexHandler(test.regex, test.replacement, test.permanent)

			if test.errorExpected {
				require.Nil(t, handler)
				require.Error(t, err)
			} else {
				require.NotNil(t, handler)
				require.NoError(t, err)

				recorder := httptest.NewRecorder()
				r := testhelpers.MustNewRequest(http.MethodGet, test.url, nil)
				r.Header.Set("X-Foo", "bar")
				next := func(rw http.ResponseWriter, req *http.Request) {}
				handler.ServeHTTP(recorder, r, next)

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
