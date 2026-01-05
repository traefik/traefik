package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_denyFragment(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "Rejects fragment character",
			url:        "http://example.com/#",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Allows without fragment",
			url:        "http://example.com/",
			wantStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := denyFragment(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, test.url, nil)
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
