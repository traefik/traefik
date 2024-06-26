package tracing

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func Test_safeFullURL(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "URL with password",
			input:    "https://user:password123@example.com",
			expected: "https://user:REDACTED@example.com",
		},
		{
			desc:     "URL with sensitive query parameters",
			input:    "https://example.com?password=secret&token=abcdef&api_key=12345&name=John",
			expected: "https://example.com?api_key=REDACTED&name=John&password=REDACTED&token=REDACTED",
		},
		{
			desc:     "URL with sensitive path",
			input:    "https://example.com/api/secret/12345/token/67890",
			expected: "https://example.com/api/REDACTED/12345/REDACTED/67890",
		},
		{
			desc:     "URL with multiple sensitive elements",
			input:    "https://user:pass@example.com/api/secret/12345?token=abcdef&name=John",
			expected: "https://user:REDACTED@example.com/api/REDACTED/12345?name=John&token=REDACTED",
		},
		{
			desc:     "URL without sensitive data",
			input:    "https://example.com/api/users/12345?name=John",
			expected: "https://example.com/api/users/12345?name=John",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			inputURL, err := url.Parse(test.input)
			require.NoError(t, err)

			result := safeURL(inputURL)

			assert.Equal(t, test.expected, result.String())
			require.Equal(t, inputURL.String(), test.input)
		})
	}
}
