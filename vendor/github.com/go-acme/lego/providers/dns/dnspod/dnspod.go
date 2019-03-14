// Package dnspod implements a DNS provider for solving the DNS-01 challenge using dnspod DNS.
package dnspod

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dnspod "github.com/decker502/dnspod-go"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	LoginToken         string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("DNSPOD_TTL", 600),
		PropagationTimeout: env.GetOrDefaultSecond("DNSPOD_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DNSPOD_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DNSPOD_HTTP_TIMEOUT", 0),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *dnspod.Client
}

// NewDNSProvider returns a DNSProvider instance configured for dnspod.
// Credentials must be passed in the environment variables: DNSPOD_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DNSPOD_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("dnspod: %v", err)
	}

	config := NewDefaultConfig()
	config.LoginToken = values["DNSPOD_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for dnspod.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnspod: the configuration of the DNS provider is nil")
	}

	if config.LoginToken == "" {
		return nil, fmt.Errorf("dnspod: credentials missing")
	}

	params := dnspod.CommonParams{LoginToken: config.LoginToken, Format: "json"}

	client := dnspod.NewClient(params)
	client.HttpClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	zoneID, zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	recordAttributes := d.newTxtRecord(zoneName, fqdn, value, d.config.TTL)
	_, _, err = d.client.Domains.CreateRecord(zoneID, *recordAttributes)
	if err != nil {
		return fmt.Errorf("API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

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

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(domain string) (string, string, error) {
	zones, _, err := d.client.Domains.List()
	if err != nil {
		return "", "", fmt.Errorf("API call failed: %v", err)
	}

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", "", err
	}

	var hostedZone dnspod.Domain
	for _, zone := range zones {
		if zone.Name == dns01.UnFqdn(authZone) {
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
		TTL:   strconv.Itoa(ttl),
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
		return records, fmt.Errorf("API call has failed: %v", err)
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
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
