package httputil

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func Test_rewriteRequestBuilder(t *testing.T) {
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
		notAppendXFF    bool
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
			name:           "Basic proxy - notAppendXFF",
			target:         testhelpers.MustParseURL("http://example.com"),
			passHostHeader: false,
			preservePath:   false,
			incomingURL:    "http://localhost/test?param=value",
			expectedScheme: "http",
			expectedHost:   "example.com",
			expectedPath:   "/test",
			expectedQuery:  "param=value",
			notAppendXFF:   true,
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

			rewriteRequest := rewriteRequestBuilder(test.target, test.passHostHeader, test.preservePath)

			ctx := t.Context()
			if test.notAppendXFF {
				ctx = SetNotAppendXFF(ctx)
			}

			reqIn := httptest.NewRequest(http.MethodGet, test.incomingURL, http.NoBody)
			reqIn = reqIn.WithContext(ctx)
			reqIn.Header.Add("X-Forwarded-For", "1.2.3.4")
			reqIn.RemoteAddr = "127.0.0.1:1234"

			reqOut := httptest.NewRequest(http.MethodGet, test.incomingURL, http.NoBody)
			pr := &httputil.ProxyRequest{
				In:  reqIn,
				Out: reqOut,
			}
			rewriteRequest(pr)

			if test.notAppendXFF {
				assert.Equal(t, "1.2.3.4", reqOut.Header.Get("X-Forwarded-For"))
			} else {
				// When not disabled, X-Forwarded-For should have RemoteAddr appended
				assert.Equal(t, "1.2.3.4, 127.0.0.1", reqOut.Header.Get("X-Forwarded-For"))
			}
			assert.Equal(t, test.expectedScheme, reqOut.URL.Scheme)
			assert.Equal(t, test.expectedHost, reqOut.Host)
			assert.Equal(t, test.expectedPath, reqOut.URL.Path)
			assert.Equal(t, test.expectedRawPath, reqOut.URL.RawPath)
			assert.Equal(t, test.expectedQuery, reqOut.URL.RawQuery)
			assert.Empty(t, reqOut.RequestURI)
			assert.Equal(t, "HTTP/1.1", reqOut.Proto)
			assert.Equal(t, 1, reqOut.ProtoMajor)
			assert.Equal(t, 1, reqOut.ProtoMinor)
			assert.False(t, !test.passHostHeader && reqOut.Host != reqOut.URL.Host)
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

// timeoutError is a net.Error that reports a timeout.
type timeoutError struct{ msg string }

func (e *timeoutError) Error() string   { return e.msg }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return false }

// netError is a net.Error that does not report a timeout.
type netError struct{ msg string }

func (e *netError) Error() string   { return e.msg }
func (e *netError) Timeout() bool   { return false }
func (e *netError) Temporary() bool { return false }

func TestErrorHandlerWithContext_OriginError(t *testing.T) {
	testCases := []struct {
		desc               string
		err                error
		wantStatus         int
		wantOriginError    bool
		wantOriginErrorMsg string
	}{
		{
			desc:               "io.EOF",
			err:                io.EOF,
			wantStatus:         http.StatusBadGateway,
			wantOriginError:    true,
			wantOriginErrorMsg: io.EOF.Error(),
		},
		{
			desc:               "network timeout",
			err:                &timeoutError{msg: "i/o timeout"},
			wantStatus:         http.StatusGatewayTimeout,
			wantOriginError:    true,
			wantOriginErrorMsg: "i/o timeout",
		},
		{
			desc:               "network non-timeout",
			err:                &netError{msg: "connection refused"},
			wantStatus:         http.StatusBadGateway,
			wantOriginError:    true,
			wantOriginErrorMsg: "connection refused",
		},
		{
			desc:            "context canceled",
			err:             context.Canceled,
			wantStatus:      StatusClientClosedRequest,
			wantOriginError: false,
		},
		{
			desc:               "TLS config error",
			err:                &tls.RecordHeaderError{RecordHeader: [5]byte{}, Conn: nil, Msg: "bad TLS record"},
			wantStatus:         http.StatusInternalServerError,
			wantOriginError:    true,
			wantOriginErrorMsg: (&tls.RecordHeaderError{RecordHeader: [5]byte{}, Msg: "bad TLS record"}).Error(),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			logData := &accesslog.LogData{Core: accesslog.CoreLogData{}}
			ctx := context.WithValue(t.Context(), accesslog.DataTableKey, logData)

			rw := httptest.NewRecorder()
			ErrorHandlerWithContext(ctx, rw, test.err)

			assert.Equal(t, test.wantStatus, rw.Code)

			if test.wantOriginError {
				val, ok := logData.Core[accesslog.OriginError]
				assert.True(t, ok)
				assert.Equal(t, test.wantOriginErrorMsg, val)
			} else {
				_, ok := logData.Core[accesslog.OriginError]
				assert.False(t, ok)
			}
		})
	}
}

func TestErrorHandlerWithContext_OriginError_NoLogData(t *testing.T) {
	// No LogData in context: the handler must not panic.
	rw := httptest.NewRecorder()
	assert.NotPanics(t, func() {
		ErrorHandlerWithContext(t.Context(), rw, io.EOF)
	})
	assert.Equal(t, http.StatusBadGateway, rw.Code)
}
