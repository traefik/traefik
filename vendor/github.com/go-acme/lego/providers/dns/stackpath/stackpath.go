// Package stackpath implements a DNS provider for solving the DNS-01 challenge using Stackpath DNS.
// https://developer.stackpath.com/en/api/dns/
package stackpath

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/log"
	"github.com/go-acme/lego/platform/config/env"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultBaseURL = "https://gateway.stackpath.com/dns/v1/stacks/"
	defaultAuthURL = "https://gateway.stackpath.com/identity/v1/oauth2/token"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	ClientID           string
	ClientSecret       string
	StackID            string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("STACKPATH_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("STACKPATH_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("STACKPATH_POLLING_INTERVAL", dns01.DefaultPollingInterval),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	BaseURL *url.URL
	client  *http.Client
	config  *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Stackpath.
// Credentials must be passed in the environment variables:
// STACKPATH_CLIENT_ID, STACKPATH_CLIENT_SECRET, and STACKPATH_STACK_ID.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("STACKPATH_CLIENT_ID", "STACKPATH_CLIENT_SECRET", "STACKPATH_STACK_ID")
	if err != nil {
		return nil, fmt.Errorf("stackpath: %v", err)
	}

	config := NewDefaultConfig()
	config.ClientID = values["STACKPATH_CLIENT_ID"]
	config.ClientSecret = values["STACKPATH_CLIENT_SECRET"]
	config.StackID = values["STACKPATH_STACK_ID"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Stackpath.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("stackpath: the configuration of the DNS provider is nil")
	}

	if len(config.ClientID) == 0 || len(config.ClientSecret) == 0 {
		return nil, errors.New("stackpath: credentials missing")
	}

	if len(config.StackID) == 0 {
		return nil, errors.New("stackpath: stack id missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &DNSProvider{
		BaseURL: baseURL,
		client:  getOathClient(config),
		config:  config,
	}, nil
}

func getOathClient(config *Config) *http.Client {
	oathConfig := &clientcredentials.Config{
		TokenURL:     defaultAuthURL,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
	}

	return oathConfig.Client(context.Background())
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	zone, err := d.getZones(domain)
	if err != nil {
		return fmt.Errorf("stackpath: %v", err)
	}

	fqdn, value := dns01.GetRecord(domain, keyAuth)
	parts := strings.Split(fqdn, ".")

	record := Record{
		Name: parts[0],
		Type: "TXT",
		TTL:  d.config.TTL,
		Data: value,
	}

	return d.createZoneRecord(zone, record)
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	zone, err := d.getZones(domain)
	if err != nil {
		return fmt.Errorf("stackpath: %v", err)
	}

	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	parts := strings.Split(fqdn, ".")

	records, err := d.getZoneRecords(parts[0], zone)
	if err != nil {
		return err
	}

	for _, record := range records {
		err = d.deleteZoneRecord(zone, record)
		if err != nil {
			log.Printf("stackpath: failed to delete TXT record: %v", err)
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
