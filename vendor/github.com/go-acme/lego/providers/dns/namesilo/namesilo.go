// Package namesilo implements a DNS provider for solving the DNS-01 challenge using namesilo DNS.
package namesilo

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/nrdcg/namesilo"
)

const (
	defaultTTL = 3600
	maxTTL     = 2592000
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("NAMESILO_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("NAMESILO_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		TTL:                env.GetOrDefaultInt("NAMESILO_TTL", defaultTTL),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *namesilo.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for namesilo.
// API_KEY must be passed in the environment variables: NAMESILO_API_KEY.
//
// See: https://www.namesilo.com/api_reference.php
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NAMESILO_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("namesilo: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["NAMESILO_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DNSimple.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("namesilo: the configuration of the DNS provider is nil")
	}

	if config.TTL < defaultTTL || config.TTL > maxTTL {
		return nil, fmt.Errorf("namesilo: TTL should be in [%d, %d]", defaultTTL, maxTTL)
	}

	transport, err := namesilo.NewTokenTransport(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("namesilo: %v", err)
	}

	return &DNSProvider{client: namesilo.NewClient(transport.Client()), config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneName, err := getZoneNameByDomain(domain)
	if err != nil {
		return fmt.Errorf("namesilo: %v", err)
	}

	_, err = d.client.DnsAddRecord(&namesilo.DnsAddRecordParams{
		Domain: zoneName,
		Type:   "TXT",
		Host:   getRecordName(fqdn, zoneName),
		Value:  value,
		TTL:    d.config.TTL,
	})
	if err != nil {
		return fmt.Errorf("namesilo: failed to add record %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zoneName, err := getZoneNameByDomain(domain)
	if err != nil {
		return fmt.Errorf("namesilo: %v", err)
	}

	resp, err := d.client.DnsListRecords(&namesilo.DnsListRecordsParams{Domain: zoneName})
	if err != nil {
		return fmt.Errorf("namesilo: %v", err)
	}

	var lastErr error
	name := getRecordName(fqdn, zoneName)
	for _, r := range resp.Reply.ResourceRecord {
		if r.Type == "TXT" && r.Host == name {
			_, err := d.client.DnsDeleteRecord(&namesilo.DnsDeleteRecordParams{Domain: zoneName, ID: r.RecordID})
			if err != nil {
				lastErr = fmt.Errorf("namesilo: %v", err)
			}
		}
	}
	return lastErr
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func getZoneNameByDomain(domain string) (string, error) {
	zone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", fmt.Errorf("failed to find zone for domain: %s, %v", domain, err)
	}
	return dns01.UnFqdn(zone), nil
}

func getRecordName(domain, zone string) string {
	return strings.TrimSuffix(dns01.ToFqdn(domain), "."+dns01.ToFqdn(zone))
}
