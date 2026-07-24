package tls

import (
	"crypto/tls"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/safe"
)

func TestGetBestCertificate(t *testing.T) {
	// TODO Add tests for defaultCert
	testCases := []struct {
		desc          string
		domainToCheck string
		dynamicCert   string
		expectedCert  string
		uppercase     bool
	}{
		{
			desc:          "Empty Store, returns no certs",
			domainToCheck: "snitest.com",
			dynamicCert:   "",
			expectedCert:  "",
		},
		{
			desc:          "Best Match with no corresponding",
			domainToCheck: "snitest.com",
			dynamicCert:   "snitest.org",
			expectedCert:  "",
		},
		{
			desc:          "Best Match",
			domainToCheck: "snitest.com",
			dynamicCert:   "snitest.com",
			expectedCert:  "snitest.com",
		},
		{
			desc:          "Best Match with dynamic wildcard",
			domainToCheck: "www.snitest.com",
			dynamicCert:   "*.snitest.com",
			expectedCert:  "*.snitest.com",
		},
		{
			desc:          "Best Match with dynamic wildcard only, case-insensitive",
			domainToCheck: "bar.www.snitest.com",
			dynamicCert:   "*.www.snitest.com",
			expectedCert:  "*.www.snitest.com",
			uppercase:     true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			dynamicMap := map[string]*CertificateData{}

			if test.dynamicCert != "" {
				cert, err := loadTestCert(test.dynamicCert, test.uppercase)
				require.NoError(t, err)
				dynamicMap[strings.ToLower(test.dynamicCert)] = &CertificateData{Certificate: cert}
			}

			store := &CertificateStore{
				DynamicCerts: safe.New(dynamicMap),
				CertCache:    cache.New(1*time.Hour, 10*time.Minute),
			}

			var expected *tls.Certificate
			if test.expectedCert != "" {
				cert, err := loadTestCert(test.expectedCert, test.uppercase)
				require.NoError(t, err)
				expected = cert
			}

			clientHello := &tls.ClientHelloInfo{
				ServerName: test.domainToCheck,
			}

			actual := store.GetBestCertificate(clientHello)
			assert.Equal(t, expected, actual)
		})
	}
}

// TestGetBestCertificate_SharedSAN ensures the selection stays deterministic
// when distinct certificates share a SAN matching the server name (https://github.com/traefik/traefik/issues/13286).
func TestGetBestCertificate_SharedSAN(t *testing.T) {
	wildcardCert := &CertificateData{Certificate: &tls.Certificate{}}
	exactCert := &CertificateData{Certificate: &tls.Certificate{}}

	// Both certificates have a SAN matching app.example.test, but the exact-only
	// certificate must always win as its identifier sorts last.
	for range 100 {
		dynamicMap := map[string]*CertificateData{
			"*.app.example.test,app.example.test": wildcardCert,
			"app.example.test":                    exactCert,
		}

		store := &CertificateStore{
			DynamicCerts: safe.New(dynamicMap),
			CertCache:    cache.New(1*time.Hour, 10*time.Minute),
		}

		clientHello := &tls.ClientHelloInfo{ServerName: "app.example.test"}
		assert.Same(t, exactCert.Certificate, store.GetBestCertificate(clientHello))
	}
}

func TestGetBestCertificateIPReverseAddress(t *testing.T) {
	cert := &tls.Certificate{}

	dynamicMap := map[string]*CertificateData{
		"2001:db8::1": {Certificate: cert},
	}

	store := &CertificateStore{
		DynamicCerts: safe.New(dynamicMap),
		CertCache:    cache.New(1*time.Hour, 10*time.Minute),
	}

	reverseAddr, err := dns.ReverseAddr("2001:db8::1")
	require.NoError(t, err)

	clientHello := &tls.ClientHelloInfo{ServerName: reverseAddr}
	assert.Same(t, cert, store.GetBestCertificate(clientHello))
}

func TestMatchDomainIPReverseAddress(t *testing.T) {
	testCases := []struct {
		desc       string
		serverIP   string
		certDomain string
		expected   bool
	}{
		{
			desc:       "IPv4 reverse address",
			serverIP:   "192.0.2.1",
			certDomain: "192.0.2.1",
			expected:   true,
		},
		{
			desc:       "IPv6 reverse address",
			serverIP:   "2001:db8::1",
			certDomain: "2001:db8::1",
			expected:   true,
		},
		{
			desc:       "different IPv6 reverse address",
			serverIP:   "2001:db8::1",
			certDomain: "2001:db8::2",
			expected:   false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			reverseAddr, err := dns.ReverseAddr(test.serverIP)
			require.NoError(t, err)

			assert.Equal(t, test.expected, matchDomain(reverseAddr, test.certDomain))
		})
	}
}

func loadTestCert(certName string, uppercase bool) (*tls.Certificate, error) {
	replacement := "wildcard"
	if uppercase {
		replacement = "uppercase_wildcard"
	}

	staticCert, err := tls.LoadX509KeyPair(
		fmt.Sprintf("../../integration/fixtures/https/%s.cert", strings.ReplaceAll(certName, "*", replacement)),
		fmt.Sprintf("../../integration/fixtures/https/%s.key", strings.ReplaceAll(certName, "*", replacement)),
	)
	if err != nil {
		return nil, err
	}

	return &staticCert, nil
}
