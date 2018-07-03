// Package vultr implements a DNS provider for solving the DNS-01 challenge using
// the vultr DNS.
// See https://www.vultr.com/api/#dns
package vultr

import (
	"fmt"
	"strings"

	vultr "github.com/JamesClonk/vultr/lib"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *vultr.Client
}

// NewDNSProvider returns a DNSProvider instance with a configured Vultr client.
// Authentication uses the VULTR_API_KEY environment variable.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("VULTR_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("Vultr: %v", err)
	}

	return NewDNSProviderCredentials(values["VULTR_API_KEY"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a DNSProvider
// instance configured for Vultr.
func NewDNSProviderCredentials(apiKey string) (*DNSProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Vultr credentials missing")
	}

	return &DNSProvider{client: vultr.NewClient(apiKey, nil)}, nil
}

// Present creates a TXT record to fulfil the DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	zoneDomain, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	name := d.extractRecordName(fqdn, zoneDomain)

	err = d.client.CreateDNSRecord(zoneDomain, name, "TXT", `"`+value+`"`, 0, ttl)
	if err != nil {
		return fmt.Errorf("Vultr API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zoneDomain, records, err := d.findTxtRecords(domain, fqdn)
	if err != nil {
		return err
	}

	for _, rec := range records {
		err := d.client.DeleteDNSRecord(zoneDomain, rec.RecordID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	domains, err := d.client.GetDNSDomains()
	if err != nil {
		return "", fmt.Errorf("Vultr API call failed: %v", err)
	}

	var hostedDomain vultr.DNSDomain
	for _, dom := range domains {
		if strings.HasSuffix(domain, dom.Domain) {
			if len(dom.Domain) > len(hostedDomain.Domain) {
				hostedDomain = dom
			}
		}
	}
	if hostedDomain.Domain == "" {
		return "", fmt.Errorf("No matching Vultr domain found for domain %s", domain)
	}

	return hostedDomain.Domain, nil
}

func (d *DNSProvider) findTxtRecords(domain, fqdn string) (string, []vultr.DNSRecord, error) {
	zoneDomain, err := d.getHostedZone(domain)
	if err != nil {
		return "", nil, err
	}

	var records []vultr.DNSRecord
	result, err := d.client.GetDNSRecords(zoneDomain)
	if err != nil {
		return "", records, fmt.Errorf("Vultr API call has failed: %v", err)
	}

	recordName := d.extractRecordName(fqdn, zoneDomain)
	for _, record := range result {
		if record.Type == "TXT" && record.Name == recordName {
			records = append(records, record)
		}
	}

	return zoneDomain, records, nil
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
