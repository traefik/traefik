package httputil

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func Test_rewriteBuilder(t *testing.T) {
	tests := []struct {
		name            string
		target          *url.URL
		passHostHeader  bool
		preservePath    bool
		incomingURL     string
		expectedScheme  string
		expectedHost    string
		expectedPath    string
		expectedRawPath string
		expectedQuery   string
	}{
		{
			name:           "Basic proxy",
			target:         testhelpers.MustParseURL("http://example.com"),
			passHostHeader: false,
			preservePath:   false,
			incomingURL:    "http://localhost/test?param=value",
			expectedScheme: "http",
			expectedHost:   "example.com",
			expectedPath:   "/test",
			expectedQuery:  "param=value",
		},
		{
			name:           "HTTPS target",
			target:         testhelpers.MustParseURL("https://secure.example.com"),
			passHostHeader: false,
			preservePath:   false,
			incomingURL:    "http://localhost/secure",
			expectedScheme: "https",
			expectedHost:   "secure.example.com",
			expectedPath:   "/secure",
		},
		{
			name:           "PassHostHeader",
			target:         testhelpers.MustParseURL("http://example.com"),
			passHostHeader: true,
			preservePath:   false,
			incomingURL:    "http://original.host/test",
			expectedScheme: "http",
			expectedHost:   "original.host",
			expectedPath:   "/test",
		},
		{
			name:            "Preserve path",
			target:          testhelpers.MustParseURL("http://example.com/base"),
			passHostHeader:  false,
			preservePath:    true,
			incomingURL:     "http://localhost/foo%2Fbar",
			expectedScheme:  "http",
			expectedHost:    "example.com",
			expectedPath:    "/base/foo/bar",
			expectedRawPath: "/base/foo%2Fbar",
		},
		{
			name:           "Handle semicolons in query",
			target:         testhelpers.MustParseURL("http://example.com"),
			passHostHeader: false,
			preservePath:   false,
			incomingURL:    "http://localhost/test?param1=value1;param2=value2",
			expectedScheme: "http",
			expectedHost:   "example.com",
			expectedPath:   "/test",
			expectedQuery:  "param1=value1&param2=value2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, test.incomingURL, http.NoBody)
			outReq := req.Clone(req.Context())

			rewrite := rewriteBuilder(test.target, test.passHostHeader, test.preservePath)
			rewrite(&httputil.ProxyRequest{
				In:  req,
				Out: outReq,
			})

			assert.Equal(t, test.expectedScheme, outReq.URL.Scheme)
			assert.Equal(t, test.expectedHost, outReq.Host)
			assert.Equal(t, test.expectedPath, outReq.URL.Path)
			assert.Equal(t, test.expectedRawPath, outReq.URL.RawPath)
			assert.Equal(t, test.expectedQuery, outReq.URL.RawQuery)
			assert.Empty(t, outReq.RequestURI)
			assert.Equal(t, "HTTP/1.1", outReq.Proto)
			assert.Equal(t, 1, outReq.ProtoMajor)
			assert.Equal(t, 1, outReq.ProtoMinor)
			assert.False(t, !test.passHostHeader && outReq.Host != outReq.URL.Host)
		})
	}
}

func Test_isTLSConfigError(t *testing.T) {
	testCases := []struct {
		desc     string
		err      error
		expected bool
	}{
		{
			desc: "nil",
		},
		{
			desc: "TLS ECHRejectionError",
			err:  &tls.ECHRejectionError{},
		},
		{
			desc: "TLS AlertError",
			err:  tls.AlertError(0),
		},
		{
			desc: "Random error",
			err:  errors.New("random error"),
		},
		{
			desc:     "TLS RecordHeaderError",
			err:      tls.RecordHeaderError{},
			expected: true,
		},
		{
			desc:     "TLS CertificateVerificationError",
			err:      &tls.CertificateVerificationError{},
			expected: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := isTLSConfigError(test.err)
			require.Equal(t, test.expected, actual)
		})
	}
}
