package acme

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	legoacme "github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/lego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/safe"
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

func TestProvider_checkRenewalInfo(t *testing.T) {
	testCases := []struct {
		desc               string
		disableARI         bool
		windowStart        time.Duration
		retryAfter         string
		expectedSuggested  bool
		expectedRetryAfter time.Duration
	}{
		{
			desc:              "Renewal window has passed",
			windowStart:       -48 * time.Hour,
			expectedSuggested: true,
		},
		{
			desc:        "Renewal window is in the future",
			windowStart: 48 * time.Hour,
		},
		{
			desc:               "Renewal window is in the future with Retry-After",
			windowStart:        48 * time.Hour,
			retryAfter:         "21600",
			expectedRetryAfter: 6 * time.Hour,
		},
		{
			desc:        "ACME server without renewal information",
			disableARI:  true,
			windowStart: -48 * time.Hour,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			crt, _ := newARITestCertificate(t)
			p := &Provider{client: newARITestClient(t, crt, test.disableARI, test.windowStart, test.retryAfter)}

			suggested, retryAfter := p.checkRenewalInfo(t.Context(), crt)
			assert.Equal(t, test.expectedSuggested, suggested)
			assert.Equal(t, test.expectedRetryAfter, retryAfter)
		})
	}
}

func TestProvider_renewCertificates_nextCheck(t *testing.T) {
	const renewInterval = 24 * time.Hour

	testCases := []struct {
		desc       string
		retryAfter string
		expected   time.Duration
	}{
		{
			desc:     "No Retry-After",
			expected: renewInterval,
		},
		{
			desc:       "Retry-After shorter than the renew interval",
			retryAfter: "21600",
			expected:   6 * time.Hour,
		},
		{
			desc:       "Retry-After longer than the renew interval",
			retryAfter: "172800",
			expected:   renewInterval,
		},
		{
			desc:       "Retry-After too short",
			retryAfter: "1",
			expected:   minRenewalInfoRetryAfter,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			crt, cert := newARITestCertificate(t)
			p := &Provider{
				client:       newARITestClient(t, crt, false, 48*time.Hour, test.retryAfter),
				certificates: []*CertAndStore{{Certificate: cert}},
			}

			nextCheck := p.renewCertificates(t.Context(), 30*24*time.Hour, renewInterval)
			assert.Equal(t, test.expected, nextCheck)
		})
	}
}

func newARITestClient(t *testing.T, crt *x509.Certificate, disableARI bool, windowStart time.Duration, retryAfter string) *lego.Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/dir", func(rw http.ResponseWriter, _ *http.Request) {
		dir := legoacme.Directory{
			NewNonceURL:   server.URL + "/nonce",
			NewAccountURL: server.URL + "/account",
			NewOrderURL:   server.URL + "/order",
		}
		if !disableARI {
			dir.RenewalInfo = server.URL + "/renewal-info"
		}
		_ = json.NewEncoder(rw).Encode(dir)
	})

	certID, err := api.MakeARICertID(crt)
	require.NoError(t, err)

	mux.HandleFunc("/renewal-info/"+certID, func(rw http.ResponseWriter, _ *http.Request) {
		if retryAfter != "" {
			rw.Header().Set("Retry-After", retryAfter)
		}

		start := time.Now().Add(windowStart)
		_ = json.NewEncoder(rw).Encode(legoacme.RenewalInfo{
			SuggestedWindow: legoacme.Window{Start: start, End: start.Add(24 * time.Hour)},
		})
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	config := lego.NewConfig(&Account{PrivateKey: x509.MarshalPKCS1PrivateKey(key)})
	config.CADirURL = server.URL + "/dir"
	config.HTTPClient = server.Client()

	client, err := lego.NewClient(config)
	require.NoError(t, err)

	return client
}

func newARITestCertificate(t *testing.T) (*x509.Certificate, Certificate) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:   big.NewInt(42),
		DNSNames:       []string{"example.com"},
		NotBefore:      time.Now().Add(-time.Hour),
		NotAfter:       time.Now().Add(90 * 24 * time.Hour),
		AuthorityKeyId: []byte{1, 2, 3, 4},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, key.Public(), key)
	require.NoError(t, err)

	crt, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	return crt, Certificate{
		Domain:      types.Domain{Main: "example.com"},
		Certificate: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		Key:         pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}),
	}
}
