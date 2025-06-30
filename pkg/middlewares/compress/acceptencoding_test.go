package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
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
			desc:           "Empty Accept-Encoding",
			acceptEncoding: []string{""},
			expected:       identityName,
		},
		{
			desc:           "gzip > br (no weight)",
			acceptEncoding: []string{"gzip, br"},
			expected:       gzipName,
		},
		{
			desc:           "gzip > br > zstd (no weight)",
			acceptEncoding: []string{"gzip, br, zstd"},
			expected:       gzipName,
		},
		{
			desc:           "known compression encoding (no weight)",
			acceptEncoding: []string{"compress, gzip"},
			expected:       gzipName,
		},
		{
			desc:           "unknown compression encoding (no weight), no encoding",
			acceptEncoding: []string{"compress, rar"},
			expected:       notAcceptable,
		},
		{
			desc:           "wildcard returns the default compression encoding",
			acceptEncoding: []string{"*"},
			expected:       gzipName,
		},
		{
			desc:            "wildcard returns the custom default compression encoding",
			acceptEncoding:  []string{"*"},
			defaultEncoding: brotliName,
			expected:        brotliName,
		},
		{
			desc:           "follows weight",
			acceptEncoding: []string{"br;q=0.8, gzip;q=1.0, *;q=0.1"},
			expected:       gzipName,
		},
		{
			desc:           "identity with higher weight is preferred",
			acceptEncoding: []string{"br;q=0.8, identity;q=1.0"},
			expected:       identityName,
		},
		{
			desc:           "identity with equal weight is not preferred",
			acceptEncoding: []string{"br;q=0.8, identity;q=0.8"},
			expected:       brotliName,
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
		{
			desc:               "mixed weight",
			acceptEncoding:     []string{"gzip, br;q=0.9"},
			supportedEncodings: []string{gzipName, brotliName},
			expected:           gzipName,
		},
		{
			desc:           "Zero weights, no compression",
			acceptEncoding: []string{"br;q=0, gzip;q=0, zstd;q=0"},
			expected:       notAcceptable,
		},
		{
			desc:            "Zero weights, default encoding, no compression",
			acceptEncoding:  []string{"br;q=0, gzip;q=0, zstd;q=0"},
			defaultEncoding: "br",
			expected:        notAcceptable,
		},
		{
			desc:           "Same weight, first supported encoding",
			acceptEncoding: []string{"br;q=1.0, gzip;q=1.0, zstd;q=1.0"},
			expected:       gzipName,
		},
		{
			desc:           "Same weight, first supported encoding, order has no effect",
			acceptEncoding: []string{"br;q=1.0, zstd;q=1.0, gzip;q=1.0"},
			expected:       gzipName,
		},
		{
			desc:            "Same weight, first supported encoding, defaultEncoding has no effect",
			acceptEncoding:  []string{"br;q=1.0, zstd;q=1.0, gzip;q=1.0"},
			defaultEncoding: "br",
			expected:        gzipName,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.supportedEncodings == nil {
				test.supportedEncodings = defaultSupportedEncodings
			}

			conf := dynamic.Compress{
				Encodings:       test.supportedEncodings,
				DefaultEncoding: test.defaultEncoding,
			}

			h, err := New(t.Context(), nil, conf, "test")
			require.NoError(t, err)

			c, ok := h.(*compress)
			require.True(t, ok)

			encoding := c.getCompressionEncoding(test.acceptEncoding)

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
				{Type: brotliName, Weight: 1},
				{Type: zstdName, Weight: 0.9},
				{Type: gzipName, Weight: 0.8},
				{Type: wildcardName, Weight: 0.1},
			},
			assertWeight: assert.True,
		},
		{
			desc:               "weight with supported encodings",
			values:             []string{"br;q=1.0, zstd;q=0.9, gzip;q=0.8, *;q=0.1"},
			supportedEncodings: []string{brotliName, gzipName},
			expected: []Encoding{
				{Type: brotliName, Weight: 1},
				{Type: gzipName, Weight: 0.8},
				{Type: wildcardName, Weight: 0.1},
			},
			assertWeight: assert.True,
		},
		{
			desc:   "mixed",
			values: []string{"zstd,gzip, br;q=1.0, *;q=0"},
			expected: []Encoding{
				{Type: zstdName, Weight: 1},
				{Type: gzipName, Weight: 1},
				{Type: brotliName, Weight: 1},
			},
			assertWeight: assert.True,
		},
		{
			desc:               "mixed with supported encodings",
			values:             []string{"zstd,gzip, br;q=1.0, *;q=0"},
			supportedEncodings: []string{zstdName},
			expected: []Encoding{
				{Type: zstdName, Weight: 1},
			},
			assertWeight: assert.True,
		},
		{
			desc:   "no weight",
			values: []string{"zstd, gzip, br, *"},
			expected: []Encoding{
				{Type: zstdName, Weight: 1},
				{Type: gzipName, Weight: 1},
				{Type: brotliName, Weight: 1},
				{Type: wildcardName, Weight: 1},
			},
			assertWeight: assert.False,
		},
		{
			desc:               "no weight with supported encodings",
			values:             []string{"zstd, gzip, br, *"},
			supportedEncodings: []string{"gzip"},
			expected: []Encoding{
				{Type: gzipName, Weight: 1},
				{Type: wildcardName, Weight: 1},
			},
			assertWeight: assert.False,
		},
		{
			desc:   "weight and identity",
			values: []string{"gzip;q=1.0, identity; q=0.5, *;q=0"},
			expected: []Encoding{
				{Type: gzipName, Weight: 1},
				{Type: identityName, Weight: 0.5},
			},
			assertWeight: assert.True,
		},
		{
			desc:               "weight and identity",
			values:             []string{"gzip;q=1.0, identity; q=0.5, *;q=0"},
			supportedEncodings: []string{"br"},
			expected: []Encoding{
				{Type: identityName, Weight: 0.5},
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

			supportedEncodings := buildSupportedEncodings(test.supportedEncodings)

			aes := parseAcceptableEncodings(test.values, supportedEncodings)

			assert.Equal(t, test.expected, aes)
		})
	}
}
