package snicheck

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/traefik/traefik/v3/pkg/tcp"
)

func TestSNICheck(t *testing.T) {
	testCases := []struct {
		name           string
		tlsState       *tls.ConnectionState
		contextOptions string
		configured     string
		wantStatus     int
		wantNext       bool
	}{
		{
			name:       "no TLS state passes through",
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
		{
			name:           "matching TLS options pass through",
			tlsState:       &tls.ConnectionState{ServerName: "example.com"},
			contextOptions: "default",
			configured:     "default",
			wantStatus:     http.StatusOK,
			wantNext:       true,
		},
		{
			name:           "mismatched TLS options return 421",
			tlsState:       &tls.ConnectionState{ServerName: "example.com"},
			contextOptions: "default",
			configured:     "mtls",
			wantStatus:     http.StatusMisdirectedRequest,
		},
		{
			name:       "empty context options return 421 for a configured policy",
			tlsState:   &tls.ConnectionState{ServerName: "example.com"},
			configured: "mtls",
			wantStatus: http.StatusMisdirectedRequest,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				nextCalled = true
				rw.WriteHeader(http.StatusOK)
			})

			handler := New("router", test.configured, next)
			req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
			req.TLS = test.tlsState
			req = req.WithContext(tcp.AddTLSOptionsNameInContext(req.Context(), test.contextOptions))
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, test.wantNext, nextCalled)
			assert.Equal(t, test.wantStatus, recorder.Code)
		})
	}
}
