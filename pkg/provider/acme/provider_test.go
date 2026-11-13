package acme

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/safe"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestGetUncheckedCertificates(t *testing.T) {
	t.Skip("Needs TLS Manager")
	wildcardMap := make(map[string]*tls.Certificate)
	wildcardMap["*.traefik.wtf"] = &tls.Certificate{}

	wildcardSafe := &safe.Safe{}
	wildcardSafe.Set(wildcardMap)

	domainMap := make(map[string]*tls.Certificate)
	domainMap["traefik.wtf"] = &tls.Certificate{}

	domainSafe := &safe.Safe{}
	domainSafe.Set(domainMap)

	// TODO Add a test for DefaultCertificate
	testCases := []struct {
		desc             string
		dynamicCerts     *safe.Safe
		resolvingDomains map[string]struct{}
		acmeCertificates []*CertAndStore
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
			desc:    "wildcard already exists in ACME certificates",
			domains: []string{"*.traefik.wtf"},
			acmeCertificates: []*CertAndStore{
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
			desc:    "domain CN already exists in ACME certificates and SANs to generate",
			domains: []string{"traefik.wtf", "foo.traefik.wtf"},
			acmeCertificates: []*CertAndStore{
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
			desc:    "domain already exists in ACME certificates",
			domains: []string{"traefik.wtf"},
			acmeCertificates: []*CertAndStore{
				{
					Domain: types.Domain{Main: "traefik.wtf"},
				},
			},
			expectedDomains: nil,
		},
		{
			desc:            "domain matched by wildcard in dynamic certificates",
			domains:         []string{"who.traefik.wtf", "foo.traefik.wtf"},
			dynamicCerts:    wildcardSafe,
			expectedDomains: nil,
		},
		{
			desc:    "domain matched by wildcard in ACME certificates",
			domains: []string{"who.traefik.wtf", "foo.traefik.wtf"},
			acmeCertificates: []*CertAndStore{
				{
					Domain: types.Domain{Main: "*.traefik.wtf"},
				},
			},
			expectedDomains: nil,
		},
		{
			desc:    "root domain with wildcard in ACME certificates",
			domains: []string{"traefik.wtf", "foo.traefik.wtf"},
			acmeCertificates: []*CertAndStore{
				{
					Domain: types.Domain{Main: "*.traefik.wtf"},
				},
			},
			expectedDomains: []string{"traefik.wtf"},
		},
		{
			desc:    "all domains already managed by ACME",
			domains: []string{"traefik.wtf", "foo.traefik.wtf"},
			resolvingDomains: map[string]struct{}{
				"traefik.wtf":     {},
				"foo.traefik.wtf": {},
			},
			expectedDomains: []string{},
		},
		{
			desc:    "one domain already managed by ACME",
			domains: []string{"traefik.wtf", "foo.traefik.wtf"},
			resolvingDomains: map[string]struct{}{
				"traefik.wtf": {},
			},
			expectedDomains: []string{"foo.traefik.wtf"},
		},
		{
			desc:    "wildcard domain already managed by ACME checks the domains",
			domains: []string{"bar.traefik.wtf", "foo.traefik.wtf"},
			resolvingDomains: map[string]struct{}{
				"*.traefik.wtf": {},
			},
			expectedDomains: []string{},
		},
		{
			desc:    "wildcard domain already managed by ACME checks domains and another domain checks one other domain, one domain still unchecked",
			domains: []string{"traefik.wtf", "bar.traefik.wtf", "foo.traefik.wtf", "acme.wtf"},
			resolvingDomains: map[string]struct{}{
				"*.traefik.wtf": {},
				"traefik.wtf":   {},
			},
			expectedDomains: []string{"acme.wtf"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.resolvingDomains == nil {
				test.resolvingDomains = make(map[string]struct{})
			}

			acmeProvider := Provider{
				// certificateStore: &traefiktls.CertificateStore{
				// 	DynamicCerts: test.dynamicCerts,
				// },
				certificates:     test.acmeCertificates,
				resolvingDomains: test.resolvingDomains,
			}

			domains := acmeProvider.getUncheckedDomains(t.Context(), test.domains, "default")
			assert.Len(t, domains, len(test.expectedDomains), "Unexpected domains.")
		})
	}
}

