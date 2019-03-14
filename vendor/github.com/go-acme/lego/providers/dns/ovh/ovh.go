// Package ovh implements a DNS provider for solving the DNS-01 challenge using OVH DNS.
package ovh

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/ovh/go-ovh/ovh"
)

// OVH API reference:       https://eu.api.ovh.com/
// Create a Token:					https://eu.api.ovh.com/createToken/

// Record a DNS record
type Record struct {
	ID        int    `json:"id,omitempty"`
	FieldType string `json:"fieldType,omitempty"`
	SubDomain string `json:"subDomain,omitempty"`
	Target    string `json:"target,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
	Zone      string `json:"zone,omitempty"`
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIEndpoint        string
	ApplicationKey     string
	ApplicationSecret  string
	ConsumerKey        string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("OVH_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("OVH_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("OVH_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("OVH_HTTP_TIMEOUT", ovh.DefaultTimeout),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
// that uses OVH's REST API to manage TXT records for a domain.
type DNSProvider struct {
	config      *Config
	client      *ovh.Client
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for OVH
// Credentials must be passed in the environment variable:
// OVH_ENDPOINT : it must be ovh-eu or ovh-ca
// OVH_APPLICATION_KEY
// OVH_APPLICATION_SECRET
// OVH_CONSUMER_KEY
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("OVH_ENDPOINT", "OVH_APPLICATION_KEY", "OVH_APPLICATION_SECRET", "OVH_CONSUMER_KEY")
	if err != nil {
		return nil, fmt.Errorf("ovh: %v", err)
	}

	config := NewDefaultConfig()
	config.APIEndpoint = values["OVH_ENDPOINT"]
	config.ApplicationKey = values["OVH_APPLICATION_KEY"]
	config.ApplicationSecret = values["OVH_APPLICATION_SECRET"]
	config.ConsumerKey = values["OVH_CONSUMER_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OVH.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ovh: the configuration of the DNS provider is nil")
	}

	if config.APIEndpoint == "" || config.ApplicationKey == "" || config.ApplicationSecret == "" || config.ConsumerKey == "" {
		return nil, fmt.Errorf("ovh: credentials missing")
	}

	client, err := ovh.NewClient(
		config.APIEndpoint,
		config.ApplicationKey,
		config.ApplicationSecret,
		config.ConsumerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("ovh: %v", err)
	}

	client.Client = config.HTTPClient

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// Parse domain name
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("ovh: could not determine zone for domain: '%s'. %s", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)
	subDomain := d.extractRecordName(fqdn, authZone)

	reqURL := fmt.Sprintf("/domain/zone/%s/record", authZone)
	reqData := Record{FieldType: "TXT", SubDomain: subDomain, Target: value, TTL: d.config.TTL}

	// Create TXT record
	var respData Record
	err = d.client.Post(reqURL, reqData, &respData)
	if err != nil {
		return fmt.Errorf("ovh: error when call api to add record (%s): %v", reqURL, err)
	}

	// Apply the change
	reqURL = fmt.Sprintf("/domain/zone/%s/refresh", authZone)
	err = d.client.Post(reqURL, nil, nil)
	if err != nil {
		return fmt.Errorf("ovh: error when call api to refresh zone (%s): %v", reqURL, err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[fqdn] = respData.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[fqdn]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("ovh: unknown record ID for '%s'", fqdn)
	}

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("ovh: could not determine zone for domain: '%s'. %s", domain, err)
	}

	authZone = dns01.UnFqdn(authZone)

	reqURL := fmt.Sprintf("/domain/zone/%s/record/%d", authZone, recordID)

	err = d.client.Delete(reqURL, nil)
	if err != nil {
		return fmt.Errorf("ovh: error when call OVH api to delete challenge record (%s): %v", reqURL, err)
	}

	// Apply the change
	reqURL = fmt.Sprintf("/domain/zone/%s/refresh", authZone)
	err = d.client.Post(reqURL, nil, nil)
	if err != nil {
		return fmt.Errorf("ovh: error when call api to refresh zone (%s): %v", reqURL, err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, fqdn)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
