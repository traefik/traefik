package tls

import (
	"crypto/tls"
	"fmt"
	"strings"
	"testing"
	"time"

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

// TestGetBestCertificate_Hostnames ensures a certificate restricted to hostnames is only served for matching server names,
// even if one of its SANs matches the server name (https://github.com/traefik/traefik/issues/12742).
func TestGetBestCertificate_Hostnames(t *testing.T) {
	wildcardCert := &CertificateData{
		Certificate: &tls.Certificate{},
		Hostnames:   []string{"*.example.test"},
	}
	// The identifier of this certificate sorts before the one of the wildcard certificate,
	// so without restriction it would be selected for any *.example.test server name.
	otherCert := &CertificateData{
		Certificate: &tls.Certificate{},
		Hostnames:   []string{"foobar.test"},
	}
	unrestrictedCert := &CertificateData{Certificate: &tls.Certificate{}}

	testCases := []struct {
		desc       string
		certs      map[string]*CertificateData
		serverName string
		expected   *CertificateData
	}{
		{
			desc: "restricted certificate is not served outside of its hostnames",
			certs: map[string]*CertificateData{
				"*.example.test":             wildcardCert,
				"*.example.test,foobar.test": otherCert,
			},
			serverName: "hello.example.test",
			expected:   wildcardCert,
		},
		{
			desc: "restricted certificate is served for its hostnames",
			certs: map[string]*CertificateData{
				"*.example.test":             wildcardCert,
				"*.example.test,foobar.test": otherCert,
			},
			serverName: "foobar.test",
			expected:   otherCert,
		},
		{
			desc: "no certificate when the only matching one is restricted to other hostnames",
			certs: map[string]*CertificateData{
				"*.example.test,foobar.test": otherCert,
			},
			serverName: "hello.example.test",
		},
		{
			desc: "certificate without hostnames is served for any matching server name",
			certs: map[string]*CertificateData{
				"*.example.test,foobar.test": unrestrictedCert,
			},
			serverName: "hello.example.test",
			expected:   unrestrictedCert,
		},
		{
			desc: "wildcard hostname matches several labels",
			certs: map[string]*CertificateData{
				"a.b.example.test": wildcardCert,
			},
			serverName: "a.b.example.test",
			expected:   wildcardCert,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			store := &CertificateStore{
				DynamicCerts: safe.New(test.certs),
				CertCache:    cache.New(1*time.Hour, 10*time.Minute),
			}

			actual := store.GetBestCertificate(&tls.ClientHelloInfo{ServerName: test.serverName})
			if test.expected == nil {
				assert.Nil(t, actual)
				return
			}

			assert.Same(t, test.expected.Certificate, actual)
		})
	}
}

func Test_matchHostname(t *testing.T) {
	testCases := []struct {
		desc       string
		hostname   string
		serverName string
		expected   bool
	}{
		{desc: "exact match", hostname: "foobar.test", serverName: "foobar.test", expected: true},
		{desc: "exact mismatch", hostname: "foobar.test", serverName: "hello.foobar.test"},
		{desc: "wildcard matches one label", hostname: "*.example.test", serverName: "hello.example.test", expected: true},
		{desc: "wildcard matches several labels", hostname: "*.example.test", serverName: "a.b.example.test", expected: true},
		{desc: "wildcard does not match the apex", hostname: "*.example.test", serverName: "example.test"},
		{desc: "wildcard does not match another domain", hostname: "*.example.test", serverName: "hello.example.org"},
		{desc: "wildcard does not match a partial label", hostname: "*.example.test", serverName: "helloexample.test"},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, matchHostname(test.hostname, test.serverName))
		})
	}
}

func Test_mergeHostnames(t *testing.T) {
	cert := &tls.Certificate{Certificate: [][]byte{[]byte("cert")}}
	otherCert := &tls.Certificate{Certificate: [][]byte{[]byte("other")}}

	testCases := []struct {
		desc     string
		existing *CertificateData
		cert     *CertificateData
		expected []string
	}{
		{
			desc:     "same certificate with other hostnames",
			existing: &CertificateData{Certificate: cert, Hostnames: []string{"b.test"}},
			cert:     &CertificateData{Certificate: cert, Hostnames: []string{"a.test", "b.test"}},
			expected: []string{"a.test", "b.test"},
		},
		{
			desc:     "same certificate without hostnames lifts the restriction",
			existing: &CertificateData{Certificate: cert, Hostnames: []string{"a.test"}},
			cert:     &CertificateData{Certificate: cert},
		},
		{
			desc:     "same certificate already without hostnames",
			existing: &CertificateData{Certificate: cert},
			cert:     &CertificateData{Certificate: cert, Hostnames: []string{"a.test"}},
		},
		{
			desc:     "different certificate keeps the existing one",
			existing: &CertificateData{Certificate: cert, Hostnames: []string{"a.test"}},
			cert:     &CertificateData{Certificate: otherCert, Hostnames: []string{"b.test"}},
			expected: []string{"a.test"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := mergeHostnames(test.existing, test.cert)

			assert.Same(t, cert, actual.Certificate)
			assert.Equal(t, test.expected, actual.Hostnames)
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
