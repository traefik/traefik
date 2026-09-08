package service

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartRoundTripper(t *testing.T) {
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprint(rw, req.Proto)
	})

	backend := httptest.NewUnstartedServer(handler)
	backend.Config.Protocols = new(http.Protocols)
	backend.Config.Protocols.SetHTTP1(true)
	backend.Config.Protocols.SetUnencryptedHTTP2(true)
	backend.Start()
	t.Cleanup(backend.Close)

	tlsBackend := httptest.NewUnstartedServer(handler)
	tlsBackend.EnableHTTP2 = true
	tlsBackend.StartTLS()
	t.Cleanup(tlsBackend.Close)

	testCases := []struct {
		desc          string
		scheme        string
		upgrade       bool
		expectedProto string
	}{
		{
			desc:          "h2c uses HTTP/2 with prior knowledge",
			scheme:        "h2c",
			expectedProto: "HTTP/2.0",
		},
		{
			desc:          "h2c with connection upgrade falls back to HTTP/1.1",
			scheme:        "h2c",
			upgrade:       true,
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "http uses HTTP/1.1",
			scheme:        "http",
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "http with connection upgrade uses HTTP/1.1",
			scheme:        "http",
			upgrade:       true,
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "https uses HTTP/2 negotiated with TLS ALPN",
			scheme:        "https",
			expectedProto: "HTTP/2.0",
		},
		{
			desc:          "https with connection upgrade falls back to HTTP/1.1",
			scheme:        "https",
			upgrade:       true,
			expectedProto: "HTTP/1.1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rt := newSmartRoundTripper(&http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			})

			targetURL := backend.URL
			switch test.scheme {
			case "https":
				targetURL = tlsBackend.URL
			case "h2c":
				targetURL = strings.Replace(targetURL, "http://", "h2c://", 1)
			}

			proto := doProtoRequest(t, rt, targetURL, test.upgrade)
			assert.Equal(t, test.expectedProto, proto)

			// The kerberos round tripper relies on Clone, which must preserve the protocol switching.
			proto = doProtoRequest(t, rt.Clone(), targetURL, test.upgrade)
			assert.Equal(t, test.expectedProto, proto)
		})
	}
}

func doProtoRequest(t *testing.T, rt http.RoundTripper, targetURL string, upgrade bool) string {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, targetURL, http.NoBody)
	require.NoError(t, err)

	if upgrade {
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
	}

	res, err := rt.RoundTrip(req)
	require.NoError(t, err)

	t.Cleanup(func() { _ = res.Body.Close() })

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return string(body)
}