func TestProvider_sanitizeDomains(t *testing.T) {
	testCases := []struct {
		desc            string
		domains         types.Domain
		dnsChallenge    *DNSChallenge
		expectedErr     string
		expectedDomains []string
	}{
		{
			desc:            "valid wildcard",
			domains:         types.Domain{Main: "*.traefik.wtf"},
			dnsChallenge:    &DNSChallenge{},
			expectedErr:     "",
			expectedDomains: []string{"*.traefik.wtf"},
		},
		{
			desc:            "no wildcard",
			domains:         types.Domain{Main: "traefik.wtf", SANs: []string{"foo.traefik.wtf"}},
			dnsChallenge:    &DNSChallenge{},
			expectedErr:     "",
			expectedDomains: []string{"traefik.wtf", "foo.traefik.wtf"},
		},
		{
			desc:            "no domain",
			domains:         types.Domain{},
			dnsChallenge:    nil,
			expectedErr:     "no domain was given",
			expectedDomains: nil,
		},
		{
			desc:            "unauthorized wildcard with SAN",
			domains:         types.Domain{Main: "*.*.traefik.wtf", SANs: []string{"foo.traefik.wtf"}},
			dnsChallenge:    &DNSChallenge{},
			expectedErr:     "unable to generate a wildcard certificate in ACME provider for domains \"*.*.traefik.wtf,foo.traefik.wtf\" : ACME does not allow '*.*' wildcard domain",
			expectedDomains: nil,
		},
		{
			desc:            "unauthorized double wildcard with SAN",
			domains:         types.Domain{Main: "**.traefik.wtf", SANs: []string{"traefik.wtf"}},
			dnsChallenge:    &DNSChallenge{},
			expectedErr:     "unable to generate a wildcard certificate in ACME provider for domains \"**.traefik.wtf,traefik.wtf\" : ACME does not allow '**.' wildcard domain",
			expectedDomains: nil,
		},
		{
			desc:            "unauthorized double wildcard as SAN",
			domains:         types.Domain{Main: "traefik.wtf", SANs: []string{"**.traefik.wtf"}},
			dnsChallenge:    &DNSChallenge{},
			expectedErr:     "unable to generate a wildcard certificate in ACME provider for domains \"traefik.wtf,**.traefik.wtf\" : ACME does not allow '**.' wildcard domain",
			expectedDomains: nil,
		},
		{
			desc:            "wildcard and SANs",
			domains:         types.Domain{Main: "*.traefik.wtf", SANs: []string{"traefik.wtf"}},
			dnsChallenge:    &DNSChallenge{},
			expectedErr:     "",
			expectedDomains: []string{"*.traefik.wtf", "traefik.wtf"},
		},
		{
			desc:            "wildcard SANs",
			domains:         types.Domain{Main: "*.traefik.wtf", SANs: []string{"*.acme.wtf"}},
			dnsChallenge:    &DNSChallenge{},
			expectedErr:     "",
			expectedDomains: []string{"*.traefik.wtf", "*.acme.wtf"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			acmeProvider := Provider{Configuration: &Configuration{DNSChallenge: test.dnsChallenge}}

			domains, err := acmeProvider.sanitizeDomains(t.Context(), test.domains)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "Unexpected error.")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedDomains, domains, "Unexpected domains.")
			}
		})
	}
}

