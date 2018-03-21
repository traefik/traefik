package acme

import (
	"crypto/tls"
	"testing"

	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestAcme_getUncheckedCertificates(t *testing.T) {

	wildcardMap := make(map[string]*tls.Certificate)
	wildcardMap["*.traefik.wtf"] = &tls.Certificate{}

	wildcardSafe := &safe.Safe{}
	wildcardSafe.Set(wildcardMap)

	domainMap := make(map[string]*tls.Certificate)
	domainMap["traefik.wtf"] = &tls.Certificate{}

	domainSafe := &safe.Safe{}
	domainSafe.Set(domainMap)

	tests := []struct {
		desc             string
		dynamicCerts     *safe.Safe
		staticCerts      map[string]*tls.Certificate
		acmeCertificates []*Certificate
		domains          []string
		expectedDomains  []string
	}{
		{
			desc:            "wildcard to generate",
			domains:         []string{"*.traefik.wtf"},
			expectedDomains: []string{"*.traefik.wtf"},
		},
		{
			desc:            "wildcard already exists in dynamic certificates",
			domains:         []string{"*.traefik.wtf"},
			dynamicCerts:    wildcardSafe,
			expectedDomains: nil,
		},
		{
			desc:            "wildcard already exists in static certificates",
			domains:         []string{"*.traefik.wtf"},
			staticCerts:     wildcardMap,
			expectedDomains: nil,
		},
		{
			desc:    "wildcard already exists in ACME certificates",
			domains: []string{"*.traefik.wtf"},
			acmeCertificates: []*Certificate{
				{
					Domain: types.Domain{Main: "*.traefik.wtf"},
				},
			},
			expectedDomains: nil,
		},
		{
			desc:            "domain CN and SANs to generate",
			domains:         []string{"traefik.wtf", "foo.traefik.wtf"},
			expectedDomains: []string{"traefik.wtf", "foo.traefik.wtf"},
		},
		{
			desc:            "domain CN already exists in dynamic certificates and SANs to generate",
			domains:         []string{"traefik.wtf", "foo.traefik.wtf"},
			dynamicCerts:    domainSafe,
			expectedDomains: []string{"foo.traefik.wtf"},
		},
		{
			desc:            "domain CN already exists in static certificates and SANs to generate",
			domains:         []string{"traefik.wtf", "foo.traefik.wtf"},
			staticCerts:     domainMap,
			expectedDomains: []string{"foo.traefik.wtf"},
		},
		{
			desc:    "domain CN already exists in ACME certificates and SANs to generate",
			domains: []string{"traefik.wtf", "foo.traefik.wtf"},
			acmeCertificates: []*Certificate{
				{
					Domain: types.Domain{Main: "traefik.wtf"},
				},
			},
			expectedDomains: []string{"foo.traefik.wtf"},
		},
		{
			desc:            "domain already exists in dynamic certificates",
			domains:         []string{"traefik.wtf"},
			dynamicCerts:    domainSafe,
			expectedDomains: nil,
		},
		{
			desc:            "domain already exists in static certificates",
			domains:         []string{"traefik.wtf"},
			staticCerts:     domainMap,
			expectedDomains: nil,
		},
		{
			desc:    "domain already exists in ACME certificates",
			domains: []string{"traefik.wtf"},
			acmeCertificates: []*Certificate{
				{
					Domain: types.Domain{Main: "traefik.wtf"},
				},
			},
			expectedDomains: nil,
		},
		{
			desc:            "domain matched by wildcard in dynamic certificates",
			domains:         []string{"traefik.wtf", "foo.traefik.wtf"},
			dynamicCerts:    wildcardSafe,
			expectedDomains: nil,
		},
		{
			desc:            "domain matched by wildcard in static certificates",
			domains:         []string{"traefik.wtf", "foo.traefik.wtf"},
			staticCerts:     wildcardMap,
			expectedDomains: nil,
		},
		{
			desc:    "domain matched by wildcard in ACME certificates",
			domains: []string{"traefik.wtf", "foo.traefik.wtf"},
			acmeCertificates: []*Certificate{
				{
					Domain: types.Domain{Main: "*.traefik.wtf"},
				},
			},
			expectedDomains: nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			acmeProvider := Provider{
				dynamicCerts: test.dynamicCerts,
				staticCerts:  test.staticCerts,
				certificates: test.acmeCertificates,
			}

			domains := acmeProvider.getUncheckedDomains(test.domains)
			assert.Equal(t, len(test.expectedDomains), len(domains), "Unexpected domains.")
		})
	}
}

func TestAcme_getValidDomain(t *testing.T) {
	tests := []struct {
		desc            string
		domains         types.Domain
		wildcardAllowed bool
		dnsChallenge    *DNSChallenge
		expectedErr     string
		expectedDomains []string
	}{
		{
			desc:            "valid wildcard",
			domains:         types.Domain{Main: "*.traefik.wtf"},
			dnsChallenge:    &DNSChallenge{},
			wildcardAllowed: true,
			expectedErr:     "",
			expectedDomains: []string{"*.traefik.wtf"},
		},
		{
			desc:            "no wildcard",
			domains:         types.Domain{Main: "traefik.wtf", SANs: []string{"foo.traefik.wtf"}},
			dnsChallenge:    &DNSChallenge{},
			expectedErr:     "",
			wildcardAllowed: true,
			expectedDomains: []string{"traefik.wtf", "foo.traefik.wtf"},
		},
		{
			desc:            "unauthorized wildcard",
			domains:         types.Domain{Main: "*.traefik.wtf"},
			dnsChallenge:    &DNSChallenge{},
			wildcardAllowed: false,
			expectedErr:     "unable to generate a wildcard certificate in ACME provider for domain \"*.traefik.wtf\" from a 'Host' rule",
			expectedDomains: nil,
		},
		{
			desc:            "no domain",
			domains:         types.Domain{},
			dnsChallenge:    nil,
			wildcardAllowed: true,
			expectedErr:     "unable to generate a certificate in ACME provider when no domain is given",
			expectedDomains: nil,
		},
		{
			desc:            "no DNSChallenge",
			domains:         types.Domain{Main: "*.traefik.wtf", SANs: []string{"foo.traefik.wtf"}},
			dnsChallenge:    nil,
			wildcardAllowed: true,
			expectedErr:     "unable to generate a wildcard certificate in ACME provider for domain \"*.traefik.wtf,foo.traefik.wtf\" : ACME needs a DNSChallenge",
			expectedDomains: nil,
		},
		{
			desc:            "unexpected SANs",
			domains:         types.Domain{Main: "*.traefik.wtf", SANs: []string{"foo.traefik.wtf"}},
			dnsChallenge:    &DNSChallenge{},
			wildcardAllowed: true,
			expectedErr:     "unable to generate a wildcard certificate in ACME provider for domain \"*.traefik.wtf,foo.traefik.wtf\" : SANs are not allowed",
			expectedDomains: nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {

			t.Parallel()

			acmeProvider := Provider{Configuration: &Configuration{}}
			if test.dnsChallenge != nil {
				acmeProvider.DNSChallenge = test.dnsChallenge
			}

			domains, err := acmeProvider.getValidDomains(test.domains, test.wildcardAllowed)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "Unexpected error.")
			} else {
				assert.Equal(t, len(test.expectedDomains), len(domains), "Unexpected domains.")
			}
		})
	}
}
