// Package ns1 implements a DNS provider for solving the DNS-01 challenge
// using NS1 DNS.
package ns1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
	"gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *rest.Client
}

// NewDNSProvider returns a DNSProvider instance configured for NS1.
// Credentials must be passed in the environment variables: NS1_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NS1_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("NS1: %v", err)
	}

	return NewDNSProviderCredentials(values["NS1_API_KEY"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for NS1.
func NewDNSProviderCredentials(key string) (*DNSProvider, error) {
	if key == "" {
		return nil, fmt.Errorf("NS1 credentials missing")
	}

	httpClient := &http.Client{Timeout: time.Second * 10}
	client := rest.NewClient(httpClient, rest.SetAPIKey(key))

	return &DNSProvider{client}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	record := d.newTxtRecord(zone, fqdn, value, ttl)
	_, err = d.client.Records.Create(record)
	if err != nil && err != rest.ErrRecordExists {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	name := acme.UnFqdn(fqdn)
	_, err = d.client.Records.Delete(zone.Zone, name, "TXT")
	return err
}

func (d *DNSProvider) getHostedZone(domain string) (*dns.Zone, error) {
	authZone, err := getAuthZone(domain)
	if err != nil {
		return nil, err
	}

	zone, _, err := d.client.Zones.Get(authZone)
	if err != nil {
		return nil, err
	}

	return zone, nil
}

func getAuthZone(fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(authZone, ".") {
		authZone = authZone[:len(authZone)-len(".")]
	}

	return authZone, err
}

func (d *DNSProvider) newTxtRecord(zone *dns.Zone, fqdn, value string, ttl int) *dns.Record {
	name := acme.UnFqdn(fqdn)

	return &dns.Record{
		Type:   "TXT",
		Zone:   zone.Zone,
		Domain: name,
		TTL:    ttl,
		Answers: []*dns.Answer{
			{Rdata: []string{value}},
		},
	}
}