func TestDeleteUnnecessaryDomains(t *testing.T) {
	testCases := []struct {
		desc            string
		domains         []types.Domain
		expectedDomains []types.Domain
	}{
		{
			desc: "no domain to delete",
			domains: []types.Domain{
				{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf", "foo.bar"},
				},
				{
					Main: "*.foo.acme.wtf",
				},
				{
					Main: "acme02.wtf",
					SANs: []string{"traefik.acme02.wtf", "bar.foo"},
				},
			},
			expectedDomains: []types.Domain{
				{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf", "foo.bar"},
				},
				{
					Main: "*.foo.acme.wtf",
					SANs: []string{},
				},
				{
					Main: "acme02.wtf",
					SANs: []string{"traefik.acme02.wtf", "bar.foo"},
				},
			},
		},
		{
			desc: "wildcard and root domain",
			domains: []types.Domain{
				{
					Main: "acme.wtf",
				},
				{
					Main: "*.acme.wtf",
					SANs: []string{"acme.wtf"},
				},
			},
			expectedDomains: []types.Domain{
				{
					Main: "acme.wtf",
					SANs: []string{},
				},
				{
					Main: "*.acme.wtf",
					SANs: []string{},
				},
			},
		},
		{
			desc: "2 equals domains",
			domains: []types.Domain{
				{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf", "foo.bar"},
				},
				{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf", "foo.bar"},
				},
			},
			expectedDomains: []types.Domain{
				{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf", "foo.bar"},
				},
			},
		},
		{
			desc: "2 domains with same values",
			domains: []types.Domain{
				{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf"},
				},
				{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf", "foo.bar"},
				},
			},
			expectedDomains: []types.Domain{
				{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf"},
				},
				{
					Main: "foo.bar",
					SANs: []string{},
				},
			},
		},
		{
			desc: "domain totally checked by wildcard",
			domains: []types.Domain{
				{
					Main: "who.acme.wtf",
					SANs: []string{"traefik.acme.wtf", "bar.acme.wtf"},
				},
				{
					Main: "*.acme.wtf",
				},
			},
			expectedDomains: []types.Domain{
				{
					Main: "*.acme.wtf",
					SANs: []string{},
				},
			},
		},
		{
			desc: "duplicated wildcard",
			domains: []types.Domain{
				{
					Main: "*.acme.wtf",
					SANs: []string{"acme.wtf"},
				},
				{
					Main: "*.acme.wtf",
				},
			},
			expectedDomains: []types.Domain{
				{
					Main: "*.acme.wtf",
					SANs: []string{"acme.wtf"},
				},
			},
		},
		{
			desc: "domain partially checked by wildcard",
			domains: []types.Domain{
				{
					Main: "traefik.acme.wtf",
					SANs: []string{"acme.wtf", "foo.bar"},
				},
				{
					Main: "*.acme.wtf",
				},
				{
					Main: "who.acme.wtf",
					SANs: []string{"traefik.acme.wtf", "bar.acme.wtf"},
				},
			},
			expectedDomains: []types.Domain{
				{
					Main: "acme.wtf",
					SANs: []string{"foo.bar"},
				},
				{
					Main: "*.acme.wtf",
					SANs: []string{},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domains := deleteUnnecessaryDomains(t.Context(), test.domains)
			assert.Equal(t, test.expectedDomains, domains, "unexpected domain")
		})
	}
}

func TestIsAccountMatchingCaServer(t *testing.T) {
	testCases := []struct {
		desc       string
		accountURI string
		serverURI  string
		expected   bool
	}{
		{
			desc:       "acme staging with matching account",
			accountURI: "https://acme-staging-v02.api.letsencrypt.org/acme/acct/1234567",
			serverURI:  "https://acme-staging-v02.api.letsencrypt.org/acme/directory",
			expected:   true,
		},
		{
			desc:       "acme production with matching account",
			accountURI: "https://acme-v02.api.letsencrypt.org/acme/acct/1234567",
			serverURI:  "https://acme-v02.api.letsencrypt.org/acme/directory",
			expected:   true,
		},
		{
			desc:       "http only acme with matching account",
			accountURI: "http://acme.api.letsencrypt.org/acme/acct/1234567",
			serverURI:  "http://acme.api.letsencrypt.org/acme/directory",
			expected:   true,
		},
		{
			desc:       "different subdomains for account and server",
			accountURI: "https://test1.example.org/acme/acct/1234567",
			serverURI:  "https://test2.example.org/acme/directory",
			expected:   false,
		},
		{
			desc:       "different domains for account and server",
			accountURI: "https://test.example1.org/acme/acct/1234567",
			serverURI:  "https://test.example2.org/acme/directory",
			expected:   false,
		},
		{
			desc:       "different tld for account and server",
			accountURI: "https://test.example.com/acme/acct/1234567",
			serverURI:  "https://test.example.org/acme/directory",
			expected:   false,
		},
		{
			desc:       "malformed account url",
			accountURI: "//|\\/test.example.com/acme/acct/1234567",
			serverURI:  "https://test.example.com/acme/directory",
			expected:   false,
		},
		{
			desc:       "malformed server url",
			accountURI: "https://test.example.com/acme/acct/1234567",
			serverURI:  "//|\\/test.example.com/acme/directory",
			expected:   false,
		},
		{
			desc:       "malformed server and account url",
			accountURI: "//|\\/test.example.com/acme/acct/1234567",
			serverURI:  "//|\\/test.example.com/acme/directory",
			expected:   false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := isAccountMatchingCaServer(t.Context(), test.accountURI, test.serverURI)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestInitAccount(t *testing.T) {
	testCases := []struct {
		desc            string
		account         *Account
		email           string
		keyType         string
		expectedAccount *Account
	}{
		{
			desc: "Existing account with all information",
			account: &Account{
				Email: "foo@foo.net",
			},
			expectedAccount: &Account{
				Email: "foo@foo.net",
			},
		},
		{
			desc:    "Account nil",
			email:   "foo@foo.net",
			keyType: "EC256",
			expectedAccount: &Account{
				Email: "foo@foo.net",
			},
		},
		{
			desc:    "Existing account with no email",
			account: &Account{},
			email:   "foo@foo.net",
			keyType: "4096",
			expectedAccount: &Account{
				Email: "foo@foo.net",
			},
		},
		{
			desc: "Existing account with no key type",
			account: &Account{
				Email: "foo@foo.net",
			},
			email:   "bar@foo.net",
			keyType: "EC256",
			expectedAccount: &Account{
				Email: "foo@foo.net",
			},
		},
		{
			desc: "Existing account and provider with no key type",
			account: &Account{
				Email: "foo@foo.net",
			},
			email: "bar@foo.net",
			expectedAccount: &Account{
				Email: "foo@foo.net",
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			acmeProvider := Provider{account: test.account, Configuration: &Configuration{Email: test.email, KeyType: test.keyType}}

			actualAccount, err := acmeProvider.initAccount()
			assert.NoError(t, err, "Init account in error")
			assert.Equal(t, test.expectedAccount.Email, actualAccount.Email, "unexpected email account")
		})
	}
}

func Test_getCertificateRenewDurations(t *testing.T) {
	testCases := []struct {
		desc                  string
		certificatesDurations int
		expectRenewPeriod     time.Duration
		expectRenewInterval   time.Duration
	}{
		{
			desc:                  "Less than 24 Hours certificates: 20 minutes renew period, 1 minutes renew interval",
			certificatesDurations: 1,
			expectRenewPeriod:     time.Minute * 20,
			expectRenewInterval:   time.Minute,
		},
		{
			desc:                  "1 Year certificates: 4 months renew period, 1 week renew interval",
			certificatesDurations: 24 * 365,
			expectRenewPeriod:     time.Hour * 24 * 30 * 4,
			expectRenewInterval:   time.Hour * 24 * 7,
		},
		{
			desc:                  "265 Days certificates: 30 days renew period, 1 day renew interval",
			certificatesDurations: 24 * 265,
			expectRenewPeriod:     time.Hour * 24 * 30,
			expectRenewInterval:   time.Hour * 24,
		},
		{
			desc:                  "90 Days certificates: 30 days renew period, 1 day renew interval",
			certificatesDurations: 24 * 90,
			expectRenewPeriod:     time.Hour * 24 * 30,
			expectRenewInterval:   time.Hour * 24,
		},
		{
			desc:                  "45 Days certificates (Let's Encrypt 2028 standard): 10 days renew period, 12 hour renew interval",
			certificatesDurations: 24 * 45,
			expectRenewPeriod:     time.Hour * 24 * 10,
			expectRenewInterval:   time.Hour * 12,
		},
		{
			desc:                  "30 Days certificates: 10 days renew period, 12 hour renew interval",
			certificatesDurations: 24 * 30,
			expectRenewPeriod:     time.Hour * 24 * 10,
			expectRenewInterval:   time.Hour * 12,
		},
		{
			desc:                  "7 Days certificates: 2 days renew period, 2 hour renew interval",
			certificatesDurations: 24 * 7,
			expectRenewPeriod:     time.Hour * 24 * 2,
			expectRenewInterval:   time.Hour * 2,
		},
		{
			desc:                  "160 hour certificate (Let's Encrypt 'shortlived' profile): 2 days renew period, 2 hour renew interval",
			certificatesDurations: 160,
			expectRenewPeriod:     time.Hour * 24 * 2,
			expectRenewInterval:   time.Hour * 2,
		},
		{
			desc:                  "24 Hours certificates: 6 hours renew period, 10 minutes renew interval",
			certificatesDurations: 24,
			expectRenewPeriod:     time.Hour * 6,
			expectRenewInterval:   time.Minute * 10,
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			renewPeriod, renewInterval := getCertificateRenewDurations(test.certificatesDurations)
			assert.Equal(t, test.expectRenewPeriod, renewPeriod)
			assert.Equal(t, test.expectRenewInterval, renewInterval)
		})
	}
}

func TestProvider_requestedDomainsFrom(t *testing.T) {
	config := dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{
				"rule": {
					Rule: "Host(`Foo.example.com`) || Host(`bar.example.com.`)",
					TLS:  &dynamic.RouterTLSConfig{CertResolver: "le"},
				},
				"domains": {
					Rule: "Host(`app.example.org`)",
					TLS: &dynamic.RouterTLSConfig{
						CertResolver: "le",
						Domains:      []types.Domain{{Main: "*.example.org", SANs: []string{"example.org"}}},
					},
				},
				"other-resolver": {
					Rule: "Host(`other.example.com`)",
					TLS:  &dynamic.RouterTLSConfig{CertResolver: "other"},
				},
				"no-resolver": {
					Rule: "Host(`tls.example.com`)",
					TLS:  &dynamic.RouterTLSConfig{},
				},
				"no-tls": {
					Rule: "Host(`plain.example.com`)",
				},
				"invalid-rule": {
					Rule: "Host(`broken.example.com`",
					TLS:  &dynamic.RouterTLSConfig{CertResolver: "le"},
				},
			},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers: map[string]*dynamic.TCPRouter{
				"sni": {
					Rule: "HostSNI(`tcp.example.com`)",
					TLS: &dynamic.RouterTCPTLSConfig{
						CertResolver: "le",
						Domains:      []types.Domain{{Main: "tcp-domains.example.com"}},
					},
				},
				"passthrough": {
					Rule: "HostSNI(`*`)",
					TLS:  &dynamic.RouterTCPTLSConfig{Passthrough: true},
				},
			},
		},
		TLS: &dynamic.TLSConfiguration{
			Stores: map[string]traefiktls.Store{
				"default": {
					DefaultGeneratedCert: &traefiktls.GeneratedCert{
						Resolver: "other",
						Domain:   &types.Domain{Main: "default.example.com"},
					},
				},
				"user-default": {
					DefaultCertificate: &traefiktls.Certificate{},
					DefaultGeneratedCert: &traefiktls.GeneratedCert{
						Resolver: "le",
						Domain:   &types.Domain{Main: "user.example.com"},
					},
				},
			},
		},
	}

	p := &Provider{ResolverName: "le"}

	expected := []string{
		"foo.example.com",
		"bar.example.com",
		"app.example.org",
		"*.example.org",
		"example.org",
		"other.example.com",
		"tls.example.com",
		"tcp.example.com",
		"tcp-domains.example.com",
		"default.example.com",
	}
	domains, bounded := p.requestedDomainsFrom(t.Context(), config)
	assert.ElementsMatch(t, expected, domains)
	assert.True(t, bounded)
}

func TestProvider_requestedDomainsFrom_bounded(t *testing.T) {
	testCases := []struct {
		desc     string
		rule     string
		tcp      bool
		cname    bool
		expected bool
	}{
		{desc: "host with CNAME flattening", rule: "Host(`a.example.com`)", cname: true},
		{desc: "sni with CNAME flattening", rule: "HostSNI(`a.example.com`)", tcp: true, cname: true, expected: true},
		{desc: "host", rule: "Host(`a.example.com`)", expected: true},
		{desc: "host and path", rule: "Host(`a.example.com`) && PathPrefix(`/x`)", expected: true},
		{desc: "host or host", rule: "Host(`a.example.com`) || Host(`b.example.com`)", expected: true},
		{desc: "host or path", rule: "Host(`a.example.com`) || PathPrefix(`/x`)"},
		{desc: "wildcard host", rule: "Host(`*.example.com`)", expected: true},
		{desc: "catch-all host", rule: "Host(`*`)"},
		{desc: "v2 host header", rule: "HostHeader(`a.example.com`)", expected: true},
		{desc: "path only", rule: "PathPrefix(`/`)"},
		{desc: "host regexp", rule: "HostRegexp(`^.+\\.example\\.com$`)"},
		{desc: "negated host", rule: "!Host(`a.example.com`)"},
		{desc: "sni", rule: "HostSNI(`a.example.com`)", tcp: true, expected: true},
		{desc: "sni catch-all", rule: "HostSNI(`*`)", tcp: true},
		{desc: "sni regexp", rule: "HostSNIRegexp(`^.+\\.example\\.com$`)", tcp: true},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			config := dynamic.Configuration{}
			if test.tcp {
				config.TCP = &dynamic.TCPConfiguration{Routers: map[string]*dynamic.TCPRouter{
					"r": {Rule: test.rule, TLS: &dynamic.RouterTCPTLSConfig{}},
				}}
			} else {
				config.HTTP = &dynamic.HTTPConfiguration{Routers: map[string]*dynamic.Router{
					"r": {Rule: test.rule, TLS: &dynamic.RouterTLSConfig{}},
				}}
			}

			_, bounded := (&Provider{ResolverName: "le", CNAMEFlattening: test.cname}).requestedDomainsFrom(t.Context(), config)
			assert.Equal(t, test.expected, bounded)
		})
	}
}

