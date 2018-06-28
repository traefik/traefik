package tls

import (
	"crypto/tls"
	"fmt"
	"strings"
	"testing"

	"github.com/containous/traefik/safe"
	"github.com/stretchr/testify/assert"
)

func TestGetBestCertificate(t *testing.T) {
	testCases := []struct {
		desc          string
		domainToCheck string
		staticCert    string
		dynamicCert   string
		expectedCert  string
	}{
		{
			desc:          "Empty Store, returns no certs",
			domainToCheck: "snitest.com",
			staticCert:    "",
			dynamicCert:   "",
			expectedCert:  "",
		},
		{
			desc:          "Empty static cert store",
			domainToCheck: "snitest.com",
			staticCert:    "",
			dynamicCert:   "snitest.com",
			expectedCert:  "snitest.com",
		},
		{
			desc:          "Empty dynamic cert store",
			domainToCheck: "snitest.com",
			staticCert:    "snitest.com",
			dynamicCert:   "",
			expectedCert:  "snitest.com",
		},
		{
			desc:          "Best Match",
			domainToCheck: "snitest.com",
			staticCert:    "snitest.com",
			dynamicCert:   "snitest.org",
			expectedCert:  "snitest.com",
		},
		{
			desc:          "Best Match with wildcard dynamic and exact static",
			domainToCheck: "www.snitest.com",
			staticCert:    "www.snitest.com",
			dynamicCert:   "*.snitest.com",
			expectedCert:  "www.snitest.com",
		},
		{
			desc:          "Best Match with wildcard static and exact dynamic",
			domainToCheck: "www.snitest.com",
			staticCert:    "*.snitest.com",
			dynamicCert:   "www.snitest.com",
			expectedCert:  "www.snitest.com",
		},
		{
			desc:          "Best Match with static wildcard only",
			domainToCheck: "www.snitest.com",
			staticCert:    "*.snitest.com",
			dynamicCert:   "",
			expectedCert:  "*.snitest.com",
		},
		{
			desc:          "Best Match with dynamic wildcard only",
			domainToCheck: "www.snitest.com",
			staticCert:    "",
			dynamicCert:   "*.snitest.com",
			expectedCert:  "*.snitest.com",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			staticMap := map[string]*tls.Certificate{}
			dynamicMap := map[string]*tls.Certificate{}

			if test.staticCert != "" {
				staticCert, _ := loadTestCert(test.staticCert)
				staticMap[test.staticCert] = staticCert
			}
			if test.dynamicCert != "" {
				dynamicCert, _ := loadTestCert(test.dynamicCert)
				dynamicMap[test.dynamicCert] = dynamicCert
			}
			store := CertificateStore{
				DynamicCerts: safe.New(dynamicMap),
				StaticCerts:  safe.New(staticMap),
			}
			var expected *tls.Certificate
			if test.expectedCert != "" {
				cert, _ := loadTestCert(test.expectedCert)
				expected = cert
			}
			actual := store.GetBestCertificate(test.domainToCheck)
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
