// Package cloudxns implements a DNS provider for solving the DNS-01 challenge using CloudXNS DNS.
package cloudxns

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/go-acme/lego/providers/dns/cloudxns/internal"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	SecretKey          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("CLOUDXNS_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("CLOUDXNS_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		TTL:                env.GetOrDefaultInt("CLOUDXNS_TTL", dns01.DefaultTTL),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("CLOUDXNS_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for CloudXNS.
// Credentials must be passed in the environment variables:
// CLOUDXNS_API_KEY and CLOUDXNS_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("CLOUDXNS_API_KEY", "CLOUDXNS_SECRET_KEY")
	if err != nil {
		return nil, fmt.Errorf("CloudXNS: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["CLOUDXNS_API_KEY"]
	config.SecretKey = values["CLOUDXNS_SECRET_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for CloudXNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("CloudXNS: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIKey, config.SecretKey)
	if err != nil {
		return nil, err
	}

	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	info, err := d.client.GetDomainInformation(fqdn)
	if err != nil {
		return err
	}

	return d.client.AddTxtRecord(info, fqdn, value, d.config.TTL)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	info, err := d.client.GetDomainInformation(fqdn)
	if err != nil {
		return err
	}

	record, err := d.client.FindTxtRecord(info.ID, fqdn)
	if err != nil {
		return err
	}

	return d.client.RemoveTxtRecord(record.RecordID, info.ID)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