func TestProvider_markDomainsRequested(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := &Provider{configuredAt: start}

	p.markDomainsRequested([]string{"a.example.com", "b.example.com"}, true, start.Add(time.Hour), time.Hour)
	p.markDomainsRequested([]string{"b.example.com"}, true, start.Add(2*time.Hour), time.Hour)

	now := start.Add(72 * time.Hour)
	assert.Equal(t, now, p.lastRequested([]string{"b.example.com"}, now), "still requested, however old the configuration")
	assert.Equal(t, start.Add(2*time.Hour), p.lastRequested([]string{"a.example.com"}, now), "dropped")
	assert.Equal(t, start, p.lastRequested([]string{"c.example.com"}, now), "never requested")

	p.markDomainsRequested(nil, true, start.Add(4*time.Hour), time.Hour)

	assert.Empty(t, p.requestedDomains)
	assert.Equal(t, map[string]time.Time{"b.example.com": start.Add(4 * time.Hour)}, p.droppedDomains, "a.example.com is past retention")
}

func TestProvider_lastRequested_beforeConfiguration(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := &Provider{}

	now := start.Add(72 * time.Hour)
	assert.Equal(t, now, p.lastRequested([]string{"a.example.com"}, now), "no configuration recorded yet")

	p.markDomainsRequested(nil, true, now, time.Hour)
	assert.Equal(t, now, p.lastRequested([]string{"a.example.com"}, now.Add(time.Hour)), "unused from the first configuration")
}

