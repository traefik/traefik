package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getCompressionType(t *testing.T) {
	testCases := []struct {
		desc        string
		values      []string
		defaultType string
		expected    string
	}{
		{
			desc:     "br > gzip (no weight)",
			values:   []string{"gzip, br"},
			expected: brotliName,
		},
		{
			desc:     "known compression type (no weight)",
			values:   []string{"compress, gzip"},
			expected: gzipName,
		},
		{
			desc:     "unknown compression type (no weight), no encoding",
			values:   []string{"compress, rar"},
			expected: identityName,
		},
		{
			desc:     "wildcard return the default compression type",
			values:   []string{"*"},
			expected: brotliName,
		},
		{
			desc:        "wildcard return the custom default compression type",
			values:      []string{"*"},
			defaultType: "foo",
			expected:    "foo",
		},
		{
			desc:     "follows weight",
			values:   []string{"br;q=0.8, gzip;q=1.0, *;q=0.1"},
			expected: gzipName,
		},
		{
			desc:     "ignore unknown compression type",
			values:   []string{"compress;q=1.0, gzip;q=0.5"},
			expected: gzipName,
		},
		{
			desc:     "not acceptable (identity)",
			values:   []string{"compress;q=1.0, identity;q=0"},
			expected: notAcceptable,
		},
		{
			desc:     "not acceptable (wildcard)",
			values:   []string{"compress;q=1.0, *;q=0"},
			expected: notAcceptable,
		},
		{
			desc:     "non-zero is higher than 0",
			values:   []string{"gzip, *;q=0"},
			expected: gzipName,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			encodingType := getCompressionType(test.values, test.defaultType)

			assert.Equal(t, test.expected, encodingType)
		})
	}
}

func Test_parseAcceptEncoding(t *testing.T) {
	testCases := []struct {
		desc         string
		values       []string
		expected     []Encoding
		assertWeight assert.BoolAssertionFunc
	}{
		{
			desc:   "weight",
			values: []string{"br;q=1.0, gzip;q=0.8, *;q=0.1"},
			expected: []Encoding{
				{Type: brotliName, Weight: ptr[float64](1)},
				{Type: gzipName, Weight: ptr(0.8)},
				{Type: wildcardName, Weight: ptr(0.1)},
			},
			assertWeight: assert.True,
		},
		{
			desc:   "mixed",
			values: []string{"gzip, br;q=1.0, *;q=0"},
			expected: []Encoding{
				{Type: brotliName, Weight: ptr[float64](1)},
				{Type: gzipName},
				{Type: wildcardName, Weight: ptr[float64](0)},
			},
			assertWeight: assert.True,
		},
		{
			desc:   "no weight",
			values: []string{"gzip, br, *"},
			expected: []Encoding{
				{Type: gzipName},
				{Type: brotliName},
				{Type: wildcardName},
			},
			assertWeight: assert.False,
		},
		{
			desc:   "weight and identity",
			values: []string{"gzip;q=1.0, identity; q=0.5, *;q=0"},
			expected: []Encoding{
				{Type: gzipName, Weight: ptr[float64](1)},
				{Type: identityName, Weight: ptr(0.5)},
				{Type: wildcardName, Weight: ptr[float64](0)},
			},
			assertWeight: assert.True,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			aes, hasWeight := parseAcceptEncoding(test.values)

			assert.Equal(t, test.expected, aes)
			test.assertWeight(t, hasWeight)
		})
	}
}

func ptr[T any](t T) *T {
	return &t
}
