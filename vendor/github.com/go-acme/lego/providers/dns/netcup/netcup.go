// Package netcup implements a DNS Provider for solving the DNS-01 challenge using the netcup DNS API.
package netcup

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/providers/dns/netcup/internal"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/log"
	"github.com/go-acme/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Key                string
	Password           string
	Customer           string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("NETCUP_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("NETCUP_PROPAGATION_TIMEOUT", 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond("NETCUP_POLLING_INTERVAL", 5*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("NETCUP_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for netcup.
// Credentials must be passed in the environment variables:
// NETCUP_CUSTOMER_NUMBER, NETCUP_API_KEY, NETCUP_API_PASSWORD
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NETCUP_CUSTOMER_NUMBER", "NETCUP_API_KEY", "NETCUP_API_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("netcup: %v", err)
	}

	config := NewDefaultConfig()
	config.Customer = values["NETCUP_CUSTOMER_NUMBER"]
	config.Key = values["NETCUP_API_KEY"]
	config.Password = values["NETCUP_API_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for netcup.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("netcup: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Customer, config.Key, config.Password)
	if err != nil {
		return nil, fmt.Errorf("netcup: %v", err)
	}

	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domainName, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domainName, keyAuth)

	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("netcup: failed to find DNSZone, %v", err)
	}

	sessionID, err := d.client.Login()
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	defer func() {
		err = d.client.Logout(sessionID)
		if err != nil {
			log.Print("netcup: %v", err)
		}
	}()

	hostname := strings.Replace(fqdn, "."+zone, "", 1)
	record := internal.DNSRecord{
		Hostname:    hostname,
		RecordType:  "TXT",
		Destination: value,
		TTL:         d.config.TTL,
	}

	zone = dns01.UnFqdn(zone)

	records, err := d.client.GetDNSRecords(zone, sessionID)
	if err != nil {
		// skip no existing records
		log.Infof("no existing records, error ignored: %v", err)
	}

	records = append(records, record)

	err = d.client.UpdateDNSRecord(sessionID, zone, records)
	if err != nil {
		return fmt.Errorf("netcup: failed to add TXT-Record: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domainName, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domainName, keyAuth)

	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("netcup: failed to find DNSZone, %v", err)
	}

	sessionID, err := d.client.Login()
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	defer func() {
		err = d.client.Logout(sessionID)
		if err != nil {
			log.Print("netcup: %v", err)
		}
	}()

	hostname := strings.Replace(fqdn, "."+zone, "", 1)

	zone = dns01.UnFqdn(zone)

	records, err := d.client.GetDNSRecords(zone, sessionID)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	record := internal.DNSRecord{
		Hostname:    hostname,
		RecordType:  "TXT",
		Destination: value,
	}

	idx, err := internal.GetDNSRecordIdx(records, record)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	records[idx].DeleteRecord = true

	err = d.client.UpdateDNSRecord(sessionID, zone, []internal.DNSRecord{records[idx]})
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
