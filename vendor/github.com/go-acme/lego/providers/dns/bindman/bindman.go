// Package bindman implements a DNS provider for solving the DNS-01 challenge.
package bindman

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/labbsr0x/bindman-dns-webhook/src/client"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	BaseURL            string
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("BINDMAN_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("BINDMAN_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("BINDMAN_HTTP_TIMEOUT", time.Minute),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// Bindman's Address Manager REST API to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
	client *client.DNSWebhookClient
}

// NewDNSProvider returns a DNSProvider instance configured for Bindman.
// BINDMAN_MANAGER_ADDRESS should have the scheme, hostname, and port (if required) of the authoritative Bindman Manager server.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("BINDMAN_MANAGER_ADDRESS")
	if err != nil {
		return nil, fmt.Errorf("bindman: %v", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values["BINDMAN_MANAGER_ADDRESS"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Bindman.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("bindman: the configuration of the DNS provider is nil")
	}

	if config.BaseURL == "" {
		return nil, fmt.Errorf("bindman: bindman manager address missing")
	}

	bClient, err := client.New(config.BaseURL, config.HTTPClient)
	if err != nil {
		return nil, fmt.Errorf("bindman: %v", err)
	}

	return &DNSProvider{config: config, client: bClient}, nil
}

// Present creates a TXT record using the specified parameters.
// This will *not* create a subzone to contain the TXT record,
// so make sure the FQDN specified is within an extant zone.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	if err := d.client.AddRecord(fqdn, "TXT", value); err != nil {
		return fmt.Errorf("bindman: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	if err := d.client.RemoveRecord(fqdn, "TXT"); err != nil {
		return fmt.Errorf("bindman: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
