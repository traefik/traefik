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
			static_map := map[string]*tls.Certificate{}
			dynamic_map := map[string]*tls.Certificate{}

			if test.staticCert != "" {
				static_cert, _ := tls.LoadX509KeyPair(
					fmt.Sprintf("../integration/fixtures/https/%s.cert", strings.Replace(test.staticCert, "*", "wildcard", -1)),
					fmt.Sprintf("../integration/fixtures/https/%s.key", strings.Replace(test.staticCert, "*", "wildcard", -1)),
				)
				static_map[test.staticCert] = &static_cert
			}
			if test.dynamicCert != "" {
				dynamic_cert, _ := tls.LoadX509KeyPair(
					fmt.Sprintf("../integration/fixtures/https/%s.cert", strings.Replace(test.dynamicCert, "*", "wildcard", -1)),
					fmt.Sprintf("../integration/fixtures/https/%s.key", strings.Replace(test.dynamicCert, "*", "wildcard", -1)),
				)
				dynamic_map[test.dynamicCert] = &dynamic_cert
			}
			store := CertificateStore{
				DynamicCerts: safe.New(dynamic_map),
				StaticCerts:  safe.New(static_map),
			}
			var expected *tls.Certificate
			if test.expectedCert != "" {
				cert, _ := tls.LoadX509KeyPair(
					fmt.Sprintf("../integration/fixtures/https/%s.cert", strings.Replace(test.expectedCert, "*", "wildcard", -1)),
					fmt.Sprintf("../integration/fixtures/https/%s.key", strings.Replace(test.expectedCert, "*", "wildcard", -1)),
				)
				expected = &cert
			}
			actual := store.GetBestCertificate(test.domainToCheck)
			assert.Equal(t, expected, actual)
		})
	}
}
