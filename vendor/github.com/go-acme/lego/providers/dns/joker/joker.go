// Package joker implements a DNS provider for solving the DNS-01 challenge using joker.com DMAPI.
package joker

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/log"
	"github.com/go-acme/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Debug              bool
	BaseURL            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
	AuthSid            string
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		Debug:              env.GetOrDefaultBool("JOKER_DEBUG", false),
		TTL:                env.GetOrDefaultInt("JOKER_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("JOKER_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("JOKER_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("JOKER_HTTP_TIMEOUT", 60*time.Second),
		},
	}
}

// DNSProvider is an implementation of the ChallengeProviderTimeout interface
// that uses Joker's DMAPI to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Joker DMAPI.
// Credentials must be passed in the environment variable JOKER_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("JOKER_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("joker: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["JOKER_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Joker DMAPI.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("joker: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("joker: credentials missing")
	}

	if !strings.HasSuffix(config.BaseURL, "/") {
		config.BaseURL += "/"
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present installs a TXT record for the DNS challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("joker: %v", err)
	}

	relative := getRelative(fqdn, zone)

	if d.config.Debug {
		log.Infof("[%s] joker: adding TXT record %q to zone %q with value %q", domain, relative, zone, value)
	}

	response, err := d.login()
	if err != nil {
		return formatResponseError(response, err)
	}

	response, err = d.getZone(zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone := addTxtEntryToZone(response.Body, relative, value, d.config.TTL)

	response, err = d.putZone(zone, dnsZone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	return nil
}

// CleanUp removes a TXT record used for a previous DNS challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("joker: %v", err)
	}

	relative := getRelative(fqdn, zone)

	if d.config.Debug {
		log.Infof("[%s] joker: removing entry %q from zone %q", domain, relative, zone)
	}

	response, err := d.login()
	if err != nil {
		return formatResponseError(response, err)
	}

	defer func() {
		// Try to logout in case of errors
		_, _ = d.logout()
	}()

	response, err = d.getZone(zone)
	if err != nil || response.StatusCode != 0 {
		return formatResponseError(response, err)
	}

	dnsZone, modified := removeTxtEntryFromZone(response.Body, relative)
	if modified {
		response, err = d.putZone(zone, dnsZone)
		if err != nil || response.StatusCode != 0 {
			return formatResponseError(response, err)
		}
	}

	response, err = d.logout()
	if err != nil {
		return formatResponseError(response, err)
	}
	return nil
}

func getRelative(fqdn, zone string) string {
	return dns01.UnFqdn(strings.TrimSuffix(fqdn, dns01.ToFqdn(zone)))
}

// formatResponseError formats error with optional details from DMAPI response
func formatResponseError(response *response, err error) error {
	if response != nil {
		return fmt.Errorf("joker: DMAPI error: %v Response: %v", err, response.Headers)
	}
	return fmt.Errorf("joker: DMAPI error: %v", err)
}
