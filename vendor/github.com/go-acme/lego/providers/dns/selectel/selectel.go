// Package selectel implements a DNS provider for solving the DNS-01 challenge using Selectel Domains API.
// Selectel Domain API reference: https://kb.selectel.com/23136054.html
// Token: https://my.selectel.ru/profile/apikeys
package selectel

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/go-acme/lego/providers/dns/selectel/internal"
)

const (
	defaultBaseURL = "https://api.selectel.ru/domains/v1"
	minTTL         = 60
)

const (
	envNamespace             = "SELECTEL_"
	baseURLEnvVar            = envNamespace + "BASE_URL"
	apiTokenEnvVar           = envNamespace + "API_TOKEN"
	ttlEnvVar                = envNamespace + "TTL"
	propagationTimeoutEnvVar = envNamespace + "PROPAGATION_TIMEOUT"
	pollingIntervalEnvVar    = envNamespace + "POLLING_INTERVAL"
	httpTimeoutEnvVar        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            env.GetOrDefaultString(baseURLEnvVar, defaultBaseURL),
		TTL:                env.GetOrDefaultInt(ttlEnvVar, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(propagationTimeoutEnvVar, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(pollingIntervalEnvVar, 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(httpTimeoutEnvVar, 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Selectel Domains API.
// API token must be passed in the environment variable SELECTEL_API_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(apiTokenEnvVar)
	if err != nil {
		return nil, fmt.Errorf("selectel: %v", err)
	}

	config := NewDefaultConfig()
	config.Token = values[apiTokenEnvVar]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for selectel.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("selectel: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("selectel: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("selectel: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	client := internal.NewClient(internal.ClientOpts{
		BaseURL:    config.BaseURL,
		Token:      config.Token,
		HTTPClient: config.HTTPClient,
	})

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	domainObj, err := d.client.GetDomainByName(domain)
	if err != nil {
		return fmt.Errorf("selectel: %v", err)
	}

	txtRecord := internal.Record{
		Type:    "TXT",
		TTL:     d.config.TTL,
		Name:    fqdn,
		Content: value,
	}
	_, err = d.client.AddRecord(domainObj.ID, txtRecord)
	if err != nil {
		return fmt.Errorf("selectel: %v", err)
	}

	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	recordName := dns01.UnFqdn(fqdn)

	domainObj, err := d.client.GetDomainByName(domain)
	if err != nil {
		return fmt.Errorf("selectel: %v", err)
	}

	records, err := d.client.ListRecords(domainObj.ID)
	if err != nil {
		return fmt.Errorf("selectel: %v", err)
	}

	// Delete records with specific FQDN
	var lastErr error
	for _, record := range records {
		if record.Name == recordName {
			err = d.client.DeleteRecord(domainObj.ID, record.ID)
			if err != nil {
				lastErr = fmt.Errorf("selectel: %v", err)
			}
		}
	}

	return lastErr
}
