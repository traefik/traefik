package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
