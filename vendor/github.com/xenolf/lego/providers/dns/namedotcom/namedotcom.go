// Package namedotcom implements a DNS provider for solving the DNS-01 challenge
// using Name.com's DNS service.
package namedotcom

import (
	"fmt"
	"os"
	"strings"

	"github.com/namedotcom/go/namecom"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *namecom.NameCom
}

// NewDNSProvider returns a DNSProvider instance configured for namedotcom.
// Credentials must be passed in the environment variables: NAMECOM_USERNAME and NAMECOM_API_TOKEN
func NewDNSProvider() (*DNSProvider, error) {
	username := os.Getenv("NAMECOM_USERNAME")
	apiToken := os.Getenv("NAMECOM_API_TOKEN")
	server := os.Getenv("NAMECOM_SERVER")

	return NewDNSProviderCredentials(username, apiToken, server)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for namedotcom.
func NewDNSProviderCredentials(username, apiToken, server string) (*DNSProvider, error) {
	if username == "" {
		return nil, fmt.Errorf("Name.com Username is required")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("Name.com API token is required")
	}

	client := namecom.New(username, apiToken)

	if server != "" {
		client.Server = server
	}

	return &DNSProvider{client: client}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	request := &namecom.Record{
		DomainName: domain,
		Host:       c.extractRecordName(fqdn, domain),
		Type:       "TXT",
		TTL:        uint32(ttl),
		Answer:     value,
	}

	_, err := c.client.CreateRecord(request)
	if err != nil {
		return fmt.Errorf("namedotcom API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	records, err := c.getRecords(domain)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if rec.Fqdn == fqdn && rec.Type == "TXT" {
			request := &namecom.DeleteRecordRequest{
				DomainName: domain,
				ID:         rec.ID,
			}
			_, err := c.client.DeleteRecord(request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *DNSProvider) getRecords(domain string) ([]*namecom.Record, error) {
	var (
		err      error
		records  []*namecom.Record
		response *namecom.ListRecordsResponse
	)

	request := &namecom.ListRecordsRequest{
		DomainName: domain,
		Page:       1,
	}

	for request.Page > 0 {
		response, err = c.client.ListRecords(request)
		if err != nil {
			return nil, err
		}

		records = append(records, response.Records...)
		request.Page = response.NextPage
	}

	return records, nil
}

func (c *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
