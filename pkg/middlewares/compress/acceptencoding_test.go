package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getCompressionEncoding(t *testing.T) {
	testCases := []struct {
		desc               string
		acceptEncoding     []string
		defaultEncoding    string
		supportedEncodings []string
		expected           string
	}{
		{
			desc:           "br > gzip (no weight)",
			acceptEncoding: []string{"gzip, br"},
			expected:       brotliName,
		},
		{
			desc:           "zstd > br > gzip (no weight)",
			acceptEncoding: []string{"zstd, gzip, br"},
			expected:       zstdName,
		},
		{
			desc:           "known compression encoding (no weight)",
			acceptEncoding: []string{"compress, gzip"},
			expected:       gzipName,
		},
		{
			desc:           "unknown compression encoding (no weight), no encoding",
			acceptEncoding: []string{"compress, rar"},
			expected:       identityName,
		},
		{
			desc:           "wildcard return the default compression encoding",
			acceptEncoding: []string{"*"},
			expected:       brotliName,
		},
		{
			desc:            "wildcard return the custom default compression encoding",
			acceptEncoding:  []string{"*"},
			defaultEncoding: "foo",
			expected:        "foo",
		},
		{
			desc:           "follows weight",
			acceptEncoding: []string{"br;q=0.8, gzip;q=1.0, *;q=0.1"},
			expected:       gzipName,
		},
		{
			desc:           "ignore unknown compression encoding",
			acceptEncoding: []string{"compress;q=1.0, gzip;q=0.5"},
			expected:       gzipName,
		},
		{
			desc:           "fallback on non-zero compression encoding",
			acceptEncoding: []string{"compress;q=1.0, gzip, identity;q=0"},
			expected:       gzipName,
		},
		{
			desc:           "not acceptable (identity)",
			acceptEncoding: []string{"compress;q=1.0, identity;q=0"},
			expected:       notAcceptable,
		},
		{
			desc:           "not acceptable (wildcard)",
			acceptEncoding: []string{"compress;q=1.0, *;q=0"},
			expected:       notAcceptable,
		},
		{
			desc:           "non-zero is higher than 0",
			acceptEncoding: []string{"gzip, *;q=0"},
			expected:       gzipName,
		},
		{
			desc:               "zstd forbidden, brotli first",
			acceptEncoding:     []string{"zstd, gzip, br"},
			supportedEncodings: []string{brotliName, gzipName},
			expected:           brotliName,
		},
		{
			desc:               "follows weight, ignores forbidden encoding",
			acceptEncoding:     []string{"br;q=0.8, gzip;q=1.0, *;q=0.1"},
			supportedEncodings: []string{zstdName, brotliName},
			expected:           brotliName,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.supportedEncodings == nil {
				test.supportedEncodings = defaultSupportedEncodings
			}

			encoding := getCompressionEncoding(test.acceptEncoding, test.defaultEncoding, test.supportedEncodings)

			assert.Equal(t, test.expected, encoding)
		})
	}
}

func Test_parseAcceptEncoding(t *testing.T) {
	testCases := []struct {
		desc               string
		values             []string
		supportedEncodings []string
		expected           []Encoding
		assertWeight       assert.BoolAssertionFunc
	}{
		{
			desc:   "weight",
			values: []string{"br;q=1.0, zstd;q=0.9, gzip;q=0.8, *;q=0.1"},
			expected: []Encoding{
				{Type: brotliName, Weight: ptr[float64](1)},
				{Type: zstdName, Weight: ptr(0.9)},
				{Type: gzipName, Weight: ptr(0.8)},
				{Type: wildcardName, Weight: ptr(0.1)},
			},
			assertWeight: assert.True,
		},
		{
			desc:               "weight with supported encodings",
			values:             []string{"br;q=1.0, zstd;q=0.9, gzip;q=0.8, *;q=0.1"},
			supportedEncodings: []string{brotliName, gzipName},
			expected: []Encoding{
				{Type: brotliName, Weight: ptr[float64](1)},
				{Type: gzipName, Weight: ptr(0.8)},
				{Type: wildcardName, Weight: ptr(0.1)},
			},
			assertWeight: assert.True,
		},
		{
			desc:   "mixed",
			values: []string{"zstd,gzip, br;q=1.0, *;q=0"},
			expected: []Encoding{
				{Type: brotliName, Weight: ptr[float64](1)},
				{Type: zstdName},
				{Type: gzipName},
				{Type: wildcardName, Weight: ptr[float64](0)},
			},
			assertWeight: assert.True,
		},
		{
			desc:               "mixed with supported encodings",
			values:             []string{"zstd,gzip, br;q=1.0, *;q=0"},
			supportedEncodings: []string{zstdName},
			expected: []Encoding{
				{Type: zstdName},
				{Type: wildcardName, Weight: ptr[float64](0)},
			},
			assertWeight: assert.True,
		},
		{
			desc:   "no weight",
			values: []string{"zstd, gzip, br, *"},
			expected: []Encoding{
				{Type: zstdName},
				{Type: gzipName},
				{Type: brotliName},
				{Type: wildcardName},
			},
			assertWeight: assert.False,
		},
		{
			desc:               "no weight with supported encodings",
			values:             []string{"zstd, gzip, br, *"},
			supportedEncodings: []string{"gzip"},
			expected: []Encoding{
				{Type: gzipName},
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
		{
			desc:               "weight and identity",
			values:             []string{"gzip;q=1.0, identity; q=0.5, *;q=0"},
			supportedEncodings: []string{"br"},
			expected: []Encoding{
				{Type: identityName, Weight: ptr(0.5)},
				{Type: wildcardName, Weight: ptr[float64](0)},
			},
			assertWeight: assert.True,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.supportedEncodings == nil {
				test.supportedEncodings = defaultSupportedEncodings
			}

			aes, hasWeight := parseAcceptEncoding(test.values, test.supportedEncodings)

			assert.Equal(t, test.expected, aes)
			test.assertWeight(t, hasWeight)
		})
	}
}

func ptr[T any](t T) *T {
	return &t
}
