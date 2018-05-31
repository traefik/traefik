// Package exoscale implements a DNS provider for solving the DNS-01 challenge
// using exoscale DNS.
package exoscale

import (
	"errors"
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *egoscale.Client
}

// NewDNSProvider Credentials must be passed in the environment variables:
// EXOSCALE_API_KEY, EXOSCALE_API_SECRET, EXOSCALE_ENDPOINT.
func NewDNSProvider() (*DNSProvider, error) {
	key := os.Getenv("EXOSCALE_API_KEY")
	secret := os.Getenv("EXOSCALE_API_SECRET")
	endpoint := os.Getenv("EXOSCALE_ENDPOINT")
	return NewDNSProviderClient(key, secret, endpoint)
}

// NewDNSProviderClient Uses the supplied parameters to return a DNSProvider instance
// configured for Exoscale.
func NewDNSProviderClient(key, secret, endpoint string) (*DNSProvider, error) {
	if key == "" || secret == "" {
		return nil, fmt.Errorf("Exoscale credentials missing")
	}
	if endpoint == "" {
		endpoint = "https://api.exoscale.ch/dns"
	}

	return &DNSProvider{
		client: egoscale.NewClient(endpoint, key, secret),
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	zone, recordName, err := c.FindZoneAndRecordName(fqdn, domain)
	if err != nil {
		return err
	}

	recordID, err := c.FindExistingRecordID(zone, recordName)
	if err != nil {
		return err
	}

	record := egoscale.DNSRecord{
		Name:       recordName,
		TTL:        ttl,
		Content:    value,
		RecordType: "TXT",
	}

	if recordID == 0 {
		_, err := c.client.CreateRecord(zone, record)
		if err != nil {
			return errors.New("Error while creating DNS record: " + err.Error())
		}
	} else {
		record.ID = recordID
		_, err := c.client.UpdateRecord(zone, record)
		if err != nil {
			return errors.New("Error while updating DNS record: " + err.Error())
		}
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	zone, recordName, err := c.FindZoneAndRecordName(fqdn, domain)
	if err != nil {
		return err
	}

	recordID, err := c.FindExistingRecordID(zone, recordName)
	if err != nil {
		return err
	}

	if recordID != 0 {
		err = c.client.DeleteRecord(zone, recordID)
		if err != nil {
			return errors.New("Error while deleting DNS record: " + err.Error())
		}
	}

	return nil
}

// FindExistingRecordID Query Exoscale to find an existing record for this name.
// Returns nil if no record could be found
func (c *DNSProvider) FindExistingRecordID(zone, recordName string) (int64, error) {
	records, err := c.client.GetRecords(zone)
	if err != nil {
		return -1, errors.New("Error while retrievening DNS records: " + err.Error())
	}
	for _, record := range records {
		if record.Name == recordName {
			return record.ID, nil
		}
	}
	return 0, nil
}

// FindZoneAndRecordName Extract DNS zone and DNS entry name
func (c *DNSProvider) FindZoneAndRecordName(fqdn, domain string) (string, string, error) {
	zone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", "", err
	}
	zone = acme.UnFqdn(zone)
	name := acme.UnFqdn(fqdn)
	name = name[:len(name)-len("."+zone)]

	return zone, name, nil
}
