package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getCompressionType(t *testing.T) {
	testCases := []struct {
		desc     string
		values   []string
		expected string
	}{
		{
			desc:     "br > gzip (no weight)",
			values:   []string{"gzip, br"},
			expected: brotliName,
		},
		{
			desc:     "unknown compression type (no weight)",
			values:   []string{"compress, gzip"},
			expected: gzipName,
		},
		{
			desc:     "unknown compression types (no weight) use default",
			values:   []string{"compress, rar"},
			expected: "foo",
		},
		{
			desc:     "wildcard return the default compression type",
			values:   []string{"*"},
			expected: "foo",
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			encodingType := getCompressionType(test.values, "foo")

			assert.Equal(t, test.expected, encodingType)
		})
	}
}

func Test_parseEncodingAccepts(t *testing.T) {
	testCases := []struct {
		desc         string
		values       []string
		expected     []Encoding
		assertWeight assert.BoolAssertionFunc
	}{
		{
			desc:         "weight",
			values:       []string{"br;q=1.0, gzip;q=0.8, *;q=0.1"},
			expected:     []Encoding{{Type: "br", Weight: ptr[float64](1)}, {Type: "gzip", Weight: ptr(0.8)}, {Type: "*", Weight: ptr(0.1)}},
			assertWeight: assert.True,
		},
		{
			desc:         "no weight",
			values:       []string{"gzip, br, *"},
			expected:     []Encoding{{Type: "gzip"}, {Type: "br"}, {Type: "*"}},
			assertWeight: assert.False,
		},
		{
			desc:         "weight and identity",
			values:       []string{"gzip;q=1.0, identity; q=0.5, *;q=0"},
			expected:     []Encoding{{Type: "gzip", Weight: ptr[float64](1)}, {Type: "identity", Weight: ptr(0.5)}, {Type: "*", Weight: ptr[float64](0)}},
			assertWeight: assert.True,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			aes, hasWeight := parseAcceptsEncoding(test.values)

			assert.Equal(t, test.expected, aes)
			test.assertWeight(t, hasWeight)
		})
	}
}

func ptr[T any](t T) *T {
	return &t
}
