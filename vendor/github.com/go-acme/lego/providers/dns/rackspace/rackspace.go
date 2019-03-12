// Package rackspace implements a DNS provider for solving the DNS-01 challenge using rackspace DNS.
package rackspace

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

// defaultBaseURL represents the Identity API endpoint to call
const defaultBaseURL = "https://identity.api.rackspacecloud.com/v2.0/tokens"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIUser            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		TTL:                env.GetOrDefaultInt("RACKSPACE_TTL", 300),
		PropagationTimeout: env.GetOrDefaultSecond("RACKSPACE_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("RACKSPACE_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("RACKSPACE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
// used to store the reusable token and DNS API endpoint
type DNSProvider struct {
	config           *Config
	token            string
	cloudDNSEndpoint string
}

// NewDNSProvider returns a DNSProvider instance configured for Rackspace.
// Credentials must be passed in the environment variables:
// RACKSPACE_USER and RACKSPACE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("RACKSPACE_USER", "RACKSPACE_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("rackspace: %v", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values["RACKSPACE_USER"]
	config.APIKey = values["RACKSPACE_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Rackspace.
// It authenticates against the API, also grabbing the DNS Endpoint.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rackspace: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIKey == "" {
		return nil, fmt.Errorf("rackspace: credentials missing")
	}

	identity, err := login(config)
	if err != nil {
		return nil, fmt.Errorf("rackspace: %v", err)
	}

	// Iterate through the Service Catalog to get the DNS Endpoint
	var dnsEndpoint string
	for _, service := range identity.Access.ServiceCatalog {
		if service.Name == "cloudDNS" {
			dnsEndpoint = service.Endpoints[0].PublicURL
			break
		}
	}

	if dnsEndpoint == "" {
		return nil, fmt.Errorf("rackspace: failed to populate DNS endpoint, check Rackspace API for changes")
	}

	return &DNSProvider{
		config:           config,
		token:            identity.Access.Token.ID,
		cloudDNSEndpoint: dnsEndpoint,
	}, nil

}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}

	rec := Records{
		Record: []Record{{
			Name: dns01.UnFqdn(fqdn),
			Type: "TXT",
			Data: value,
			TTL:  d.config.TTL,
		}},
	}

	body, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}

	_, err = d.makeRequest(http.MethodPost, fmt.Sprintf("/domains/%d/records", zoneID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}

	record, err := d.findTxtRecord(fqdn, zoneID)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}

	_, err = d.makeRequest(http.MethodDelete, fmt.Sprintf("/domains/%d/records?id=%s", zoneID, record.ID), nil)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
