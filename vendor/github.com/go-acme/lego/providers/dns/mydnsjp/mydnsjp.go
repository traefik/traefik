// Package mydnsjp implements a DNS provider for solving the DNS-01 challenge using MyDNS.jp.
package mydnsjp

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

const defaultBaseURL = "https://www.mydns.jp/directedit.html"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	MasterID           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("MYDNSJP_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("MYDNSJP_POLLING_INTERVAL", 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("MYDNSJP_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for MyDNS.jp.
// Credentials must be passed in the environment variables: MYDNSJP_MASTER_ID and MYDNSJP_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("MYDNSJP_MASTER_ID", "MYDNSJP_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("mydnsjp: %v", err)
	}

	config := NewDefaultConfig()
	config.MasterID = values["MYDNSJP_MASTER_ID"]
	config.Password = values["MYDNSJP_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for MyDNS.jp.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("mydnsjp: the configuration of the DNS provider is nil")
	}

	if config.MasterID == "" || config.Password == "" {
		return nil, errors.New("mydnsjp: some credentials information are missing")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	_, value := dns01.GetRecord(domain, keyAuth)
	err := d.doRequest(domain, value, "REGIST")
	if err != nil {
		return fmt.Errorf("mydnsjp: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	_, value := dns01.GetRecord(domain, keyAuth)
	err := d.doRequest(domain, value, "DELETE")
	if err != nil {
		return fmt.Errorf("mydnsjp: %v", err)
	}
	return nil
}
