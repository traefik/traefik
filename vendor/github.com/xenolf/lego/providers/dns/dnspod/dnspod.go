// Package dnspod implements a DNS provider for solving the DNS-01 challenge
// using dnspod DNS.
package dnspod

import (
	"fmt"
	"strings"

	"github.com/decker502/dnspod-go"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *dnspod.Client
}

// NewDNSProvider returns a DNSProvider instance configured for dnspod.
// Credentials must be passed in the environment variables: DNSPOD_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DNSPOD_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("DNSPod: %v", err)
	}

	return NewDNSProviderCredentials(values["DNSPOD_API_KEY"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for dnspod.
func NewDNSProviderCredentials(key string) (*DNSProvider, error) {
	if key == "" {
		return nil, fmt.Errorf("dnspod credentials missing")
	}

	params := dnspod.CommonParams{LoginToken: key, Format: "json"}
	return &DNSProvider{
		client: dnspod.NewClient(params),
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	zoneID, zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	recordAttributes := d.newTxtRecord(zoneName, fqdn, value, ttl)
	_, _, err = d.client.Domains.CreateRecord(zoneID, *recordAttributes)
	if err != nil {
		return fmt.Errorf("dnspod API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	records, err := d.findTxtRecords(domain, fqdn)
	if err != nil {
		return err
	}

	zoneID, _, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	for _, rec := range records {
		_, err := d.client.Domains.DeleteRecord(zoneID, rec.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (string, string, error) {
	zones, _, err := d.client.Domains.List()
	if err != nil {
		return "", "", fmt.Errorf("dnspod API call failed: %v", err)
	}

	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", "", err
	}

	var hostedZone dnspod.Domain
	for _, zone := range zones {
		if zone.Name == acme.UnFqdn(authZone) {
			hostedZone = zone
		}
	}

	if hostedZone.ID == 0 {
		return "", "", fmt.Errorf("zone %s not found in dnspod for domain %s", authZone, domain)

	}

	return fmt.Sprintf("%v", hostedZone.ID), hostedZone.Name, nil
}

func (d *DNSProvider) newTxtRecord(zone, fqdn, value string, ttl int) *dnspod.Record {
	name := d.extractRecordName(fqdn, zone)

	return &dnspod.Record{
		Type:  "TXT",
		Name:  name,
		Value: value,
		Line:  "默认",
		TTL:   "600",
	}
}

func (d *DNSProvider) findTxtRecords(domain, fqdn string) ([]dnspod.Record, error) {
	zoneID, zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return nil, err
	}

	var records []dnspod.Record
	result, _, err := d.client.Domains.ListRecords(zoneID, "")
	if err != nil {
		return records, fmt.Errorf("dnspod API call has failed: %v", err)
	}

	recordName := d.extractRecordName(fqdn, zoneName)

	for _, record := range result {
		if record.Name == recordName {
			records = append(records, record)
		}
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
