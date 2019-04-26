// Package cloudns implements a DNS provider for solving the DNS-01 challenge using ClouDNS DNS.
package cloudns

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/go-acme/lego/providers/dns/cloudns/internal"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	AuthID             string
	AuthPassword       string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("CLOUDNS_PROPAGATION_TIMEOUT", 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond("CLOUDNS_POLLING_INTERVAL", 4*time.Second),
		TTL:                env.GetOrDefaultInt("CLOUDNS_TTL", 60),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("CLOUDNS_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for ClouDNS.
// Credentials must be passed in the environment variables:
// CLOUDNS_AUTH_ID and CLOUDNS_AUTH_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("CLOUDNS_AUTH_ID", "CLOUDNS_AUTH_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("ClouDNS: %v", err)
	}

	config := NewDefaultConfig()
	config.AuthID = values["CLOUDNS_AUTH_ID"]
	config.AuthPassword = values["CLOUDNS_AUTH_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ClouDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ClouDNS: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.AuthID, config.AuthPassword)
	if err != nil {
		return nil, fmt.Errorf("ClouDNS: %v", err)
	}

	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := d.client.GetZone(fqdn)
	if err != nil {
		return fmt.Errorf("ClouDNS: %v", err)
	}

	err = d.client.AddTxtRecord(zone.Name, fqdn, value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("ClouDNS: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := d.client.GetZone(fqdn)
	if err != nil {
		return fmt.Errorf("ClouDNS: %v", err)
	}

	record, err := d.client.FindTxtRecord(zone.Name, fqdn)
	if err != nil {
		return fmt.Errorf("ClouDNS: %v", err)
	}

	if record == nil {
		return nil
	}

	err = d.client.RemoveTxtRecord(record.ID, zone.Name)
	if err != nil {
		return fmt.Errorf("ClouDNS: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
