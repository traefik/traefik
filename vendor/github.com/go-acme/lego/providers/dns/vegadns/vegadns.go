// Package vegadns implements a DNS provider for solving the DNS-01 challenge using VegaDNS.
package vegadns

import (
	"errors"
	"fmt"
	"strings"
	"time"

	vegaClient "github.com/OpenDNS/vegadns2client"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIKey             string
	APISecret          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("VEGADNS_TTL", 10),
		PropagationTimeout: env.GetOrDefaultSecond("VEGADNS_PROPAGATION_TIMEOUT", 12*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("VEGADNS_POLLING_INTERVAL", 1*time.Minute),
	}
}

// DNSProvider describes a provider for VegaDNS
type DNSProvider struct {
	config *Config
	client vegaClient.VegaDNSClient
}

// NewDNSProvider returns a DNSProvider instance configured for VegaDNS.
// Credentials must be passed in the environment variables:
// VEGADNS_URL, SECRET_VEGADNS_KEY, SECRET_VEGADNS_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("VEGADNS_URL")
	if err != nil {
		return nil, fmt.Errorf("vegadns: %v", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values["VEGADNS_URL"]
	config.APIKey = env.GetOrFile("SECRET_VEGADNS_KEY")
	config.APISecret = env.GetOrFile("SECRET_VEGADNS_SECRET")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for VegaDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vegadns: the configuration of the DNS provider is nil")
	}

	vega := vegaClient.NewVegaDNSClient(config.BaseURL)
	vega.APIKey = config.APIKey
	vega.APISecret = config.APISecret

	return &DNSProvider{client: vega, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	_, domainID, err := d.client.GetAuthZone(fqdn)
	if err != nil {
		return fmt.Errorf("vegadns: can't find Authoritative Zone for %s in Present: %v", fqdn, err)
	}

	err = d.client.CreateTXT(domainID, fqdn, value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("vegadns: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	_, domainID, err := d.client.GetAuthZone(fqdn)
	if err != nil {
		return fmt.Errorf("vegadns: can't find Authoritative Zone for %s in CleanUp: %v", fqdn, err)
	}

	txt := strings.TrimSuffix(fqdn, ".")

	recordID, err := d.client.GetRecordID(domainID, txt, "TXT")
	if err != nil {
		return fmt.Errorf("vegadns: couldn't get Record ID in CleanUp: %s", err)
	}

	err = d.client.DeleteRecord(recordID)
	if err != nil {
		return fmt.Errorf("vegadns: %v", err)
	}
	return nil
}
