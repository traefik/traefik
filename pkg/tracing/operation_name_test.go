package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateOperationName(t *testing.T) {
	testCases := []struct {
		desc      string
		prefix    string
		parts     []string
		sep       string
		spanLimit int
		expected  string
	}{
		{
			desc:     "empty",
			expected: " ",
		},
		{
			desc:      "with prefix, without parts",
			prefix:    "foo",
			parts:     []string{},
			sep:       "-",
			spanLimit: 0,
			expected:  "foo ",
		},
		{
			desc:      "with prefix, without parts, too small span limit",
			prefix:    "foo",
			parts:     []string{},
			sep:       "-",
			spanLimit: 1,
			expected:  "foo 6c2d2c76",
		},
		{
			desc:      "with prefix, with parts",
			prefix:    "foo",
			parts:     []string{"fii", "fuu", "fee", "faa"},
			sep:       "-",
			spanLimit: 0,
			expected:  "foo fii-fuu-fee-faa",
		},
		{
			desc:      "with prefix, with parts, with span limit",
			prefix:    "foo",
			parts:     []string{"fff", "ooo", "ooo", "bbb", "aaa", "rrr"},
			sep:       "-",
			spanLimit: 20,
			expected:  "foo fff-ooo-ooo-bbb-aaa-rrr-1a8e8ac1",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			opName := generateOperationName(test.prefix, test.parts, test.sep, test.spanLimit)
			assert.Equal(t, test.expected, opName)
		})
	}
}

func TestComputeHash(t *testing.T) {
	testCases := []struct {
		desc     string
		text     string
		expected string
	}{
		{
			desc:     "hashing",
			text:     "some very long pice of text",
			expected: "0258ea1c",
		},
		{
			desc:     "short text less than limit 10",
			text:     "short",
			expected: "f9b0078b",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := computeHash(test.text)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestTruncateString(t *testing.T) {
	testCases := []struct {
		desc     string
		text     string
		limit    int
		expected string
	}{
		{
			desc:     "short text less than limit 10",
			text:     "short",
			limit:    10,
			expected: "short",
		},
		{
			desc:     "basic truncate with limit 10",
			text:     "some very long pice of text",
			limit:    10,
			expected: "some ve...",
		},
		{
			desc:     "truncate long FQDN to 39 chars",
			text:     "some-service-100.slug.namespace.environment.domain.tld",
			limit:    39,
			expected: "some-service-100.slug.namespace.envi...",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := truncateString(test.text, test.limit)

			assert.Equal(t, test.expected, actual)
			assert.True(t, len(actual) <= test.limit)
		})
	}
}
