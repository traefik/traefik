package service

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

type staticTransport struct {
	res *http.Response
}

func (t *staticTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.res, nil
}

func BenchmarkProxy(b *testing.B) {
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	w := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/", nil)

	pool := newBufferPool()
	handler, _ := buildProxy(pointer(false), nil, &staticTransport{res}, pool)

	b.ReportAllocs()
	for range b.N {
		handler.ServeHTTP(w, req)
	}
}

func TestIsTLSConfigError(t *testing.T) {
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
