// Package sakuracloud implements a DNS provider for solving the DNS-01 challenge using SakuraCloud DNS.
package sakuracloud

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/sacloud/libsacloud/api"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Token              string
	Secret             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("SAKURACLOUD_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("SAKURACLOUD_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("SAKURACLOUD_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("SAKURACLOUD_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *api.DNSAPI
}

// NewDNSProvider returns a DNSProvider instance configured for SakuraCloud.
// Credentials must be passed in the environment variables: SAKURACLOUD_ACCESS_TOKEN & SAKURACLOUD_ACCESS_TOKEN_SECRET
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("SAKURACLOUD_ACCESS_TOKEN", "SAKURACLOUD_ACCESS_TOKEN_SECRET")
	if err != nil {
		return nil, fmt.Errorf("sakuracloud: %v", err)
	}

	config := NewDefaultConfig()
	config.Token = values["SAKURACLOUD_ACCESS_TOKEN"]
	config.Secret = values["SAKURACLOUD_ACCESS_TOKEN_SECRET"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for SakuraCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("sakuracloud: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("sakuracloud: AccessToken is missing")
	}

	if config.Secret == "" {
		return nil, errors.New("sakuracloud: AccessSecret is missing")
	}

	apiClient := api.NewClient(config.Token, config.Secret, "is1a")
	if config.HTTPClient == nil {
		apiClient.HTTPClient = http.DefaultClient
	} else {
		apiClient.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		client: apiClient.GetDNSAPI(),
		config: config,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.addTXTRecord(fqdn, domain, value, d.config.TTL)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	return d.cleanupTXTRecord(fqdn, domain)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
