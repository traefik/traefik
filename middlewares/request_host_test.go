package middlewares

import (
	"net/http"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestRequestHost(t *testing.T) {
	testCases := []struct {
		desc     string
		url      string
		expected string
	}{
		{
			desc:     "host without :",
			url:      "http://host",
			expected: "host",
		},
		{
			desc:     "host with : and without port",
			url:      "http://host:",
			expected: "host",
		},
		{
			desc:     "IP host with : and with port",
			url:      "http://127.0.0.1:123",
			expected: "127.0.0.1",
		},
		{
			desc:     "IP host with : and without port",
			url:      "http://127.0.0.1:",
			expected: "127.0.0.1",
		},
	}

	rh := &RequestHost{}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := testhelpers.MustNewRequest(http.MethodGet, test.url, nil)

			rh.ServeHTTP(nil, req, func(_ http.ResponseWriter, r *http.Request) {
				host := GetCanonizedHost(r.Context())
				assert.Equal(t, test.expected, host)
			})
		})
	}
}

func TestRequestHostParseHost(t *testing.T) {
	testCases := []struct {
		desc     string
		host     string
		expected string
	}{
		{
			desc:     "host without :",
			host:     "host",
			expected: "host",
		},
		{
			desc:     "host with : and without port",
			host:     "host:",
			expected: "host",
		},
		{
			desc:     "IP host with : and with port",
			host:     "127.0.0.1:123",
			expected: "127.0.0.1",
		},
		{
			desc:     "IP host with : and without port",
			host:     "127.0.0.1:",
			expected: "127.0.0.1",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := parseHost(test.host)

			assert.Equal(t, test.expected, actual)
		})
	}
}
