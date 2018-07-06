package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
