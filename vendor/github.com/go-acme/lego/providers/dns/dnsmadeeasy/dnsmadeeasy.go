// Package dnsmadeeasy implements a DNS provider for solving the DNS-01 challenge using DNS Made Easy.
package dnsmadeeasy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/go-acme/lego/providers/dns/dnsmadeeasy/internal"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIKey             string
	APISecret          string
	Sandbox            bool
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("DNSMADEEASY_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("DNSMADEEASY_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DNSMADEEASY_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DNSMADEEASY_HTTP_TIMEOUT", 10*time.Second),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// DNSMadeEasy's DNS API to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for DNSMadeEasy DNS.
// Credentials must be passed in the environment variables:
// DNSMADEEASY_API_KEY and DNSMADEEASY_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DNSMADEEASY_API_KEY", "DNSMADEEASY_API_SECRET")
	if err != nil {
		return nil, fmt.Errorf("dnsmadeeasy: %v", err)
	}

	config := NewDefaultConfig()
	config.Sandbox = env.GetOrDefaultBool("DNSMADEEASY_SANDBOX", false)
	config.APIKey = values["DNSMADEEASY_API_KEY"]
	config.APISecret = values["DNSMADEEASY_API_SECRET"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DNS Made Easy.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsmadeeasy: the configuration of the DNS provider is nil")
	}

	var baseURL string
	if config.Sandbox {
		baseURL = "https://api.sandbox.dnsmadeeasy.com/V2.0"
	} else {
		if len(config.BaseURL) > 0 {
			baseURL = config.BaseURL
		} else {
			baseURL = "https://api.dnsmadeeasy.com/V2.0"
		}
	}

	client, err := internal.NewClient(config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("dnsmadeeasy: %v", err)
	}

	client.HTTPClient = config.HTTPClient
	client.BaseURL = baseURL

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domainName, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domainName, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to find zone for %s: %v", fqdn, err)
	}

	// fetch the domain details
	domain, err := d.client.GetDomain(authZone)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to get domain for zone %s: %v", authZone, err)
	}

	// create the TXT record
	name := strings.Replace(fqdn, "."+authZone, "", 1)
	record := &internal.Record{Type: "TXT", Name: name, Value: value, TTL: d.config.TTL}

	err = d.client.CreateRecord(domain, record)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to create record for %s: %v", name, err)
	}
	return nil
}

// CleanUp removes the TXT records matching the specified parameters
func (d *DNSProvider) CleanUp(domainName, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domainName, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to find zone for %s: %v", fqdn, err)
	}

	// fetch the domain details
	domain, err := d.client.GetDomain(authZone)
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to get domain for zone %s: %v", authZone, err)
	}

	// find matching records
	name := strings.Replace(fqdn, "."+authZone, "", 1)
	records, err := d.client.GetRecords(domain, name, "TXT")
	if err != nil {
		return fmt.Errorf("dnsmadeeasy: unable to get records for domain %s: %v", domain.Name, err)
	}

	// delete records
	var lastError error
	for _, record := range *records {
		err = d.client.DeleteRecord(record)
		if err != nil {
			lastError = fmt.Errorf("dnsmadeeasy: unable to delete record [id=%d, name=%s]: %v", record.ID, record.Name, err)
		}
	}

	return lastError
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
