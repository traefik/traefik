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
