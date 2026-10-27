package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
)

func Test_VerifyPeerCertificate(t *testing.T) {
	pki := newTestPKI(t)

	tests := []struct {
		desc     string
		sans     []SAN
		rawCerts [][]byte
		rootCAs  *x509.CertPool
		expErr   require.ErrorAssertionFunc
	}{
		{
			desc:     "returns error when no certificates are provided",
			sans:     []SAN{{Type: SANURIType, Value: "spiffe://foo.com"}},
			rawCerts: nil,
			rootCAs:  pki.caPool,
			expErr:   require.Error,
		},
		{
			desc:     "returns error when certificate has no URIs",
			sans:     []SAN{{Type: SANURIType, Value: "spiffe://foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, nil, nil)},
			rootCAs:  pki.caPool,
			expErr:   require.Error,
		},
		{
			desc: "returns error when no URI matches",
			sans: []SAN{{Type: SANURIType, Value: "spiffe://foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, nil, []*url.URL{
				{Scheme: "spiffe", Host: "other.org"},
			})},
			rootCAs: pki.caPool,
			expErr:  require.Error,
		},
		{
			desc: "returns nil when URI matches",
			sans: []SAN{{Type: SANURIType, Value: "spiffe://foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, nil, []*url.URL{
				{Scheme: "spiffe", Host: "foo.com"},
			})},
			rootCAs: pki.caPool,
			expErr:  require.NoError,
		},
		{
			desc: "returns nil when one of the URIs matches",
			sans: []SAN{{Type: SANURIType, Value: "spiffe://foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, nil, []*url.URL{
				{Scheme: "spiffe", Host: "example.org"},
				{Scheme: "spiffe", Host: "foo.com"},
			})},
			rootCAs: pki.caPool,
			expErr:  require.NoError,
		},
		{
			desc:     "returns error when certificate has no DNS names",
			sans:     []SAN{{Type: SANDNSNameType, Value: "foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, nil, nil)},
			rootCAs:  pki.caPool,
			expErr:   require.Error,
		},
		{
			desc:     "returns error when no DNS name matches",
			sans:     []SAN{{Type: SANDNSNameType, Value: "foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, []string{"other.com"}, nil)},
			rootCAs:  pki.caPool,
			expErr:   require.Error,
		},
		{
			desc:     "returns nil when DNS name matches",
			sans:     []SAN{{Type: SANDNSNameType, Value: "foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, []string{"foo.com"}, nil)},
			rootCAs:  pki.caPool,
			expErr:   require.NoError,
		},
		{
			desc:     "returns nil when DNS name matches a wildcard",
			sans:     []SAN{{Type: SANDNSNameType, Value: "bar.foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, []string{"*.foo.com"}, nil)},
			rootCAs:  pki.caPool,
			expErr:   require.NoError,
		},
		{
			desc:     "returns nil when one of the DNS names matches",
			sans:     []SAN{{Type: SANDNSNameType, Value: "foo.com"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, []string{"example.com", "foo.com"}, nil)},
			rootCAs:  pki.caPool,
			expErr:   require.NoError,
		},
		{
			desc:     "returns nil when DNS name matches case-insensitively",
			sans:     []SAN{{Type: SANDNSNameType, Value: "FOO.COM"}},
			rawCerts: [][]byte{pki.newLeafCertDER(t, []string{"foo.com"}, nil)},
			rootCAs:  pki.caPool,
			expErr:   require.NoError,
		},
		{
			desc: "returns nil when URI matches in mixed sans",
			sans: []SAN{
				{Type: SANURIType, Value: "spiffe://foo.com"},
				{Type: SANDNSNameType, Value: "foo.com"},
			},
			rawCerts: [][]byte{pki.newLeafCertDER(t, nil, []*url.URL{
				{Scheme: "spiffe", Host: "foo.com"},
			})},
			rootCAs: pki.caPool,
			expErr:  require.NoError,
		},
		{
			desc: "returns nil when DNS name matches in mixed sans",
			sans: []SAN{
				{Type: SANURIType, Value: "spiffe://foo.com"},
				{Type: SANDNSNameType, Value: "foo.com"},
			},
			rawCerts: [][]byte{pki.newLeafCertDER(t, []string{"foo.com"}, nil)},
			rootCAs:  pki.caPool,
			expErr:   require.NoError,
		},
		{
			desc: "returns error when neither URI nor DNS name matches in mixed sans",
			sans: []SAN{
				{Type: SANURIType, Value: "spiffe://foo.com"},
				{Type: SANDNSNameType, Value: "foo.com"},
			},
			rawCerts: [][]byte{pki.newLeafCertDER(t, []string{"other.com"}, []*url.URL{
				{Scheme: "spiffe", Host: "other.org"},
			})},
			rootCAs: pki.caPool,
			expErr:  require.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := VerifyPeerCertificate(test.sans, test.rootCAs, test.rawCerts)
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

type testPKI struct {
	caPool *x509.CertPool
	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
}

func newTestPKI(t *testing.T) *testPKI {
	t.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Test CA"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caCert, err := x509.ParseCertificate(caCertDER)
	require.NoError(t, err)

	pool := x509.NewCertPool()
	pool.AddCert(caCert)

	return &testPKI{caPool: pool, caCert: caCert, caKey: caKey}
}

func (p *testPKI) newLeafCertDER(t *testing.T, dnsNames []string, uris []*url.URL) []byte {
	t.Helper()

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{Organization: []string{"Test Leaf"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     dnsNames,
		URIs:         uris,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, p.caCert, &leafKey.PublicKey, p.caKey)
	require.NoError(t, err)

	return certDER
}
