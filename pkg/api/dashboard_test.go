package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_safePrefix(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected string
	}{
		{
			desc:     "host",
			value:    "https://example.com",
			expected: "",
		},
		{
			desc:     "host with path",
			value:    "https://example.com/foo/bar?test",
			expected: "",
		},
		{
			desc:     "path",
			value:    "/foo/bar",
			expected: "/foo/bar",
		},
		{
			desc:     "path without leading slash",
			value:    "foo/bar",
			expected: "foo/bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
			require.NoError(t, err)

			req.Header.Set("X-Forwarded-Prefix", test.value)

			prefix := safePrefix(req)

			assert.Equal(t, test.expected, prefix)
		})
	}
}
