// Package cloudxns implements a DNS provider for solving the DNS-01 challenge
// using CloudXNS DNS.
package cloudxns

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
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
	client := acme.HTTPClient
	client.Timeout = time.Second * time.Duration(env.GetOrDefaultInt("CLOUDXNS_HTTP_TIMEOUT", 30))

	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("AKAMAI_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("AKAMAI_POLLING_INTERVAL", acme.DefaultPollingInterval),
		TTL:                env.GetOrDefaultInt("CLOUDXNS_TTL", 120),
		HTTPClient:         &client,
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
	client *Client
}

// NewDNSProvider returns a DNSProvider instance configured for CloudXNS.
// Credentials must be passed in the environment variables:
// CLOUDXNS_API_KEY and CLOUDXNS_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("CLOUDXNS_API_KEY", "CLOUDXNS_SECRET_KEY")
	if err != nil {
		return nil, fmt.Errorf("CloudXNS: %v", err)
	}

	return NewDNSProviderCredentials(values["CLOUDXNS_API_KEY"], values["CLOUDXNS_SECRET_KEY"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for CloudXNS.
func NewDNSProviderCredentials(apiKey, secretKey string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.APIKey = apiKey
	config.SecretKey = secretKey

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for CloudXNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("CloudXNS: the configuration of the DNS provider is nil")
	}

	client, err := NewClient(config.APIKey, config.SecretKey)
	if err != nil {
		return nil, err
	}

	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	info, err := d.client.GetDomainInformation(fqdn)
	if err != nil {
		return err
	}

	return d.client.AddTxtRecord(info, fqdn, value, d.config.TTL)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

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
