// Package dyn implements a DNS provider for solving the DNS-01 challenge using Dyn Managed DNS.
package dyn

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	CustomerName       string
	UserName           string
	Password           string
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("DYN_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("DYN_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DYN_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DYN_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// Dyn's Managed DNS API to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
	token  string
}

// NewDNSProvider returns a DNSProvider instance configured for Dyn DNS.
// Credentials must be passed in the environment variables:
// DYN_CUSTOMER_NAME, DYN_USER_NAME and DYN_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DYN_CUSTOMER_NAME", "DYN_USER_NAME", "DYN_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("dyn: %v", err)
	}

	config := NewDefaultConfig()
	config.CustomerName = values["DYN_CUSTOMER_NAME"]
	config.UserName = values["DYN_USER_NAME"]
	config.Password = values["DYN_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Dyn DNS
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dyn: the configuration of the DNS provider is nil")
	}

	if config.CustomerName == "" || config.UserName == "" || config.Password == "" {
		return nil, fmt.Errorf("dyn: credentials missing")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	err = d.login()
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	data := map[string]interface{}{
		"rdata": map[string]string{
			"txtdata": value,
		},
		"ttl": strconv.Itoa(d.config.TTL),
	}

	resource := fmt.Sprintf("TXTRecord/%s/%s/", authZone, fqdn)
	_, err = d.sendRequest(http.MethodPost, resource, data)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	err = d.publish(authZone, "Added TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	return d.logout()
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	err = d.login()
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	resource := fmt.Sprintf("TXTRecord/%s/%s/", authZone, fqdn)
	url := fmt.Sprintf("%s/%s", defaultBaseURL, resource)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", d.token)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dyn: API request failed to delete TXT record HTTP status code %d", resp.StatusCode)
	}

	err = d.publish(authZone, "Removed TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	return d.logout()
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
