// Package hostingde implements a DNS provider for solving the DNS-01 challenge using hosting.de.
package hostingde

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
	APIKey             string
	ZoneName           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("HOSTINGDE_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("HOSTINGDE_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("HOSTINGDE_POLLING_INTERVAL", 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("HOSTINGDE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config      *Config
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for hosting.de.
// Credentials must be passed in the environment variables:
// HOSTINGDE_ZONE_NAME and HOSTINGDE_API_KEY
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("HOSTINGDE_API_KEY", "HOSTINGDE_ZONE_NAME")
	if err != nil {
		return nil, fmt.Errorf("hostingde: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["HOSTINGDE_API_KEY"]
	config.ZoneName = values["HOSTINGDE_ZONE_NAME"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for hosting.de.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hostingde: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("hostingde: API key missing")
	}

	if config.ZoneName == "" {
		return nil, errors.New("hostingde: Zone Name missing")
	}

	return &DNSProvider{
		config:    config,
		recordIDs: make(map[string]string),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// get the ZoneConfig for that domain
	zonesFind := ZoneConfigsFindRequest{
		Filter: Filter{
			Field: "zoneName",
			Value: d.config.ZoneName,
		},
		Limit: 1,
		Page:  1,
	}
	zonesFind.AuthToken = d.config.APIKey

	zoneConfig, err := d.getZone(zonesFind)
	if err != nil {
		return fmt.Errorf("hostingde: %v", err)
	}
	zoneConfig.Name = d.config.ZoneName

	rec := []DNSRecord{{
		Type:    "TXT",
		Name:    dns01.UnFqdn(fqdn),
		Content: value,
		TTL:     d.config.TTL,
	}}

	req := ZoneUpdateRequest{
		ZoneConfig:   *zoneConfig,
		RecordsToAdd: rec,
	}
	req.AuthToken = d.config.APIKey

	resp, err := d.updateZone(req)
	if err != nil {
		return fmt.Errorf("hostingde: %v", err)
	}

	for _, record := range resp.Response.Records {
		if record.Name == dns01.UnFqdn(fqdn) && record.Content == fmt.Sprintf(`"%s"`, value) {
			d.recordIDsMu.Lock()
			d.recordIDs[fqdn] = record.ID
			d.recordIDsMu.Unlock()
		}
	}

	if d.recordIDs[fqdn] == "" {
		return fmt.Errorf("hostingde: error getting ID of just created record, for domain %s", domain)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	rec := []DNSRecord{{
		Type:    "TXT",
		Name:    dns01.UnFqdn(fqdn),
		Content: `"` + value + `"`,
	}}

	// get the ZoneConfig for that domain
	zonesFind := ZoneConfigsFindRequest{
		Filter: Filter{
			Field: "zoneName",
			Value: d.config.ZoneName,
		},
		Limit: 1,
		Page:  1,
	}
	zonesFind.AuthToken = d.config.APIKey

	zoneConfig, err := d.getZone(zonesFind)
	if err != nil {
		return fmt.Errorf("hostingde: %v", err)
	}
	zoneConfig.Name = d.config.ZoneName

	req := ZoneUpdateRequest{
		ZoneConfig:      *zoneConfig,
		RecordsToDelete: rec,
	}
	req.AuthToken = d.config.APIKey

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, fqdn)
	d.recordIDsMu.Unlock()

	_, err = d.updateZone(req)
	if err != nil {
		return fmt.Errorf("hostingde: %v", err)
	}
	return nil
}
