package tls

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/tls/certificate"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	rsaCipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	}
	ecCipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	}
	ecSupportedCurves = []tls.CurveID{
		tls.CurveP384,
		tls.CurveP256,
	}
	ecSignatureSchemes = []tls.SignatureScheme{
		tls.ECDSAWithP384AndSHA384,
		tls.ECDSAWithP256AndSHA256,
	}
)

func TestGetBestCertificate(t *testing.T) {
	// FIXME Add tests for defaultCert
	testCases := []struct {
		cipherSuites     []uint16
		desc             string
		domainToCheck    string
		dynamicCert      string
		expectedCert     string
		signatureSchemes []tls.SignatureScheme
		supportedCurves  []tls.CurveID
		uppercase        bool
	}{
		{
			cipherSuites:  rsaCipherSuites,
			desc:          "Empty Store, returns no certs",
			domainToCheck: "snitest.com",
			dynamicCert:   "",
			expectedCert:  "",
		},
		{
			cipherSuites:  rsaCipherSuites,
			desc:          "Best Match with no corresponding",
			domainToCheck: "snitest.com",
			dynamicCert:   "snitest.org",
			expectedCert:  "",
		},
		{
			cipherSuites:  rsaCipherSuites,
			desc:          "Best Match",
			domainToCheck: "snitest.com",
			dynamicCert:   "snitest.com",
			expectedCert:  "snitest.com",
		},
		{
			cipherSuites:  rsaCipherSuites,
			desc:          "Best Match with dynamic wildcard",
			domainToCheck: "www.snitest.com",
			dynamicCert:   "*.snitest.com",
			expectedCert:  "*.snitest.com",
		},
		{
			cipherSuites:  rsaCipherSuites,
			desc:          "Best Match with dynamic wildcard only, case insensitive",
			domainToCheck: "bar.www.snitest.com",
			dynamicCert:   "*.www.snitest.com",
			expectedCert:  "*.www.snitest.com",
			uppercase:     true,
		},
		{
			cipherSuites:  rsaCipherSuites,
			desc:          "Best Match with non-EC compatible cipher suites and no corresponding non-EC",
			domainToCheck: "snitest.com",
			dynamicCert:   "snitest.com/ec",
			expectedCert:  "",
		},
		{
			cipherSuites:     ecCipherSuites,
			desc:             "Best Match with EC compatible cipher suites",
			domainToCheck:    "snitest.com",
			dynamicCert:      "snitest.com/ec",
			expectedCert:     "snitest.com/ec",
			signatureSchemes: ecSignatureSchemes,
			supportedCurves:  ecSupportedCurves,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			dynamicMap := map[certificateKey]*tls.Certificate{}

			if test.dynamicCert != "" {
				cert, err := loadTestCert(test.dynamicCert, test.uppercase)
				require.NoError(t, err)

				key := certificateKey{
					hostname: strings.Split(strings.ToLower(test.dynamicCert), "/")[0],
				}

				switch cert.PrivateKey.(type) {
				case *rsa.PrivateKey:
					key.certType = certificate.RSA
				case *ecdsa.PrivateKey, *ed25519.PrivateKey:
					key.certType = certificate.EC
				}

				dynamicMap[key] = cert
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
				ServerName:       test.domainToCheck,
				CipherSuites:     test.cipherSuites,
				SupportedCurves:  test.supportedCurves,
				SignatureSchemes: test.signatureSchemes,
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

	fileBasename := strings.Replace(certName, "*", replacement, -1)
	fileBasename = strings.Replace(fileBasename, certTypeDelimiter, "_type_", -1)

	staticCert, err := tls.LoadX509KeyPair(
		fmt.Sprintf("../../integration/fixtures/https/%s.cert", fileBasename),
		fmt.Sprintf("../../integration/fixtures/https/%s.key", fileBasename),
	)
	if err != nil {
		return nil, err
	}

	return &staticCert, nil
}
