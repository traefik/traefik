// Package cloudflare implements a DNS provider for solving the DNS-01 challenge using cloudflare DNS.
package cloudflare

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/log"
	"github.com/go-acme/lego/platform/config/env"
)

const (
	minTTL = 120
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	AuthEmail          string
	AuthKey            string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("CLOUDFLARE_TTL", minTTL),
		PropagationTimeout: env.GetOrDefaultSecond("CLOUDFLARE_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("CLOUDFLARE_POLLING_INTERVAL", 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("CLOUDFLARE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	client *cloudflare.API
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Cloudflare.
// Credentials must be passed in the environment variables:
// CLOUDFLARE_EMAIL and CLOUDFLARE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.GetWithFallback(
		[]string{"CLOUDFLARE_EMAIL", "CF_API_EMAIL"},
		[]string{"CLOUDFLARE_API_KEY", "CF_API_KEY"})
	if err != nil {
		return nil, fmt.Errorf("cloudflare: %v", err)
	}

	config := NewDefaultConfig()
	config.AuthEmail = values["CLOUDFLARE_EMAIL"]
	config.AuthKey = values["CLOUDFLARE_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Cloudflare.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("cloudflare: the configuration of the DNS provider is nil")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("cloudflare: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client, err := cloudflare.New(config.AuthKey, config.AuthEmail, cloudflare.HTTPClient(config.HTTPClient))
	if err != nil {
		return nil, err
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("cloudflare: %v", err)
	}

	zoneID, err := d.client.ZoneIDByName(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("cloudflare: failed to find zone %s: %v", authZone, err)
	}

	dnsRecord := cloudflare.DNSRecord{
		Type:    "TXT",
		Name:    dns01.UnFqdn(fqdn),
		Content: value,
		TTL:     d.config.TTL,
	}

	response, err := d.client.CreateDNSRecord(zoneID, dnsRecord)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to create TXT record: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("cloudflare: failed to create TXT record: %+v %+v", response.Errors, response.Messages)
	}

	log.Infof("cloudflare: new record for %s, ID %s", domain, response.Result.ID)

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("cloudflare: %v", err)
	}

	zoneID, err := d.client.ZoneIDByName(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("cloudflare: failed to find zone %s: %v", authZone, err)
	}

	dnsRecord := cloudflare.DNSRecord{
		Type: "TXT",
		Name: dns01.UnFqdn(fqdn),
	}

	records, err := d.client.DNSRecords(zoneID, dnsRecord)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to find TXT records: %v", err)
	}

	for _, record := range records {
		err = d.client.DeleteDNSRecord(zoneID, record.ID)
		if err != nil {
			log.Printf("cloudflare: failed to delete TXT record: %v", err)
		}
	}

	return nil
}
