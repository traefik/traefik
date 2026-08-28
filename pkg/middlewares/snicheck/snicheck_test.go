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

func TestSNICheck_matchingOptions(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := New("router-name", "default", next)

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	req.TLS = &tls.ConnectionState{ServerName: "example.com"}
	req = req.WithContext(tcp.AddTLSOptionsNameInContext(req.Context(), "default"))

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Empty(t, rw.Header().Get("Connection"))
}

func TestSNICheck_nonTLSRequest(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := New("router-name", "tls-strict@file", next)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Empty(t, rw.Header().Get("Connection"))
}

// TestSNICheck_staleConnectionOptions covers the scenario where a connection's TLS
// options name was resolved at handshake time (and baked into its context) against a
// dynamic configuration that has since changed: the router now expects different TLS
// options than what the still-open connection negotiated under. The middleware must
// reject the request as it does today, but it must also mark the connection for
// closure so the client is forced to re-handshake instead of getting stuck retrying
// the same stale connection forever.
func TestSNICheck_staleConnectionOptions(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := New("router-name", "tls-strict@file", next)

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	req.TLS = &tls.ConnectionState{ServerName: "example.com"}
	req = req.WithContext(tcp.AddTLSOptionsNameInContext(req.Context(), "default"))

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusMisdirectedRequest, rw.Code)
	assert.Equal(t, "close", rw.Header().Get("Connection"))
}

func TestSNICheck_missingConnectionContext(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := New("router-name", "tls-strict@file", next)

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	req.TLS = &tls.ConnectionState{ServerName: "example.com"}
	req = req.WithContext(context.Background())

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusMisdirectedRequest, rw.Code)
	assert.Equal(t, "close", rw.Header().Get("Connection"))
}
