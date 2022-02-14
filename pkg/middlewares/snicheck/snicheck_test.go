package snicheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSNICheck_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc              string
		tlsOptionsForHost map[string]string
		host              string
		expected          int
	}{
		{
			desc:     "no TLS options",
			expected: http.StatusOK,
		},
		{
			desc: "with TLS options",
			tlsOptionsForHost: map[string]string{
				"example.com": "foo",
			},
			expected: http.StatusOK,
		},
		{
			desc: "server name and host doesn't have the same TLS configuration",
			tlsOptionsForHost: map[string]string{
				"example.com": "foo",
			},
			host:     "example.com",
			expected: http.StatusMisdirectedRequest,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

			sniCheck := New(test.tlsOptionsForHost, next)

			req := httptest.NewRequest(http.MethodGet, "https://localhost", nil)
			if test.host != "" {
				req.Host = test.host
			}

			recorder := httptest.NewRecorder()

			sniCheck.ServeHTTP(recorder, req)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}
