package service

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
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
	handler, _ := buildProxy(new(false), nil, &staticTransport{res}, pool)

	b.ReportAllocs()
	for b.Loop() {
		handler.ServeHTTP(w, req)
	}
}

func TestProxy_ConnectRejected(t *testing.T) {
	testCases := []struct {
		desc  string
		http2 bool
	}{
		{
			desc: "HTTP/1.1",
		},
		{
			desc:  "HTTP/2",
			http2: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var backendCalled atomic.Bool
			backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				backendCalled.Store(true)
			}))
			t.Cleanup(backend.Close)

			backendURL, err := url.Parse(backend.URL)
			require.NoError(t, err)

			proxy, err := buildProxy(new(true), nil, http.DefaultTransport, newBufferPool())
			require.NoError(t, err)

			// The load balancer sets the backend server on the request URL before the proxy runs.
			handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				req.URL.Scheme = backendURL.Scheme
				req.URL.Host = backendURL.Host
				proxy.ServeHTTP(rw, req)
			})

			serverProtocols := new(http.Protocols)
			serverProtocols.SetHTTP1(true)
			serverProtocols.SetUnencryptedHTTP2(true)

			srv := httptest.NewUnstartedServer(handler)
			srv.Config.Protocols = serverProtocols
			srv.Start()
			t.Cleanup(srv.Close)

			clientProtocols := new(http.Protocols)
			if test.http2 {
				clientProtocols.SetUnencryptedHTTP2(true)
			} else {
				clientProtocols.SetHTTP1(true)
			}
			transport := &http.Transport{Protocols: clientProtocols}
			t.Cleanup(transport.CloseIdleConnections)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodConnect, srv.URL, strings.NewReader("foo"))
			require.NoError(t, err)

			res, err := transport.RoundTrip(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = res.Body.Close() })

			if test.http2 {
				assert.Equal(t, 2, res.ProtoMajor)
			} else {
				assert.Equal(t, 1, res.ProtoMajor)
			}
			assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
			assert.False(t, backendCalled.Load())
		})
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
