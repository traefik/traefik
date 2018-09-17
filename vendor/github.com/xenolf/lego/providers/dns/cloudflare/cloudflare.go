// Package cloudflare implements a DNS provider for solving the DNS-01
// challenge using cloudflare DNS.
package cloudflare

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// CloudFlareAPIURL represents the API endpoint to call.
const CloudFlareAPIURL = defaultBaseURL // Deprecated

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
		TTL:                env.GetOrDefaultInt("CLOUDFLARE_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("CLOUDFLARE_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("CLOUDFLARE_POLLING_INTERVAL", 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("CLOUDFLARE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	client *Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Cloudflare.
// Credentials must be passed in the environment variables:
// CLOUDFLARE_EMAIL and CLOUDFLARE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("CLOUDFLARE_EMAIL", "CLOUDFLARE_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("cloudflare: %v", err)
	}

	config := NewDefaultConfig()
	config.AuthEmail = values["CLOUDFLARE_EMAIL"]
	config.AuthKey = values["CLOUDFLARE_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Cloudflare.
// Deprecated
func NewDNSProviderCredentials(email, key string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.AuthEmail = email
	config.AuthKey = key

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Cloudflare.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("cloudflare: the configuration of the DNS provider is nil")
	}

	client, err := NewClient(config.AuthEmail, config.AuthKey)
	if err != nil {
		return nil, err
	}

	client.HTTPClient = config.HTTPClient

	// TODO: must be remove. keep only for compatibility reason.
	client.BaseURL = CloudFlareAPIURL

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	rec := TxtRecord{
		Type:    "TXT",
		Name:    acme.UnFqdn(fqdn),
		Content: value,
		TTL:     d.config.TTL,
	}

	return d.client.AddTxtRecord(fqdn, rec)
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	return d.client.RemoveTxtRecord(fqdn)
}
