// Package exoscale implements a DNS provider for solving the DNS-01 challenge
// using exoscale DNS.
package exoscale

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

const defaultBaseURL = "https://api.exoscale.ch/dns"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	APISecret          string
	Endpoint           string
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("EXOSCALE_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("EXOSCALE_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("EXOSCALE_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("EXOSCALE_HTTP_TIMEOUT", 0),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *egoscale.Client
}

// NewDNSProvider Credentials must be passed in the environment variables:
// EXOSCALE_API_KEY, EXOSCALE_API_SECRET, EXOSCALE_ENDPOINT.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("EXOSCALE_API_KEY", "EXOSCALE_API_SECRET")
	if err != nil {
		return nil, fmt.Errorf("exoscale: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["EXOSCALE_API_KEY"]
	config.APISecret = values["EXOSCALE_API_SECRET"]
	config.Endpoint = os.Getenv("EXOSCALE_ENDPOINT")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderClient Uses the supplied parameters
// to return a DNSProvider instance configured for Exoscale.
// Deprecated
func NewDNSProviderClient(key, secret, endpoint string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.APIKey = key
	config.APISecret = secret
	config.Endpoint = endpoint

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Exoscale.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" || config.APISecret == "" {
		return nil, fmt.Errorf("exoscale: credentials missing")
	}

	if config.Endpoint == "" {
		config.Endpoint = defaultBaseURL
	}

	client := egoscale.NewClient(config.Endpoint, config.APIKey, config.APISecret)
	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, recordName, err := d.FindZoneAndRecordName(fqdn, domain)
	if err != nil {
		return err
	}

	recordID, err := d.FindExistingRecordID(zone, recordName)
	if err != nil {
		return err
	}

	if recordID == 0 {
		record := egoscale.DNSRecord{
			Name:       recordName,
			TTL:        d.config.TTL,
			Content:    value,
			RecordType: "TXT",
		}

		_, err := d.client.CreateRecord(zone, record)
		if err != nil {
			return errors.New("Error while creating DNS record: " + err.Error())
		}
	} else {
		record := egoscale.UpdateDNSRecord{
			ID:         recordID,
			Name:       recordName,
			TTL:        d.config.TTL,
			Content:    value,
			RecordType: "TXT",
		}

		_, err := d.client.UpdateRecord(zone, record)
		if err != nil {
			return errors.New("Error while updating DNS record: " + err.Error())
		}
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	zone, recordName, err := d.FindZoneAndRecordName(fqdn, domain)
	if err != nil {
		return err
	}

	recordID, err := d.FindExistingRecordID(zone, recordName)
	if err != nil {
		return err
	}

	if recordID != 0 {
		err = d.client.DeleteRecord(zone, recordID)
		if err != nil {
			return errors.New("Error while deleting DNS record: " + err.Error())
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// FindExistingRecordID Query Exoscale to find an existing record for this name.
// Returns nil if no record could be found
func (d *DNSProvider) FindExistingRecordID(zone, recordName string) (int64, error) {
	records, err := d.client.GetRecords(zone)
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
func (d *DNSProvider) FindZoneAndRecordName(fqdn, domain string) (string, string, error) {
	zone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", "", err
	}
	zone = acme.UnFqdn(zone)
	name := acme.UnFqdn(fqdn)
	name = name[:len(name)-len("."+zone)]

	return zone, name, nil
}
