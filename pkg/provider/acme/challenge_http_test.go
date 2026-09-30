package acme

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeHTTPServeHTTP(t *testing.T) {
	challenge := NewChallengeHTTP()
	require.NoError(t, challenge.Present(t.Context(), "2001:db8::1", "token", "keyAuth"))

	req := httptest.NewRequest(http.MethodGet, "http://[2001:db8::1]/.well-known/acme-challenge/token", nil)
	rw := httptest.NewRecorder()

	challenge.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "keyAuth", rw.Body.String())
}

func TestChallengeHTTPServeHTTPNotFound(t *testing.T) {
	testCases := []struct {
		desc string
		url  string
	}{
		{
			desc: "unknown host",
			url:  "http://[2001:db8::2]/.well-known/acme-challenge/token",
		},
		{
			desc: "unknown host with port",
			url:  "http://[2001:db8::2]:80/.well-known/acme-challenge/token",
		},
		{
			desc: "unknown token",
			url:  "http://[2001:db8::1]/.well-known/acme-challenge/unknown",
		},
		{
			desc: "unknown token with port",
			url:  "http://[2001:db8::1]:80/.well-known/acme-challenge/unknown",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			challenge := NewChallengeHTTP()
			require.NoError(t, challenge.Present(t.Context(), "2001:db8::1", "token", "keyAuth"))

			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			rw := httptest.NewRecorder()
			challenge.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusNotFound, rw.Code)
			assert.Empty(t, rw.Body.String())
		})
	}
}

func TestChallengeHTTPCleanUpHostIsolation(t *testing.T) {
	testCases := []struct {
		desc   string
		domain string
	}{
		{
			desc:   "remove first host",
			domain: "2001:db8::1",
		},
		{
			desc:   "remove second host",
			domain: "2001:db8::2",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			challenge := NewChallengeHTTP()
			hosts := []struct {
				domain  string
				keyAuth string
			}{
				{domain: "2001:db8::1", keyAuth: "firstKeyAuth"},
				{domain: "2001:db8::2", keyAuth: "secondKeyAuth"},
			}
			for _, host := range hosts {
				require.NoError(t, challenge.Present(t.Context(), host.domain, "token", host.keyAuth))
			}

			for _, host := range hosts {
				req := httptest.NewRequest(http.MethodGet, "http://["+host.domain+"]/.well-known/acme-challenge/token", nil)
				rw := httptest.NewRecorder()
				challenge.ServeHTTP(rw, req)

				require.Equal(t, http.StatusOK, rw.Code)
				require.Equal(t, host.keyAuth, rw.Body.String())
			}

			require.NoError(t, challenge.CleanUp(t.Context(), test.domain, "token", ""))

			for _, host := range hosts {
				req := httptest.NewRequest(http.MethodGet, "http://["+host.domain+"]/.well-known/acme-challenge/token", nil)
				rw := httptest.NewRecorder()
				challenge.ServeHTTP(rw, req)

				expectedStatus := http.StatusOK
				expectedBody := host.keyAuth
				if host.domain == test.domain {
					expectedStatus = http.StatusNotFound
					expectedBody = ""
				}
				assert.Equal(t, expectedStatus, rw.Code)
				assert.Equal(t, expectedBody, rw.Body.String())
			}
		})
	}
}

func TestParseHTTPChallengeHost(t *testing.T) {
	testCases := []struct {
		desc     string
		host     string
		expected string
	}{
		{
			desc:     "host without port",
			host:     "example.com",
			expected: "example.com",
		},
		{
			desc:     "host with port",
			host:     "example.com:80",
			expected: "example.com",
		},
		{
			desc:     "IPv4 with port",
			host:     "127.0.0.1:80",
			expected: "127.0.0.1",
		},
		{
			desc:     "IPv6 without brackets",
			host:     "2001:db8::1",
			expected: "2001:db8::1",
		},
		{
			desc:     "IPv6 with brackets",
			host:     "[2001:db8::1]",
			expected: "2001:db8::1",
		},
		{
			desc:     "IPv6 with brackets and port",
			host:     "[2001:db8::1]:80",
			expected: "2001:db8::1",
		},
		{
			desc:     "empty address",
			host:     "",
			expected: "",
		},
		{
			desc:     "only colon",
			host:     ":",
			expected: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := parseHTTPChallengeHost(test.host)
			assert.Equal(t, test.expected, actual)
		})
	}
}