func TestProvider_markDomainsRequested_unbounded(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := &Provider{configuredAt: start, skippedDomains: [][]string{{"a.example.com"}}}

	assert.True(t, p.markDomainsRequested(nil, false, start.Add(time.Hour), time.Hour), "every skipped certificate is used again")
	assert.False(t, p.markDomainsRequested(nil, false, start.Add(2*time.Hour), time.Hour), "already unbounded")

	now := start.Add(72 * time.Hour)
	assert.Equal(t, now, p.lastRequested([]string{"c.example.com"}, now), "any certificate is used")

	p.markDomainsRequested(nil, true, start.Add(3*time.Hour), time.Hour)
	assert.Equal(t, start.Add(3*time.Hour), p.lastRequested([]string{"c.example.com"}, now), "used until the configuration was bounded")
}

func TestProvider_markDomainsRequested_revived(t *testing.T) {
	now := time.Now()
	p := &Provider{skippedDomains: [][]string{{"*.example.com"}}}

	assert.False(t, p.markDomainsRequested([]string{"other.example.org"}, true, now, time.Hour), "not covered by a skipped certificate")
	assert.True(t, p.markDomainsRequested([]string{"other.example.org", "app.example.com"}, true, now, time.Hour), "newly requested")
	assert.False(t, p.markDomainsRequested([]string{"other.example.org", "app.example.com"}, true, now, time.Hour), "already requested")
}

