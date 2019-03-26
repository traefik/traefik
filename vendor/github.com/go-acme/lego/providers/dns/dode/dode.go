// Package dode implements a DNS provider for solving the DNS-01 challenge using do.de.
package dode

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("DODE_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DODE_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond("DODE_SEQUENCE_INTERVAL", dns01.DefaultPropagationTimeout),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DODE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider adds and removes the record for the DNS challenge
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a new DNS provider using
// environment variable DODE_TOKEN for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DODE_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("do.de: %v", err)
	}

	config := NewDefaultConfig()
	config.Token = values["DODE_TOKEN"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for do.de.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("do.de: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("do.de: credentials missing")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, txtRecord := dns01.GetRecord(domain, keyAuth)
	return d.updateTxtRecord(fqdn, d.config.Token, txtRecord, false)
}

// CleanUp clears TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	return d.updateTxtRecord(fqdn, d.config.Token, "", true)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}
