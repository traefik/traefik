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
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			desc:           "known host and token",
			url:            "http://[2001:db8::1]/.well-known/acme-challenge/token",
			expectedStatus: http.StatusOK,
			expectedBody:   "keyAuth",
		},
		{
			desc:           "unknown host",
			url:            "http://[2001:db8::2]/.well-known/acme-challenge/token",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			desc:           "unknown host with port",
			url:            "http://[2001:db8::2]:80/.well-known/acme-challenge/token",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			desc:           "unknown token",
			url:            "http://[2001:db8::1]/.well-known/acme-challenge/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			desc:           "unknown token with port",
			url:            "http://[2001:db8::1]:80/.well-known/acme-challenge/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
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

			assert.Equal(t, test.expectedStatus, rw.Code)
			assert.Equal(t, test.expectedBody, rw.Body.String())
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

func TestChallengeHTTPCleanUpTokenIsolation(t *testing.T) {
	tokenCases := []struct {
		removedToken     string
		remainingToken   string
		remainingKeyAuth string
	}{
		{
			removedToken:     "first",
			remainingToken:   "second",
			remainingKeyAuth: "secondKeyAuth",
		},
		{
			removedToken:     "second",
			remainingToken:   "first",
			remainingKeyAuth: "firstKeyAuth",
		},
	}

	for _, tokenCase := range tokenCases {
		t.Run(tokenCase.removedToken, func(t *testing.T) {
			t.Parallel()

			newChallenge := func(t *testing.T) *ChallengeHTTP {
				t.Helper()

				challenge := NewChallengeHTTP()
				require.NoError(t, challenge.Present(t.Context(), "2001:db8::1", "first", "firstKeyAuth"))
				require.NoError(t, challenge.Present(t.Context(), "2001:db8::1", "second", "secondKeyAuth"))
				return challenge
			}
			assertTokenResponse := func(t *testing.T, challenge *ChallengeHTTP, token string, status int, body string) {
				t.Helper()

				req := httptest.NewRequest(http.MethodGet, "http://[2001:db8::1]/.well-known/acme-challenge/"+token, nil)
				rw := httptest.NewRecorder()
				challenge.ServeHTTP(rw, req)

				assert.Equal(t, status, rw.Code, "token %s", token)
				assert.Equal(t, body, rw.Body.String(), "token %s", token)
			}
			checkCleanupAndReuse := func(t *testing.T, challenge *ChallengeHTTP) {
				t.Helper()

				for _, token := range []string{"first", "second"} {
					require.NoError(t, challenge.CleanUp(t.Context(), "2001:db8::1", token, ""))
					assertTokenResponse(t, challenge, token, http.StatusNotFound, "")
				}

				require.NoError(t, challenge.Present(t.Context(), "2001:db8::1", tokenCase.removedToken, "newKeyAuth"))
				assertTokenResponse(t, challenge, tokenCase.removedToken, http.StatusOK, "newKeyAuth")
			}

			// Unknown cleanup targets must leave both tokens available.
			{
				challenge := newChallenge(t)
				require.NoError(t, challenge.CleanUp(t.Context(), "2001:db8::2", tokenCase.removedToken, ""))
				assertTokenResponse(t, challenge, "first", http.StatusOK, "firstKeyAuth")
				assertTokenResponse(t, challenge, "second", http.StatusOK, "secondKeyAuth")

				require.NoError(t, challenge.CleanUp(t.Context(), "2001:db8::1", "unknown", ""))
				assertTokenResponse(t, challenge, "first", http.StatusOK, "firstKeyAuth")
				assertTokenResponse(t, challenge, "second", http.StatusOK, "secondKeyAuth")
				checkCleanupAndReuse(t, challenge)
			}

			// Repeated cleanup must preserve the other token and allow reuse.
			{
				challenge := newChallenge(t)
				require.NoError(t, challenge.CleanUp(t.Context(), "2001:db8::1", tokenCase.removedToken, ""))
				assertTokenResponse(t, challenge, tokenCase.removedToken, http.StatusNotFound, "")
				assertTokenResponse(t, challenge, tokenCase.remainingToken, http.StatusOK, tokenCase.remainingKeyAuth)

				require.NoError(t, challenge.CleanUp(t.Context(), "2001:db8::1", tokenCase.removedToken, ""))
				assertTokenResponse(t, challenge, tokenCase.removedToken, http.StatusNotFound, "")
				assertTokenResponse(t, challenge, tokenCase.remainingToken, http.StatusOK, tokenCase.remainingKeyAuth)
				checkCleanupAndReuse(t, challenge)
			}
		})
	}
}
