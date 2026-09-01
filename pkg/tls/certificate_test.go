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

	// Decodable PEM bodies, as the identifier is re-encoded from the decoded block.
	// The first one is longer than maxCertificateNameLen, so that truncation is observable.
	body := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("traefik!", 8)))
	shortBody := base64.StdEncoding.EncodeToString([]byte("short"))
	keyBlock := "-----BEGIN PRIVATE KEY-----\n" + base64.StdEncoding.EncodeToString([]byte("SUPERSECRETKEY")) + "\n-----END PRIVATE KEY-----\n"

	longMissingPath := "/etc/traefik/certificates/" + strings.Repeat("z", 40) + ".crt"

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
			desc:     "missing file path is returned as-is",
			certFile: "/ssl/does-not-exist.crt",
			expected: "/ssl/does-not-exist.crt",
		},
		{
			desc:     "missing file path longer than the limit is truncated",
			certFile: longMissingPath,
			expected: longMissingPath[:maxCertificateNameLen],
		},
		{
			desc:     "inlined certificate is truncated",
			certFile: "-----BEGIN CERTIFICATE-----\n" + body + "\n-----END CERTIFICATE-----\n",
			expected: body[:maxCertificateNameLen],
		},
		{
			desc:     "inlined certificate with CRLF line endings is truncated",
			certFile: "-----BEGIN CERTIFICATE-----\r\n" + body + "\r\n-----END CERTIFICATE-----\r\n",
			expected: body[:maxCertificateNameLen],
		},
		{
			desc:     "inlined certificate shorter than the limit is returned without its armor",
			certFile: "-----BEGIN CERTIFICATE-----\n" + shortBody + "\n-----END CERTIFICATE-----\n",
			expected: shortBody,
		},
		{
			desc:     "content preceding the certificate block is ignored",
			certFile: "Bag Attributes: friendlyName=example.com\n-----BEGIN CERTIFICATE-----\n" + body + "\n-----END CERTIFICATE-----\n",
			expected: body[:maxCertificateNameLen],
		},
		{
			desc:     "inlined bundle holding the key after the certificate is truncated",
			certFile: "-----BEGIN CERTIFICATE-----\n" + body + "\n-----END CERTIFICATE-----\n" + keyBlock,
			expected: body[:maxCertificateNameLen],
		},
		{
			desc:     "inlined bundle starting with the key is not echoed",
			certFile: keyBlock + "-----BEGIN CERTIFICATE-----\n" + body + "\n-----END CERTIFICATE-----\n",
			expected: inlinedCertificateName,
		},
		{
			desc:     "inlined PEM block which is not a certificate is not echoed",
			certFile: keyBlock,
			expected: inlinedCertificateName,
		},
		{
			desc:     "certificate block too short to decode is not echoed",
			certFile: "-----BEGIN CERTIFICATE-----\nABC\n-----BEGIN PRIVATE KEY-----\nSECRET\n",
			expected: inlinedCertificateName,
		},
		{
			desc:     "PEM block without its end marker is not echoed",
			certFile: "-----BEGIN PRIVATE KEY-----\nSUPERSECRETKEY",
			expected: inlinedCertificateName,
		},
		{
			desc:     "unarmored content longer than the limit is truncated",
			certFile: strings.Repeat("u", 80),
			expected: strings.Repeat("u", maxCertificateNameLen),
		},
		{
			desc:     "empty certificate",
			certFile: "",
			expected: "",
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
