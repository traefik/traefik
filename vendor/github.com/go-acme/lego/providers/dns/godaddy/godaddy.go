// Package godaddy implements a DNS provider for solving the DNS-01 challenge using godaddy DNS.
package godaddy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

const (
	// defaultBaseURL represents the API endpoint to call.
	defaultBaseURL = "https://api.godaddy.com"
	minTTL         = 600
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	APISecret          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("GODADDY_TTL", minTTL),
		PropagationTimeout: env.GetOrDefaultSecond("GODADDY_PROPAGATION_TIMEOUT", 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond("GODADDY_POLLING_INTERVAL", 2*time.Second),
		SequenceInterval:   env.GetOrDefaultSecond("GODADDY_SEQUENCE_INTERVAL", dns01.DefaultPropagationTimeout),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("GODADDY_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for godaddy.
// Credentials must be passed in the environment variables:
// GODADDY_API_KEY and GODADDY_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("GODADDY_API_KEY", "GODADDY_API_SECRET")
	if err != nil {
		return nil, fmt.Errorf("godaddy: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["GODADDY_API_KEY"]
	config.APISecret = values["GODADDY_API_SECRET"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for godaddy.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("godaddy: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" || config.APISecret == "" {
		return nil, fmt.Errorf("godaddy: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("godaddy: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	domainZone, err := d.getZone(fqdn)
	if err != nil {
		return err
	}

	recordName := d.extractRecordName(fqdn, domainZone)
	rec := []DNSRecord{
		{
			Type: "TXT",
			Name: recordName,
			Data: value,
			TTL:  d.config.TTL,
		},
	}

	return d.updateRecords(rec, domainZone, recordName)
}

// CleanUp sets null value in the TXT DNS record as GoDaddy has no proper DELETE record method
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	domainZone, err := d.getZone(fqdn)
	if err != nil {
		return err
	}

	recordName := d.extractRecordName(fqdn, domainZone)
	rec := []DNSRecord{
		{
			Type: "TXT",
			Name: recordName,
			Data: "null",
		},
	}

	return d.updateRecords(rec, domainZone, recordName)
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}

func (d *DNSProvider) getZone(fqdn string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
	}

	return dns01.UnFqdn(authZone), nil
}