func TestCoversDomain(t *testing.T) {
	testCases := []struct {
		certDomain string
		domain     string
		expected   bool
	}{
		{certDomain: "a.example.com", domain: "a.example.com", expected: true},
		{certDomain: "*.example.com", domain: "a.example.com", expected: true},
		{certDomain: "*.example.com", domain: "example.com"},
		{certDomain: "a.example.com", domain: "*.example.com", expected: true},
		{certDomain: "a.b.example.com", domain: "*.example.com"},
		{certDomain: "a.b.example.com", domain: "**.example.com", expected: true},
		{certDomain: "*.b.example.com", domain: "**.example.com", expected: true},
		{certDomain: "*.example.com", domain: "*.example.com", expected: true},
		{certDomain: "example.com", domain: "*.example.com"},
		{certDomain: "a.example.org", domain: "*.example.com"},
		{certDomain: "A.Example.com", domain: "a.example.com", expected: true},
		{certDomain: "*.Example.com.", domain: "a.example.com", expected: true},
	}

	for _, test := range testCases {
		t.Run(test.certDomain+" for "+test.domain, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, coversDomain([]string{test.certDomain}, test.domain))
		})
	}
}

func TestProvider_coversRequestedLocked(t *testing.T) {
	certs := [][]string{{"*.example.com"}}

	p := &Provider{requestedDomains: map[string]struct{}{"other.example.org": {}}}
	assert.False(t, p.coversRequestedLocked(certs))

	p.requestedDomains["app.example.com"] = struct{}{}
	assert.True(t, p.coversRequestedLocked(certs), "requested again during selection")

	p = &Provider{unbounded: true}
	assert.True(t, p.coversRequestedLocked(certs), "a router can serve any domain")
	assert.False(t, p.coversRequestedLocked(nil))
}

func TestProvider_renewCertificates_unused(t *testing.T) {
	now := time.Now()
	cert := &CertAndStore{Certificate: generateCertificate(t, types.Domain{Main: "a.example.com"}, now.Add(24*time.Hour), "a.example.com")}

	p := &Provider{
		configuredAt: now.Add(-72 * time.Hour),
		certificates: []*CertAndStore{cert},
		renewNow:     make(chan struct{}, 1),
	}

	p.renewCertificates(t.Context(), 30*24*time.Hour, time.Hour)

	assert.Equal(t, [][]string{{"a.example.com"}}, p.skippedDomains)
	assert.Empty(t, p.renewNow)
	assert.Equal(t, []*CertAndStore{cert}, p.certificates)
}

