// Package digitalocean implements a DNS provider for solving the DNS-01 challenge using digitalocean DNS.
package digitalocean

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	AuthToken          string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		TTL:                env.GetOrDefaultInt("DO_TTL", 30),
		PropagationTimeout: env.GetOrDefaultSecond("DO_PROPAGATION_TIMEOUT", 60*time.Second),
		PollingInterval:    env.GetOrDefaultSecond("DO_POLLING_INTERVAL", 5*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DO_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
// that uses DigitalOcean's REST API to manage TXT records for a domain.
type DNSProvider struct {
	config      *Config
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Digital
// Ocean. Credentials must be passed in the environment variable:
// DO_AUTH_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DO_AUTH_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("digitalocean: %v", err)
	}

	config := NewDefaultConfig()
	config.AuthToken = values["DO_AUTH_TOKEN"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Digital Ocean.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("digitalocean: the configuration of the DNS provider is nil")
	}

	if config.AuthToken == "" {
		return nil, fmt.Errorf("digitalocean: credentials missing")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &DNSProvider{
		config:    config,
		recordIDs: make(map[string]int),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	respData, err := d.addTxtRecord(fqdn, value)
	if err != nil {
		return fmt.Errorf("digitalocean: %v", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[fqdn] = respData.DomainRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("digitalocean: %v", err)
	}

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[fqdn]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("digitalocean: unknown record ID for '%s'", fqdn)
	}

	err = d.removeTxtRecord(authZone, recordID)
	if err != nil {
		return fmt.Errorf("digitalocean: %v", err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, fqdn)
	d.recordIDsMu.Unlock()

	return nil
}
