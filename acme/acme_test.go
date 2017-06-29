package acme

import (
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xenolf/lego/acme"
)

func TestDomainsSet(t *testing.T) {
	checkMap := map[string]Domains{
		"":                                   {},
		"foo.com":                            {Domain{Main: "foo.com", SANs: []string{}}},
		"foo.com,bar.net":                    {Domain{Main: "foo.com", SANs: []string{"bar.net"}}},
		"foo.com,bar1.net,bar2.net,bar3.net": {Domain{Main: "foo.com", SANs: []string{"bar1.net", "bar2.net", "bar3.net"}}},
	}
	for in, check := range checkMap {
		ds := Domains{}
		ds.Set(in)
		if !reflect.DeepEqual(check, ds) {
			t.Errorf("Expected %+v\nGot %+v", check, ds)
		}
	}
}

func TestDomainsSetAppend(t *testing.T) {
	inSlice := []string{
		"",
		"foo1.com",
		"foo2.com,bar.net",
		"foo3.com,bar1.net,bar2.net,bar3.net",
	}
	checkSlice := []Domains{
		{},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}}},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}},
			Domain{
				Main: "foo2.com",
				SANs: []string{"bar.net"}}},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}},
			Domain{
				Main: "foo2.com",
				SANs: []string{"bar.net"}},
			Domain{Main: "foo3.com",
				SANs: []string{"bar1.net", "bar2.net", "bar3.net"}}},
	}
	ds := Domains{}
	for i, in := range inSlice {
		ds.Set(in)
		if !reflect.DeepEqual(checkSlice[i], ds) {
			t.Errorf("Expected  %s %+v\nGot %+v", in, checkSlice[i], ds)
		}
	}
}

func TestCertificatesRenew(t *testing.T) {
	foo1Cert, foo1Key, _ := generateKeyPair("foo1.com", time.Now())
	foo2Cert, foo2Key, _ := generateKeyPair("foo2.com", time.Now())
	domainsCertificates := DomainsCertificates{
		lock: sync.RWMutex{},
		Certs: []*DomainsCertificate{
			{
				Domains: Domain{
					Main: "foo1.com",
					SANs: []string{}},
				Certificate: &Certificate{
					Domain:        "foo1.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    foo1Key,
					Certificate:   foo1Cert,
				},
			},
			{
				Domains: Domain{
					Main: "foo2.com",
					SANs: []string{}},
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
	foo1Cert, foo1Key, _ = generateKeyPair("foo1.com", time.Now())
	newCertificate := &Certificate{
		Domain:        "foo1.com",
		CertURL:       "url",
		CertStableURL: "url",
		PrivateKey:    foo1Key,
		Certificate:   foo1Cert,
	}

	err := domainsCertificates.renewCertificates(
		newCertificate,
		Domain{
			Main: "foo1.com",
			SANs: []string{}})
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
	fooCert, fooKey, _ := generateKeyPair("foo.com", now)
	foo24Cert, foo24Key, _ := generateKeyPair("foo.com", now.Add(24*time.Hour))
	foo48Cert, foo48Key, _ := generateKeyPair("foo.com", now.Add(48*time.Hour))
	barCert, barKey, _ := generateKeyPair("bar.com", now)
	domainsCertificates := DomainsCertificates{
		lock: sync.RWMutex{},
		Certs: []*DomainsCertificate{
			{
				Domains: Domain{
					Main: "foo.com",
					SANs: []string{}},
				Certificate: &Certificate{
					Domain:        "foo.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    foo24Key,
					Certificate:   foo24Cert,
				},
			},
			{
				Domains: Domain{
					Main: "foo.com",
					SANs: []string{}},
				Certificate: &Certificate{
					Domain:        "foo.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    foo48Key,
					Certificate:   foo48Cert,
				},
			},
			{
				Domains: Domain{
					Main: "foo.com",
					SANs: []string{}},
				Certificate: &Certificate{
					Domain:        "foo.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    fooKey,
					Certificate:   fooCert,
				},
			},
			{
				Domains: Domain{
					Main: "bar.com",
					SANs: []string{}},
				Certificate: &Certificate{
					Domain:        "bar.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey:    barKey,
					Certificate:   barCert,
				},
			},
			{
				Domains: Domain{
					Main: "foo.com",
					SANs: []string{}},
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

func TestNoPreCheckOverride(t *testing.T) {
	acme.PreCheckDNS = nil // Irreversable - but not expecting real calls into this during testing process
	err := dnsOverrideDelay(0)
	if err != nil {
		t.Errorf("Error in dnsOverrideDelay :%v", err)
	}
	if acme.PreCheckDNS != nil {
		t.Error("Unexpected change to acme.PreCheckDNS when leaving DNS verification as is.")
	}
}

func TestSillyPreCheckOverride(t *testing.T) {
	err := dnsOverrideDelay(-5)
	if err == nil {
		t.Error("Missing expected error in dnsOverrideDelay!")
	}
}

func TestPreCheckOverride(t *testing.T) {
	acme.PreCheckDNS = nil // Irreversable - but not expecting real calls into this during testing process
	err := dnsOverrideDelay(5)
	if err != nil {
		t.Errorf("Error in dnsOverrideDelay :%v", err)
	}
	if acme.PreCheckDNS == nil {
		t.Error("No change to acme.PreCheckDNS when meant to be adding enforcing override function.")
	}
}

func TestAcmeClientCreation(t *testing.T) {
	acme.PreCheckDNS = nil // Irreversable - but not expecting real calls into this during testing process
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
		w.Write([]byte(`{
"new-authz": "https://foo/acme/new-authz",
"new-cert": "https://foo/acme/new-cert",
"new-reg": "https://foo/acme/new-reg",
"revoke-cert": "https://foo/acme/revoke-cert"
}`))
	}))
	defer ts.Close()
	a := ACME{DNSProvider: "manual", DelayDontCheckDNS: 10, CAServer: ts.URL}

	client, err := a.buildACMEClient(account)
	if err != nil {
		t.Errorf("Error in buildACMEClient: %v", err)
	}
	if client == nil {
		t.Error("No client from buildACMEClient!")
	}
	if acme.PreCheckDNS == nil {
		t.Error("No change to acme.PreCheckDNS when meant to be adding enforcing override function.")
	}
}

func TestAcme_getProvidedCertificate(t *testing.T) {
	mm := make(map[string]*tls.Certificate)
	mm["*.containo.us"] = &tls.Certificate{}
	mm["traefik.acme.io"] = &tls.Certificate{}

	a := ACME{TLSConfig: &tls.Config{NameToCertificate: mm}}

	domains := []string{"traefik.containo.us", "trae.containo.us"}
	certificate := a.getProvidedCertificate(domains)
	assert.NotNil(t, certificate)
	domains = []string{"traefik.acme.io", "trae.acme.io"}
	certificate = a.getProvidedCertificate(domains)
	assert.Nil(t, certificate)
}