func TestProvider_selectCertificates(t *testing.T) {
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	renewPeriod := 30 * 24 * time.Hour
	unusedAfter := 24 * time.Hour

	testCases := []struct {
		desc            string
		configuredAt    time.Time
		requested       []string
		dropped         map[string]time.Time
		unbounded       bool
		domain          types.Domain
		sans            []string
		notAfter        time.Time
		broken          bool
		expectedRenew   bool
		expectedRemoval bool
		expectedUnused  bool
	}{
		{
			desc:          "unused, a router can serve any domain",
			unbounded:     true,
			sans:          []string{"a.example.com"},
			notAfter:      now.Add(-time.Hour),
			expectedRenew: true,
		},
		{
			desc:          "used, in renew period",
			requested:     []string{"b.example.com"},
			sans:          []string{"a.example.com", "b.example.com"},
			notAfter:      now.Add(10 * 24 * time.Hour),
			expectedRenew: true,
		},
		{
			desc:      "used, not in renew period",
			requested: []string{"a.example.com"},
			sans:      []string{"a.example.com"},
			notAfter:  now.Add(60 * 24 * time.Hour),
		},
		{
			desc:           "unused, in renew period",
			sans:           []string{"a.example.com"},
			notAfter:       now.Add(10 * 24 * time.Hour),
			expectedUnused: true,
		},
		{
			desc:     "unused, not in renew period",
			sans:     []string{"a.example.com"},
			notAfter: now.Add(60 * 24 * time.Hour),
		},
		{
			desc:           "dropped longer ago than unusedAfter",
			dropped:        map[string]time.Time{"a.example.com": now.Add(-48 * time.Hour)},
			sans:           []string{"a.example.com"},
			notAfter:       now.Add(10 * 24 * time.Hour),
			expectedUnused: true,
		},
		{
			desc:          "dropped recently",
			dropped:       map[string]time.Time{"a.example.com": now.Add(-time.Hour)},
			sans:          []string{"a.example.com"},
			notAfter:      now.Add(10 * 24 * time.Hour),
			expectedRenew: true,
		},
		{
			desc:            "unused and expired",
			sans:            []string{"a.example.com"},
			notAfter:        now.Add(-time.Hour),
			expectedRemoval: true,
			expectedUnused:  true,
		},
		{
			desc:          "used and expired",
			requested:     []string{"a.example.com"},
			sans:          []string{"a.example.com"},
			notAfter:      now.Add(-time.Hour),
			expectedRenew: true,
		},
		{
			desc:          "unused, provider started recently",
			configuredAt:  now.Add(-time.Hour),
			sans:          []string{"a.example.com"},
			notAfter:      now.Add(10 * 24 * time.Hour),
			expectedRenew: true,
		},
		{
			desc:          "wildcard certificate used by a subdomain",
			requested:     []string{"app.example.com"},
			sans:          []string{"*.example.com"},
			notAfter:      now.Add(10 * 24 * time.Hour),
			expectedRenew: true,
		},
		{
			desc:          "stored domain narrower than the certificate",
			requested:     []string{"a.example.com"},
			domain:        types.Domain{Main: "b.example.com"},
			sans:          []string{"a.example.com", "b.example.com"},
			notAfter:      now.Add(10 * 24 * time.Hour),
			expectedRenew: true,
		},
		{
			desc:          "broken certificate, used",
			requested:     []string{"a.example.com"},
			domain:        types.Domain{Main: "a.example.com"},
			broken:        true,
			expectedRenew: true,
		},
		{
			desc:           "broken certificate, unused",
			domain:         types.Domain{Main: "a.example.com"},
			broken:         true,
			expectedUnused: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			configuredAt := test.configuredAt
			if configuredAt.IsZero() {
				configuredAt = now.Add(-72 * time.Hour)
			}

			p := &Provider{configuredAt: configuredAt, droppedDomains: test.dropped, unbounded: test.unbounded}
			p.requestedDomains = make(map[string]struct{})
			for _, domain := range test.requested {
				p.requestedDomains[domain] = struct{}{}
			}

			certificate := Certificate{Domain: test.domain, Certificate: []byte("broken")}
			if !test.broken {
				certificate = generateCertificate(t, test.domain, test.notAfter, test.sans...)
			}

			certs := []*CertAndStore{{Certificate: certificate}}
			toRenew, toRemove, unused := p.selectCertificates(t.Context(), certs, now, renewPeriod, unusedAfter)

			assert.Equal(t, test.expectedRenew, len(toRenew) == 1, "renew")
			assert.Equal(t, test.expectedRemoval, len(toRemove) == 1, "remove")
			assert.Equal(t, test.expectedUnused, len(unused) == 1, "unused")
		})
	}
}

