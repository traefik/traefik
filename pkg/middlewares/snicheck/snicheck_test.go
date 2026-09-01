package snicheck

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

func TestSNICheck(t *testing.T) {
	testCases := []struct {
		desc               string
		routerTLSOptions   string
		connTLSOptionsName string
		setConnContext     bool
		nonTLSRequest      bool
		expectedStatusCode int
		expectedConnHeader string
	}{
		{
			desc:               "matching options",
			routerTLSOptions:   "default",
			connTLSOptionsName: "default",
			setConnContext:     true,
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "non-TLS request",
			routerTLSOptions:   "tls-strict@file",
			nonTLSRequest:      true,
			expectedStatusCode: http.StatusOK,
		},
		{
			// The connection's TLS options name was resolved at handshake time (and baked
			// into its context) against a dynamic configuration that has since changed:
			// the router now expects different TLS options than what the still-open
			// connection negotiated under. The middleware must reject the request as it
			// does today, but it must also mark the connection for closure so the client
			// is forced to re-handshake instead of getting stuck retrying the same stale
			// connection forever.
			desc:               "stale connection options",
			routerTLSOptions:   "tls-strict@file",
			connTLSOptionsName: "default",
			setConnContext:     true,
			expectedStatusCode: http.StatusMisdirectedRequest,
			expectedConnHeader: "close",
		},
		{
			desc:               "missing connection context",
			routerTLSOptions:   "tls-strict@file",
			setConnContext:     false,
			expectedStatusCode: http.StatusMisdirectedRequest,
			expectedConnHeader: "close",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := New("router-name", test.routerTLSOptions, next)

			var req *http.Request
			if test.nonTLSRequest {
				req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			} else {
				req = httptest.NewRequest(http.MethodGet, "https://example.com", nil)
				req.TLS = &tls.ConnectionState{ServerName: "example.com"}
				if test.setConnContext {
					req = req.WithContext(tcp.AddTLSOptionsNameInContext(req.Context(), test.connTLSOptionsName))
				} else {
					req = req.WithContext(context.Background())
				}
			}

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)
			assert.Equal(t, test.expectedConnHeader, rw.Header().Get("Connection"))
		})
	}
}
