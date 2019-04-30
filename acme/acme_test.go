package acme

import (
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/tls/generate"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestDomainsSet(t *testing.T) {
	testCases := []struct {
		input    string
		expected types.Domains
	}{
		{
			input:    "",
			expected: types.Domains{},
		},
		{
			input: "foo1.com",
			expected: types.Domains{
				types.Domain{Main: "foo1.com"},
			},
		},
		{
			input: "foo2.com,bar.net",
			expected: types.Domains{
				types.Domain{
					Main: "foo2.com",
					SANs: []string{"bar.net"},
				},
			},
		},
		{
			input: "foo3.com,bar1.net,bar2.net,bar3.net",
			expected: types.Domains{
				types.Domain{
					Main: "foo3.com",
					SANs: []string{"bar1.net", "bar2.net", "bar3.net"},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()

			domains := types.Domains{}
			domains.Set(test.input)
			assert.Exactly(t, test.expected, domains)
		})
	}
}

func TestDomainsSetAppend(t *testing.T) {
	testCases := []struct {
		input    string
		expected types.Domains
	}{
		{
			input:    "",
			expected: types.Domains{},
		},
		{
			input: "foo1.com",
			expected: types.Domains{
				types.Domain{Main: "foo1.com"},
			},
		},
		{
			input: "foo2.com,bar.net",
			expected: types.Domains{
				types.Domain{Main: "foo1.com"},
				types.Domain{
					Main: "foo2.com",
					SANs: []string{"bar.net"},
				},
			},
		},
		{
			input: "foo3.com,bar1.net,bar2.net,bar3.net",
			expected: types.Domains{
				types.Domain{Main: "foo1.com"},
				types.Domain{
					Main: "foo2.com",
					SANs: []string{"bar.net"},
				},
				types.Domain{
					Main: "foo3.com",
					SANs: []string{"bar1.net", "bar2.net", "bar3.net"},
				},
			},
		},
	}

	// append to
	domains := types.Domains{}
	for _, test := range testCases {
		t.Run(test.input, func(t *testing.T) {

			domains.Set(test.input)
			assert.Exactly(t, test.expected, domains)
		})
	}
}

func TestCertificatesRenew(t *testing.T) {
	foo1Cert, foo1Key, _ := generate.KeyPair("foo1.com", time.Now())
	foo2Cert, foo2Key, _ := generate.KeyPair("foo2.com", time.Now())

	domainsCertificates := DomainsCertificates{
		lock: sync.RWMutex{},
		Certs: []*DomainsCertificate{
			{
				Domains: types.Domain{
					Main: "foo1.com"},
				Certificate: &Certificate{
					Domain:        "foo1.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    foo1Key,
					Certificate:   foo1Cert,
				},
			},
			{
				Domains: types.Domain{
					Main: "foo2.com"},
				Certificate: &Certificate{
					Domain:        "foo2.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    foo2Key,
					Certificate:   foo2Cert,
				},
			},
		},
	}

	foo1Cert, foo1Key, _ = generate.KeyPair("foo1.com", time.Now())
	newCertificate := &Certificate{
		Domain:        "foo1.com",
		CertURL:       "url",
		CertStableURL: "url",
		PrivateKey:    foo1Key,
		Certificate:   foo1Cert,
	}

	err := domainsCertificates.renewCertificates(newCertificate, types.Domain{Main: "foo1.com"})
	if err != nil {
		t.Errorf("Error in renewCertificates :%v", err)
	}

	if len(domainsCertificates.Certs) != 2 {
		t.Errorf("Expected domainsCertificates length %d %+v\nGot %+v", 2, domainsCertificates.Certs, len(domainsCertificates.Certs))
	}

	if !reflect.DeepEqual(domainsCertificates.Certs[0].Certificate, newCertificate) {
		t.Errorf("Expected new certificate %+v \nGot %+v", newCertificate, domainsCertificates.Certs[0].Certificate)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	now := time.Now()
	fooCert, fooKey, _ := generate.KeyPair("foo.com", now)
	foo24Cert, foo24Key, _ := generate.KeyPair("foo.com", now.Add(24*time.Hour))
	foo48Cert, foo48Key, _ := generate.KeyPair("foo.com", now.Add(48*time.Hour))
	barCert, barKey, _ := generate.KeyPair("bar.com", now)
	domainsCertificates := DomainsCertificates{
		lock: sync.RWMutex{},
		Certs: []*DomainsCertificate{
			{
				Domains: types.Domain{
					Main: "foo.com"},
				Certificate: &Certificate{
					Domain:        "foo.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    foo24Key,
					Certificate:   foo24Cert,
				},
			},
			{
				Domains: types.Domain{
					Main: "foo.com"},
				Certificate: &Certificate{
					Domain:        "foo.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    foo48Key,
					Certificate:   foo48Cert,
				},
			},
			{
				Domains: types.Domain{
					Main: "foo.com"},
				Certificate: &Certificate{
					Domain:        "foo.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    fooKey,
					Certificate:   fooCert,
				},
			},
			{
				Domains: types.Domain{
					Main: "bar.com"},
				Certificate: &Certificate{
					Domain:        "bar.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    barKey,
					Certificate:   barCert,
				},
			},
			{
				Domains: types.Domain{
					Main: "foo.com"},
				Certificate: &Certificate{
					Domain:        "foo.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    foo48Key,
					Certificate:   foo48Cert,
				},
			},
		},
	}
	domainsCertificates.Init()

	if len(domainsCertificates.Certs) != 2 {
		t.Errorf("Expected domainsCertificates length %d %+v\nGot %+v", 2, domainsCertificates.Certs, len(domainsCertificates.Certs))
	}

	for _, cert := range domainsCertificates.Certs {
		switch cert.Domains.Main {
		case "bar.com":
			continue
		case "foo.com":
			if !cert.tlsCert.Leaf.NotAfter.Equal(now.Add(48 * time.Hour).Truncate(1 * time.Second)) {
				t.Errorf("Bad expiration %s date for domain %+v, now %s", cert.tlsCert.Leaf.NotAfter.String(), cert, now.Add(48*time.Hour).Truncate(1*time.Second).String())
			}
		default:
			t.Errorf("Unknown domain %+v", cert)
		}
	}
}

func TestAcmeClientCreation(t *testing.T) {
	// Lengthy setup to avoid external web requests - oh for easier golang testing!
	account := &Account{Email: "f@f"}

	account.PrivateKey, _ = base64.StdEncoding.DecodeString(`
MIIBPAIBAAJBAMp2Ni92FfEur+CAvFkgC12LT4l9D53ApbBpDaXaJkzzks+KsLw9zyAxvlrfAyTCQ
7tDnEnIltAXyQ0uOFUUdcMCAwEAAQJAK1FbipATZcT9cGVa5x7KD7usytftLW14heQUPXYNV80r/3
lmnpvjL06dffRpwkYeN8DATQF/QOcy3NNNGDw/4QIhAPAKmiZFxA/qmRXsuU8Zhlzf16WrNZ68K64
asn/h3qZrAiEA1+wFR3WXCPIolOvd7AHjfgcTKQNkoMPywU4FYUNQ1AkCIQDv8yk0qPjckD6HVCPJ
llJh9MC0svjevGtNlxJoE3lmEQIhAKXy1wfZ32/XtcrnENPvi6lzxI0T94X7s5pP3aCoPPoJAiEAl
cijFkALeQp/qyeXdFld2v9gUN3eCgljgcl0QweRoIc=---`)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{
  "GPHhmRVEDas": "https://community.letsencrypt.org/t/adding-random-entries-to-the-directory/33417",
  "keyChange": "https://foo/acme/key-change",
  "meta": {
    "termsOfService": "https://boulder:4431/terms/v7"
  },
  "newAccount": "https://foo/acme/new-acct",
  "newNonce": "https://foo/acme/new-nonce",
  "newOrder": "https://foo/acme/new-order",
  "revokeCert": "https://foo/acme/revoke-cert"
}`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	a := ACME{
		CAServer: ts.URL,
		DNSChallenge: &acmeprovider.DNSChallenge{
			Provider:                "manual",
			DelayBeforeCheck:        10,
			DisablePropagationCheck: true,
		},
	}

	client, err := a.buildACMEClient(account)
	if err != nil {
		t.Errorf("Error in buildACMEClient: %v", err)
	}
	if client == nil {
		t.Error("No client from buildACMEClient!")
	}
}

func TestAcme_getUncheckedCertificates(t *testing.T) {
	mm := make(map[string]*tls.Certificate)
	mm["*.containo.us"] = &tls.Certificate{}
	mm["traefik.acme.io"] = &tls.Certificate{}

	dm := make(map[string]struct{})
	dm["*.traefik.wtf"] = struct{}{}

	a := ACME{TLSConfig: &tls.Config{NameToCertificate: mm}, resolvingDomains: dm}

	domains := []string{"traefik.containo.us", "trae.containo.us", "foo.traefik.wtf"}
	uncheckedDomains := a.getUncheckedDomains(domains, nil)
	assert.Empty(t, uncheckedDomains)
	domains = []string{"traefik.acme.io", "trae.acme.io"}
	uncheckedDomains = a.getUncheckedDomains(domains, nil)
	assert.Len(t, uncheckedDomains, 1)
	domainsCertificates := DomainsCertificates{Certs: []*DomainsCertificate{
		{
			tlsCert: &tls.Certificate{},
			Domains: types.Domain{
				Main: "*.acme.wtf",
				SANs: []string{"trae.acme.io"},
			},
		},
	}}
	account := Account{DomainsCertificate: domainsCertificates}
	uncheckedDomains = a.getUncheckedDomains(domains, &account)
	assert.Empty(t, uncheckedDomains)
	domains = []string{"traefik.containo.us", "trae.containo.us", "traefik.wtf"}
	uncheckedDomains = a.getUncheckedDomains(domains, nil)
	assert.Len(t, uncheckedDomains, 1)
}

func TestAcme_getProvidedCertificate(t *testing.T) {
	mm := make(map[string]*tls.Certificate)
	mm["*.containo.us"] = &tls.Certificate{}
	mm["traefik.acme.io"] = &tls.Certificate{}

	a := ACME{TLSConfig: &tls.Config{NameToCertificate: mm}}

	domain := "traefik.containo.us"
	certificate := a.getProvidedCertificate(domain)
	assert.NotNil(t, certificate)
	domain = "trae.acme.io"
	certificate = a.getProvidedCertificate(domain)
	assert.Nil(t, certificate)
}

func TestAcme_getValidDomain(t *testing.T) {
	testCases := []struct {
		desc            string
		domains         []string
		wildcardAllowed bool
		dnsChallenge    *acmeprovider.DNSChallenge
		expectedErr     string
		expectedDomains []string
	}{
		{
			desc:            "valid wildcard",
			domains:         []string{"*.traefik.wtf"},
			dnsChallenge:    &acmeprovider.DNSChallenge{},
			wildcardAllowed: true,
			expectedErr:     "",
			expectedDomains: []string{"*.traefik.wtf"},
		},
		{
			desc:            "no wildcard",
			domains:         []string{"traefik.wtf", "foo.traefik.wtf"},
			dnsChallenge:    &acmeprovider.DNSChallenge{},
			expectedErr:     "",
			wildcardAllowed: true,
			expectedDomains: []string{"traefik.wtf", "foo.traefik.wtf"},
		},
		{
			desc:            "unauthorized wildcard",
			domains:         []string{"*.traefik.wtf"},
			dnsChallenge:    &acmeprovider.DNSChallenge{},
			wildcardAllowed: false,
			expectedErr:     "unable to generate a wildcard certificate for domain \"*.traefik.wtf\" from a 'Host' rule",
			expectedDomains: nil,
		},
		{
			desc:            "no domain",
			domains:         []string{},
			dnsChallenge:    nil,
			wildcardAllowed: true,
			expectedErr:     "unable to generate a certificate when no domain is given",
			expectedDomains: nil,
		},
		{
			desc:            "no DNSChallenge",
			domains:         []string{"*.traefik.wtf", "foo.traefik.wtf"},
			dnsChallenge:    nil,
			wildcardAllowed: true,
			expectedErr:     "unable to generate a wildcard certificate for domain \"*.traefik.wtf,foo.traefik.wtf\" : ACME needs a DNSChallenge",
			expectedDomains: nil,
		},
		{
			desc:            "unauthorized wildcard with SAN",
			domains:         []string{"*.*.traefik.wtf", "foo.traefik.wtf"},
			dnsChallenge:    &acmeprovider.DNSChallenge{},
			wildcardAllowed: true,
			expectedErr:     "unable to generate a wildcard certificate for domain \"*.*.traefik.wtf,foo.traefik.wtf\" : ACME does not allow '*.*' wildcard domain",
			expectedDomains: nil,
		},
		{
			desc:            "wildcard with SANs",
			domains:         []string{"*.traefik.wtf", "traefik.wtf"},
			dnsChallenge:    &acmeprovider.DNSChallenge{},
			wildcardAllowed: true,
			expectedErr:     "",
			expectedDomains: []string{"*.traefik.wtf", "traefik.wtf"},
		},
		{
			desc:            "wildcard SANs",
			domains:         []string{"*.traefik.wtf", "*.acme.wtf"},
			dnsChallenge:    &acmeprovider.DNSChallenge{},
			wildcardAllowed: true,
			expectedErr:     "",
			expectedDomains: []string{"*.traefik.wtf", "*.acme.wtf"},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			a := ACME{}
			if test.dnsChallenge != nil {
				a.DNSChallenge = test.dnsChallenge
			}
			domains, err := a.getValidDomains(test.domains, test.wildcardAllowed)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "Unexpected error.")
			} else {
				assert.Equal(t, len(test.expectedDomains), len(domains), "Unexpected domains.")
			}
		})
	}
}

func TestAcme_getCertificateForDomain(t *testing.T) {
	testCases := []struct {
		desc          string
		domain        string
		dc            *DomainsCertificates
		expected      *DomainsCertificate
		expectedFound bool
	}{
		{
			desc:   "non-wildcard exact match",
			domain: "foo.traefik.wtf",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Domains: types.Domain{
							Main: "foo.traefik.wtf",
						},
					},
				},
			},
			expected: &DomainsCertificate{
				Domains: types.Domain{
					Main: "foo.traefik.wtf",
				},
			},
			expectedFound: true,
		},
		{
			desc:   "non-wildcard no match",
			domain: "bar.traefik.wtf",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Domains: types.Domain{
							Main: "foo.traefik.wtf",
						},
					},
				},
			},
			expected:      nil,
			expectedFound: false,
		},
		{
			desc:   "wildcard match",
			domain: "foo.traefik.wtf",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Domains: types.Domain{
							Main: "*.traefik.wtf",
						},
					},
				},
			},
			expected: &DomainsCertificate{
				Domains: types.Domain{
					Main: "*.traefik.wtf",
				},
			},
			expectedFound: true,
		},
		{
			desc:   "wildcard no match",
			domain: "foo.traefik.wtf",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Domains: types.Domain{
							Main: "*.bar.traefik.wtf",
						},
					},
				},
			},
			expected:      nil,
			expectedFound: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got, found := test.dc.getCertificateForDomain(test.domain)
			assert.Equal(t, test.expectedFound, found)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestRemoveEmptyCertificates(t *testing.T) {
	now := time.Now()
	fooCert, fooKey, _ := generate.KeyPair("foo.com", now)
	acmeCert, acmeKey, _ := generate.KeyPair("acme.wtf", now.Add(24*time.Hour))
	barCert, barKey, _ := generate.KeyPair("bar.com", now)
	testCases := []struct {
		desc       string
		dc         *DomainsCertificates
		expectedDc *DomainsCertificates
	}{
		{
			desc: "No empty certificate",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Certificate: &Certificate{
							Certificate: fooCert,
							PrivateKey:  fooKey,
						},
						Domains: types.Domain{
							Main: "foo.com",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: acmeCert,
							PrivateKey:  acmeKey,
						},
						Domains: types.Domain{
							Main: "acme.wtf",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: barCert,
							PrivateKey:  barKey,
						},
						Domains: types.Domain{
							Main: "bar.com",
						},
					},
				},
			},
			expectedDc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Certificate: &Certificate{
							Certificate: fooCert,
							PrivateKey:  fooKey,
						},
						Domains: types.Domain{
							Main: "foo.com",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: acmeCert,
							PrivateKey:  acmeKey,
						},
						Domains: types.Domain{
							Main: "acme.wtf",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: barCert,
							PrivateKey:  barKey,
						},
						Domains: types.Domain{
							Main: "bar.com",
						},
					},
				},
			},
		},
		{
			desc: "First certificate is nil",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Domains: types.Domain{
							Main: "foo.com",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: acmeCert,
							PrivateKey:  acmeKey,
						},
						Domains: types.Domain{
							Main: "acme.wtf",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: barCert,
							PrivateKey:  barKey,
						},
						Domains: types.Domain{
							Main: "bar.com",
						},
					},
				},
			},
			expectedDc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Certificate: &Certificate{
							Certificate: acmeCert,
							PrivateKey:  acmeKey,
						},
						Domains: types.Domain{
							Main: "acme.wtf",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: nil,
							PrivateKey:  barKey,
						},
						Domains: types.Domain{
							Main: "bar.com",
						},
					},
				},
			},
		},
		{
			desc: "Last certificate is empty",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Certificate: &Certificate{
							Certificate: fooCert,
							PrivateKey:  fooKey,
						},
						Domains: types.Domain{
							Main: "foo.com",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: acmeCert,
							PrivateKey:  acmeKey,
						},
						Domains: types.Domain{
							Main: "acme.wtf",
						},
					},
					{
						Certificate: &Certificate{},
						Domains: types.Domain{
							Main: "bar.com",
						},
					},
				},
			},
			expectedDc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Certificate: &Certificate{
							Certificate: fooCert,
							PrivateKey:  fooKey,
						},
						Domains: types.Domain{
							Main: "foo.com",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: acmeCert,
							PrivateKey:  acmeKey,
						},
						Domains: types.Domain{
							Main: "acme.wtf",
						},
					},
				},
			},
		},
		{
			desc: "First and last certificates are nil or empty",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Domains: types.Domain{
							Main: "foo.com",
						},
					},
					{
						Certificate: &Certificate{
							Certificate: acmeCert,
							PrivateKey:  acmeKey,
						},
						Domains: types.Domain{
							Main: "acme.wtf",
						},
					},
					{
						Certificate: &Certificate{},
						Domains: types.Domain{
							Main: "bar.com",
						},
					},
				},
			},
			expectedDc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Certificate: &Certificate{
							Certificate: acmeCert,
							PrivateKey:  acmeKey,
						},
						Domains: types.Domain{
							Main: "acme.wtf",
						},
					},
				},
			},
		},
		{
			desc: "All certificates are nil or empty",
			dc: &DomainsCertificates{
				Certs: []*DomainsCertificate{
					{
						Domains: types.Domain{
							Main: "foo.com",
						},
					},
					{
						Domains: types.Domain{
							Main: "foo24.com",
						},
					},
					{
						Certificate: &Certificate{},
						Domains: types.Domain{
							Main: "bar.com",
						},
					},
				},
			},
			expectedDc: &DomainsCertificates{
				Certs: []*DomainsCertificate{},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			a := &Account{DomainsCertificate: *test.dc}
			a.Init()

			assert.Equal(t, len(test.expectedDc.Certs), len(a.DomainsCertificate.Certs))
			sort.Sort(&a.DomainsCertificate)
			sort.Sort(test.expectedDc)
			for key, value := range test.expectedDc.Certs {
				assert.Equal(t, value.Domains.Main, a.DomainsCertificate.Certs[key].Domains.Main)
			}
		})
	}
}