func TestProvider_removeCertificates(t *testing.T) {
	now := time.Now()

	expired := &CertAndStore{Certificate: generateCertificate(t, types.Domain{Main: "expired.example.com"}, now.Add(-time.Hour), "expired.example.com")}
	valid := &CertAndStore{Certificate: generateCertificate(t, types.Domain{Main: "valid.example.com"}, now.Add(time.Hour), "valid.example.com")}
	renewed := &CertAndStore{Certificate: generateCertificate(t, types.Domain{Main: "renewed.example.com"}, now.Add(time.Hour), "renewed.example.com")}

	store := &fakeStore{}
	configurationChan := make(chan dynamic.Message, 1)
	p := &Provider{
		ResolverName:      "le",
		Store:             store,
		configuredAt:      now.Add(-72 * time.Hour),
		certificates:      []*CertAndStore{expired, valid, renewed},
		configurationChan: configurationChan,
	}

	// renewed was selected while expired, then renewed before the removal.
	saved := p.certificates
	err := p.removeCertificates(t.Context(), []*CertAndStore{expired, renewed}, now, time.Hour)
	require.NoError(t, err)

	assert.Equal(t, []*CertAndStore{valid, renewed}, p.certificates)
	assert.Equal(t, []*CertAndStore{valid, renewed}, store.certificates)
	assert.Equal(t, []*CertAndStore{expired, valid, renewed}, saved, "a save in progress keeps its slice")

	msg := <-configurationChan
	assert.Len(t, msg.Configuration.TLS.Certificates, 2)

	// Nothing left to remove: no configuration push, no save.
	store.certificates = nil
	err = p.removeCertificates(t.Context(), []*CertAndStore{renewed}, now, time.Hour)
	require.NoError(t, err)

	assert.Nil(t, store.certificates)
	assert.Empty(t, configurationChan)
}

func TestProvider_removeCertificates_requestedAgain(t *testing.T) {
	now := time.Now()
	expired := &CertAndStore{Certificate: generateCertificate(t, types.Domain{Main: "a.example.com"}, now.Add(-time.Hour), "a.example.com")}

	store := &fakeStore{}
	configurationChan := make(chan dynamic.Message, 1)
	p := &Provider{
		Store:             store,
		certificates:      []*CertAndStore{expired},
		configurationChan: configurationChan,
		renewNow:          make(chan struct{}, 1),
	}

	// A configuration asked for its domain after the certificate was selected.
	p.markDomainsRequested([]string{"a.example.com"}, true, now, time.Hour)

	err := p.removeCertificates(t.Context(), []*CertAndStore{expired}, now, time.Hour)
	require.NoError(t, err)

	assert.Equal(t, []*CertAndStore{expired}, p.certificates)
	assert.Nil(t, store.certificates)
	assert.Empty(t, configurationChan)
	assert.Len(t, p.renewNow, 1, "renewal pass requested")
}

func TestProvider_watchNewDomains_revivesSkipped(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	pool := safe.NewPool(ctx)
	defer pool.Stop()

	p := &Provider{
		Configuration:          &Configuration{},
		ResolverName:           "le",
		pool:                   pool,
		configFromListenerChan: make(chan dynamic.Configuration),
		renewNow:               make(chan struct{}, 1),
		configuredAt:           time.Now().Add(-72 * time.Hour),
		skippedDomains:         [][]string{{"a.example.com"}},
	}
	p.watchNewDomains(ctx, time.Hour)

	// A router of another resolver is served the skipped certificate by SNI.
	p.configFromListenerChan <- dynamic.Configuration{HTTP: &dynamic.HTTPConfiguration{
		Routers: map[string]*dynamic.Router{
			"r": {Rule: "Host(`a.example.com`)", TLS: &dynamic.RouterTLSConfig{CertResolver: "other"}},
		},
	}}

	select {
	case <-p.renewNow:
	case <-time.After(5 * time.Second):
		t.Fatal("no renewal pass requested")
	}
}

func TestProvider_watchRenewals(t *testing.T) {
	p := &Provider{renewNow: make(chan struct{}, 1)}

	renewed := make(chan struct{}, 2)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.watchRenewals(ctx, nil, func() { renewed <- struct{}{} })
	}()

	// A router asking for a skipped certificate again, before the next tick.
	p.signalRenew()

	select {
	case <-renewed:
	case <-time.After(5 * time.Second):
		t.Fatal("no renewal pass after signalRenew")
	}

	cancel()
	<-done
}

type fakeStore struct {
	certificates []*CertAndStore
}

func (s *fakeStore) GetAccount(string) (*Account, error) { return nil, nil }

func (s *fakeStore) SaveAccount(string, *Account) error { return nil }

func (s *fakeStore) GetCertificates(string) ([]*CertAndStore, error) { return s.certificates, nil }

func (s *fakeStore) SaveCertificates(_ string, certificates []*CertAndStore) error {
	s.certificates = certificates
	return nil
}

func generateCertificate(t *testing.T, domain types.Domain, notAfter time.Time, sans ...string) Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    notAfter.Add(-90 * 24 * time.Hour),
		NotAfter:     notAfter,
		DNSNames:     sans,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	return Certificate{
		Domain:      domain,
		Certificate: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		Key:         pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}),
	}
}
