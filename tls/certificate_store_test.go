package tls

import (
	"crypto/tls"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/containous/traefik/safe"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBestCertificate(t *testing.T) {
	// FIXME Add tests for defaultCert
	testCases := []struct {
		desc          string
		domainToCheck string
		dynamicCert   string
		expectedCert  string
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
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			dynamicMap := map[string]*tls.Certificate{}

			if test.dynamicCert != "" {
				cert, err := loadTestCert(test.dynamicCert)
				require.NoError(t, err)
				dynamicMap[test.dynamicCert] = cert
			}

			store := &CertificateStore{
				DynamicCerts: safe.New(dynamicMap),
				CertCache:    cache.New(1*time.Hour, 10*time.Minute),
			}

			var expected *tls.Certificate
			if test.expectedCert != "" {
				cert, err := loadTestCert(test.expectedCert)
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

func loadTestCert(certName string) (*tls.Certificate, error) {
	staticCert, err := tls.LoadX509KeyPair(
		fmt.Sprintf("../integration/fixtures/https/%s.cert", strings.Replace(certName, "*", "wildcard", -1)),
		fmt.Sprintf("../integration/fixtures/https/%s.key", strings.Replace(certName, "*", "wildcard", -1)),
	)
	if err != nil {
		return nil, err
	}

	return &staticCert, nil
}
