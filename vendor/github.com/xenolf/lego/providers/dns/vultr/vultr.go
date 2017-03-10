// Package vultr implements a DNS provider for solving the DNS-01 challenge using
// the vultr DNS.
// See https://www.vultr.com/api/#dns
package vultr

import (
	"fmt"
	"os"
	"strings"

	vultr "github.com/JamesClonk/vultr/lib"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *vultr.Client
}

// NewDNSProvider returns a DNSProvider instance with a configured Vultr client.
// Authentication uses the VULTR_API_KEY environment variable.
func NewDNSProvider() (*DNSProvider, error) {
	apiKey := os.Getenv("VULTR_API_KEY")
	return NewDNSProviderCredentials(apiKey)
}

// NewDNSProviderCredentials uses the supplied credentials to return a DNSProvider
// instance configured for Vultr.
func NewDNSProviderCredentials(apiKey string) (*DNSProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Vultr credentials missing")
	}

	c := &DNSProvider{
		client: vultr.NewClient(apiKey, nil),
	}

	return c, nil
}

// Present creates a TXT record to fulfil the DNS-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	zoneDomain, err := c.getHostedZone(domain)
	if err != nil {
		return err
	}

	name := c.extractRecordName(fqdn, zoneDomain)

	err = c.client.CreateDNSRecord(zoneDomain, name, "TXT", `"`+value+`"`, 0, ttl)
	if err != nil {
		return fmt.Errorf("Vultr API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zoneDomain, records, err := c.findTxtRecords(domain, fqdn)
	if err != nil {
		return err
	}

	for _, rec := range records {
		err := c.client.DeleteDNSRecord(zoneDomain, rec.RecordID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *DNSProvider) getHostedZone(domain string) (string, error) {
	domains, err := c.client.GetDNSDomains()
	if err != nil {
		return "", fmt.Errorf("Vultr API call failed: %v", err)
	}

	var hostedDomain vultr.DNSDomain
	for _, d := range domains {
		if strings.HasSuffix(domain, d.Domain) {
			if len(d.Domain) > len(hostedDomain.Domain) {
				hostedDomain = d
			}
		}
	}
	if hostedDomain.Domain == "" {
		return "", fmt.Errorf("No matching Vultr domain found for domain %s", domain)
	}

	return hostedDomain.Domain, nil
}

func (c *DNSProvider) findTxtRecords(domain, fqdn string) (string, []vultr.DNSRecord, error) {
	zoneDomain, err := c.getHostedZone(domain)
	if err != nil {
		return "", nil, err
	}

	var records []vultr.DNSRecord
	result, err := c.client.GetDNSRecords(zoneDomain)
	if err != nil {
		return "", records, fmt.Errorf("Vultr API call has failed: %v", err)
	}

	recordName := c.extractRecordName(fqdn, zoneDomain)
	for _, record := range result {
		if record.Type == "TXT" && record.Name == recordName {
			records = append(records, record)
		}
	}

	return zoneDomain, records, nil
}

func (c *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
