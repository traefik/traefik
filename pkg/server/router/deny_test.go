package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_denyConnect(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		protoMajor int
		wantStatus int
	}{
		{
			name:       "Rejects HTTP/1 CONNECT",
			method:     http.MethodConnect,
			protoMajor: 1,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "Rejects HTTP/2 CONNECT",
			method:     http.MethodConnect,
			protoMajor: 2,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "Allows GET",
			method:     http.MethodGet,
			protoMajor: 1,
			wantStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := denyConnect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(test.method, "http://example.com", nil)
			req.ProtoMajor = test.protoMajor
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			assert.Equal(t, test.wantStatus, res.Code)
		})
	}
}

func Test_denyEncodedPathCharacters(t *testing.T) {
	tests := []struct {
		name       string
		encoded    map[string]struct{}
		url        string
		wantStatus int
	}{
		{
			name: "Rejects disallowed characters",
			encoded: map[string]struct{}{
				"%0A": {},
				"%0D": {},
			},
			url:        "http://example.com/foo%0Abar",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Allows valid paths",
			encoded: map[string]struct{}{
				"%0A": {},
				"%0D": {},
			},
			url:        "http://example.com/foo%20bar",
			wantStatus: http.StatusOK,
		},
		{
			name: "Handles empty path",
			encoded: map[string]struct{}{
				"%0A": {},
			},
			url:        "http://example.com/",
			wantStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := denyEncodedPathCharacters(test.encoded, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			assert.Equal(t, test.wantStatus, res.Code)
		})
	}
}
