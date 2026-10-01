package acme

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeHTTPServeHTTP(t *testing.T) {
	testCases := []struct {
		desc           string
		domain         string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			desc:           "IPv6 without port",
			domain:         "2001:db8::1",
			url:            "http://[2001:db8::1]/.well-known/acme-challenge/token",
			expectedStatus: http.StatusOK,
			expectedBody:   "keyAuth",
		},
		{
			desc:           "IPv6 with port",
			domain:         "2001:db8::1",
			url:            "http://[2001:db8::1]:80/.well-known/acme-challenge/token",
			expectedStatus: http.StatusOK,
			expectedBody:   "keyAuth",
		},
		{
			desc:           "DNS without port",
			domain:         "example.com",
			url:            "http://example.com/.well-known/acme-challenge/token",
			expectedStatus: http.StatusOK,
			expectedBody:   "keyAuth",
		},
		{
			desc:           "DNS with port",
			domain:         "example.com",
			url:            "http://example.com:80/.well-known/acme-challenge/token",
			expectedStatus: http.StatusOK,
			expectedBody:   "keyAuth",
		},
		{
			desc:           "IPv4 with port",
			domain:         "192.0.2.1",
			url:            "http://192.0.2.1:80/.well-known/acme-challenge/token",
			expectedStatus: http.StatusOK,
			expectedBody:   "keyAuth",
		},
		{
			desc:           "unknown host",
			domain:         "2001:db8::1",
			url:            "http://[2001:db8::2]/.well-known/acme-challenge/token",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			desc:           "unknown token",
			domain:         "2001:db8::1",
			url:            "http://[2001:db8::1]/.well-known/acme-challenge/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			challenge := NewChallengeHTTP()
			require.NoError(t, challenge.Present(t.Context(), test.domain, "token", "keyAuth"))

			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			rw := httptest.NewRecorder()
			challenge.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatus, rw.Code)
			assert.Equal(t, test.expectedBody, rw.Body.String())
		})
	}
}
