// Package namedotcom implements a DNS provider for solving the DNS-01 challenge
// using Name.com's DNS service.
package namedotcom

import (
	"fmt"
	"os"
	"strings"

	"github.com/namedotcom/go/namecom"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *namecom.NameCom
}

// NewDNSProvider returns a DNSProvider instance configured for namedotcom.
// Credentials must be passed in the environment variables: NAMECOM_USERNAME and NAMECOM_API_TOKEN
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NAMECOM_USERNAME", "NAMECOM_API_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("Name.com: %v", err)
	}

	server := os.Getenv("NAMECOM_SERVER")
	return NewDNSProviderCredentials(values["NAMECOM_USERNAME"], values["NAMECOM_API_TOKEN"], server)
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
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	request := &namecom.Record{
		DomainName: domain,
		Host:       d.extractRecordName(fqdn, domain),
		Type:       "TXT",
		TTL:        uint32(ttl),
		Answer:     value,
	}

	_, err := d.client.CreateRecord(request)
	if err != nil {
		return fmt.Errorf("Name.com API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	records, err := d.getRecords(domain)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if rec.Fqdn == fqdn && rec.Type == "TXT" {
			request := &namecom.DeleteRecordRequest{
				DomainName: domain,
				ID:         rec.ID,
			}
			_, err := d.client.DeleteRecord(request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DNSProvider) getRecords(domain string) ([]*namecom.Record, error) {
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
		response, err = d.client.ListRecords(request)
		if err != nil {
			return nil, err
		}

		records = append(records, response.Records...)
		request.Page = response.NextPage
	}

	return records, nil
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
