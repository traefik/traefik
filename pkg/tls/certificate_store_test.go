package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/safe"
)

func TestGetBestCertificate(t *testing.T) {
	// FIXME Add tests for defaultCert
	testCases := []struct {
		desc          string
		domainToCheck string
		dynamicCerts  []string
		expectedCert  string
		uppercase     bool
		rsaSuitesOnly bool
	}{
		{
			desc:          "Empty Store, returns no certs",
			domainToCheck: "snitest.com",
			dynamicCerts:  nil,
			expectedCert:  "",
		},
		{
			desc:          "Best Match with no corresponding",
			domainToCheck: "snitest.com",
			dynamicCerts:  []string{"snitest.org"},
			expectedCert:  "",
		},
		{
			desc:          "Best Match",
			domainToCheck: "snitest.com",
			dynamicCerts:  []string{"snitest.com"},
			expectedCert:  "snitest.com",
		},
		{
			desc:          "Best Match with dynamic wildcard",
			domainToCheck: "www.snitest.com",
			dynamicCerts:  []string{"*.snitest.com"},
			expectedCert:  "*.snitest.com",
		},
		{
			desc:          "Best Match with dynamic wildcard only, case insensitive",
			domainToCheck: "bar.www.snitest.com",
			dynamicCerts:  []string{"*.www.snitest.com"},
			expectedCert:  "*.www.snitest.com",
			uppercase:     true,
		},
		{
			desc:          "Best Match with both RSA and ECDSA, when client supports RSA only",
			domainToCheck: "www.snitest.com",
			dynamicCerts:  []string{"ecdsa.snitest.com_www.snitest.com", "www.snitest.com"},
			expectedCert:  "www.snitest.com",
			rsaSuitesOnly: true,
		},
		{
			desc:          "Best Match RSA only",
			domainToCheck: "snitest.com",
			dynamicCerts:  []string{"snitest.com"},
			expectedCert:  "snitest.com",
			rsaSuitesOnly: true,
		},
		{
			desc:          "Best Match with both RSA and ECDSA",
			domainToCheck: "www.snitest.com",
			dynamicCerts:  []string{"www.snitest.com", "alt.snitest.com_www.snitest.com", "ecdsa.snitest.com_www.snitest.com"},
			expectedCert:  "ecdsa.snitest.com_www.snitest.com",
		},
		{
			desc:          "Best Match with ECDSA, when client supports RSA only",
			domainToCheck: "www.snitest.com",
			dynamicCerts:  []string{"ecdsa.snitest.com_www.snitest.com"},
			expectedCert:  "",
			rsaSuitesOnly: true,
		},
		{
			desc:          "Best Match with SAN",
			domainToCheck: "alt.snitest.com",
			dynamicCerts:  []string{"www.snitest.com", "alt.snitest.com_www.snitest.com", "ecdsa.snitest.com_www.snitest.com"},
			expectedCert:  "alt.snitest.com_www.snitest.com",
		},
	}

	var allCipherSuites, rsaCipherSuites []uint16
	for _, suite := range tls.CipherSuites() {
		allCipherSuites = append(allCipherSuites, suite.ID)
		if strings.Contains(suite.Name, "TLS_RSA_") {
			rsaCipherSuites = append(rsaCipherSuites, suite.ID)
		}
	}

	allSignatureSchemes := []tls.SignatureScheme{
		tls.PKCS1WithSHA256, tls.PKCS1WithSHA384, tls.PKCS1WithSHA512,
		tls.PSSWithSHA256, tls.PSSWithSHA384, tls.PSSWithSHA512,
		tls.ECDSAWithP256AndSHA256, tls.ECDSAWithP384AndSHA384, tls.ECDSAWithP521AndSHA512,
		tls.Ed25519,
		tls.PKCS1WithSHA1, tls.ECDSAWithSHA1,
	}

	rsaSignatureSchemes := []tls.SignatureScheme{
		tls.PKCS1WithSHA256, tls.PKCS1WithSHA384, tls.PKCS1WithSHA512,
		tls.PSSWithSHA256, tls.PSSWithSHA384, tls.PSSWithSHA512,
		tls.PKCS1WithSHA1,
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			dynamicMap := map[string][]*tls.Certificate{}

			for _, certName := range test.dynamicCerts {
				cert, err := loadTestCert(certName, test.uppercase)
				require.NoError(t, err)
				key := strings.ReplaceAll(strings.ToLower(certName), "_", ",")
				if _, exists := dynamicMap[key]; !exists {
					dynamicMap[key] = []*tls.Certificate{}
				}

				dynamicMap[key] = append(dynamicMap[key], cert)
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
				ServerName:        test.domainToCheck,
				CipherSuites:      allCipherSuites,
				SignatureSchemes:  allSignatureSchemes,
				SupportedVersions: []uint16{tls.VersionTLS13, tls.VersionTLS12, tls.VersionTLS11, tls.VersionTLS10},
			}

			if test.rsaSuitesOnly {
				clientHello.CipherSuites = rsaCipherSuites
				clientHello.SignatureSchemes = rsaSignatureSchemes
			}

			actual := store.GetBestCertificate(clientHello)
			assert.Equal(t, expected, actual)
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

	staticCert.Leaf, err = x509.ParseCertificate(staticCert.Certificate[0])
	if err != nil {
		return nil, err
	}

	return &staticCert, nil
}
