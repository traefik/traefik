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
			desc:     "inlined certificate is truncated to its PEM block type",
			certFile: certBlock,
			expected: "-----BEGIN CERTI...",
		},
		{
			desc:     "inlined bundle is truncated before its key block",
			certFile: certBlock + keyBlock,
			expected: "-----BEGIN CERTI...",
		},
		{
			desc:     "inlined key configured as the certificate is truncated to its PEM block type",
			certFile: keyBlock,
			expected: "-----BEGIN PRIVA...",
		},
		{
			desc:     "unarmored content is truncated",
			certFile: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("s", 32))),
			expected: "c3Nzc3Nzc3Nzc3Nz...",
		},
		{
			desc:     "missing file path shorter than the limit is halved",
			certFile: "/ssl/x.crt",
			expected: "/ssl/...",
		},
		{
			desc:     "missing file path longer than the limit is truncated",
			certFile: "/etc/traefik/certificates/example.com.crt",
			expected: "/etc/traefik/cer...",
		},
		{
			desc:     "empty certificate",
			certFile: "",
			expected: "...",
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
