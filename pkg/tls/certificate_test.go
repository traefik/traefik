package tls

import (
	"crypto/x509"
	"encoding/base64"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
)

func Test_verifyServerCertMatchesURI(t *testing.T) {
	tests := []struct {
		desc   string
		uri    string
		cert   *x509.Certificate
		expErr require.ErrorAssertionFunc
	}{
		{
			desc:   "returns error when certificate is nil",
			uri:    "spiffe://foo.com",
			expErr: require.Error,
		},
		{
			desc:   "returns error when certificate has no URIs",
			uri:    "spiffe://foo.com",
			cert:   &x509.Certificate{URIs: nil},
			expErr: require.Error,
		},
		{
			desc: "returns error when no URI matches",
			uri:  "spiffe://foo.com",
			cert: &x509.Certificate{URIs: []*url.URL{
				{Scheme: "spiffe", Host: "other.org"},
			}},
			expErr: require.Error,
		},
		{
			desc: "returns nil when URI matches",
			uri:  "spiffe://foo.com",
			cert: &x509.Certificate{URIs: []*url.URL{
				{Scheme: "spiffe", Host: "foo.com"},
			}},
			expErr: require.NoError,
		},
		{
			desc: "returns nil when one of the URI matches",
			uri:  "spiffe://foo.com",
			cert: &x509.Certificate{URIs: []*url.URL{
				{Scheme: "spiffe", Host: "example.org"},
				{Scheme: "spiffe", Host: "foo.com"},
			}},
			expErr: require.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := verifyServerCertMatchesURI(test.uri, test.cert)
			test.expErr(t, err)
		})
	}
}

func TestCertificate_GetTruncatedCertificateName(t *testing.T) {
	t.Parallel()

	existingPath := filepath.Join(t.TempDir(), "example.com.crt")
	require.NoError(t, os.WriteFile(existingPath, []byte("content"), 0o600))

	certBlock := "-----BEGIN CERTIFICATE-----\n" + base64.StdEncoding.EncodeToString([]byte(strings.Repeat("traefik!", 8))) + "\n-----END CERTIFICATE-----\n"
	keyBlock := "-----BEGIN PRIVATE KEY-----\n" + base64.StdEncoding.EncodeToString([]byte("SUPERSECRETKEY")) + "\n-----END PRIVATE KEY-----\n"

	testCases := []struct {
		desc     string
		certFile string
		expected string
	}{
		{
			desc:     "existing file path is returned as-is",
			certFile: existingPath,
			expected: existingPath,
		},
		{
			desc:     "inlined certificate keeps both PEM block types",
			certFile: certBlock,
			expected: "-----BEGIN CERTI[...]ERTIFICATE-----\n",
		},
		{
			desc:     "inlined bundle holding the key after the certificate shows the key block",
			certFile: certBlock + keyBlock,
			expected: "-----BEGIN CERTI[...]RIVATE KEY-----\n",
		},
		{
			desc:     "inlined bundle holding the key before the certificate shows the key block",
			certFile: keyBlock + certBlock,
			expected: "-----BEGIN PRIVA[...]ERTIFICATE-----\n",
		},
		{
			desc:     "inlined key configured as the certificate is recognizable",
			certFile: keyBlock,
			expected: "-----BEGIN PRIVA[...]RIVATE KEY-----\n",
		},
		{
			desc:     "unarmored content is excerpted at both ends",
			certFile: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("s", 32))),
			expected: "c3Nzc3Nzc3N[...]3Nzc3Nzc3M=",
		},
		{
			desc:     "missing file path keeps its directory and its file name",
			certFile: "/etc/traefik/certificates/example.com.crt",
			expected: "/etc/traef[...]le.com.crt",
		},
		{
			desc:     "missing file paths sharing a directory stay distinguishable",
			certFile: "/etc/traefik/certificates/wildcard.example.com.crt",
			expected: "/etc/traefik[...]mple.com.crt",
		},
		{
			desc:     "missing file path shorter than the limit is halved",
			certFile: "/ssl/x.crt",
			expected: "/s[...]rt",
		},
		{
			desc:     "content above the excerpt cap",
			certFile: strings.Repeat("a", 48) + strings.Repeat("z", 48),
			expected: "aaaaaaaaaaaaaaaa[...]zzzzzzzzzzzzzzzz",
		},
		{
			desc:     "content just above the smallest excerpt",
			certFile: "abcd",
			expected: "a[...]d",
		},
		{
			desc:     "content too short to be excerpted",
			certFile: "abc",
			expected: "[...]",
		},
		{
			desc:     "empty certificate",
			certFile: "",
			expected: "[...]",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			certificate := Certificate{CertFile: types.FileOrContent(test.certFile)}

			assert.Equal(t, test.expected, certificate.GetTruncatedCertificateName())
		})
	}
}
