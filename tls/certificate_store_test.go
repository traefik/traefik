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
	testCases := []struct {
		desc          string
		domainToCheck string
		staticCert    string
		dynamicCert   string
		expectedCert  string
		uppercase     bool
	}{
		{
			desc:          "Empty Store, returns no certs",
			domainToCheck: "snitest.com",
			staticCert:    "",
			dynamicCert:   "",
			expectedCert:  "",
			uppercase:     false,
		},
		{
			desc:          "Empty static cert store",
			domainToCheck: "snitest.com",
			staticCert:    "",
			dynamicCert:   "snitest.com",
			expectedCert:  "snitest.com",
			uppercase:     false,
		},
		{
			desc:          "Empty dynamic cert store",
			domainToCheck: "snitest.com",
			staticCert:    "snitest.com",
			dynamicCert:   "",
			expectedCert:  "snitest.com",
			uppercase:     false,
		},
		{
			desc:          "Best Match",
			domainToCheck: "snitest.com",
			staticCert:    "snitest.com",
			dynamicCert:   "snitest.org",
			expectedCert:  "snitest.com",
			uppercase:     false,
		},
		{
			desc:          "Best Match with wildcard dynamic and exact static",
			domainToCheck: "www.snitest.com",
			staticCert:    "www.snitest.com",
			dynamicCert:   "*.snitest.com",
			expectedCert:  "www.snitest.com",
			uppercase:     false,
		},
		{
			desc:          "Best Match with wildcard static and exact dynamic",
			domainToCheck: "www.snitest.com",
			staticCert:    "*.snitest.com",
			dynamicCert:   "www.snitest.com",
			expectedCert:  "www.snitest.com",
			uppercase:     false,
		},
		{
			desc:          "Best Match with static wildcard only",
			domainToCheck: "www.snitest.com",
			staticCert:    "*.snitest.com",
			dynamicCert:   "",
			expectedCert:  "*.snitest.com",
			uppercase:     false,
		},
		{
			desc:          "Best Match with dynamic wildcard only",
			domainToCheck: "www.snitest.com",
			staticCert:    "",
			dynamicCert:   "*.snitest.com",
			expectedCert:  "*.snitest.com",
			uppercase:     false,
		},
		{
			desc:          "Best Match with two wildcard certs",
			domainToCheck: "foo.www.snitest.com",
			staticCert:    "*.www.snitest.com",
			dynamicCert:   "*.snitest.com",
			expectedCert:  "*.www.snitest.com",
			uppercase:     false,
		},
		{
			desc:          "Best Match with static wildcard only, case insensitive",
			domainToCheck: "bar.www.snitest.com",
			staticCert:    "*.www.snitest.com",
			dynamicCert:   "",
			expectedCert:  "*.www.snitest.com",
			uppercase:     true,
		},
		{
			desc:          "Best Match with dynamic wildcard only, case insensitive",
			domainToCheck: "bar.www.snitest.com",
			staticCert:    "",
			dynamicCert:   "*.www.snitest.com",
			expectedCert:  "*.www.snitest.com",
			uppercase:     true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			staticMap := map[string]*tls.Certificate{}
			dynamicMap := map[string]*tls.Certificate{}

			if test.staticCert != "" {
				cert, err := loadTestCert(test.staticCert, test.uppercase)
				require.NoError(t, err)
				staticMap[strings.ToLower(test.staticCert)] = cert
			}

			if test.dynamicCert != "" {
				cert, err := loadTestCert(test.dynamicCert, test.uppercase)
				require.NoError(t, err)
				dynamicMap[strings.ToLower(test.dynamicCert)] = cert
			}

			store := &CertificateStore{
				DynamicCerts: safe.New(dynamicMap),
				StaticCerts:  safe.New(staticMap),
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

func loadTestCert(certName string, uppercase bool) (*tls.Certificate, error) {
	replacement := "wildcard"
	if uppercase {
		replacement = "uppercase_wildcard"
	}

	staticCert, err := tls.LoadX509KeyPair(
		fmt.Sprintf("../integration/fixtures/https/%s.cert", strings.Replace(certName, "*", replacement, -1)),
		fmt.Sprintf("../integration/fixtures/https/%s.key", strings.Replace(certName, "*", replacement, -1)),
	)
	if err != nil {
		return nil, err
	}

	return &staticCert, nil
}
