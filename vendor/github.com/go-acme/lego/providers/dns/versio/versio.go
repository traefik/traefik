// Package versio implements a DNS provider for solving the DNS-01 challenge using versio DNS.
package versio

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            *url.URL
	TTL                int
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	baseURL, err := url.Parse(env.GetOrDefaultString("VERSIO_ENDPOINT", defaultBaseURL))
	if err != nil {
		baseURL, _ = url.Parse(defaultBaseURL)
	}

	return &Config{
		BaseURL:            baseURL,
		TTL:                env.GetOrDefaultInt("VERSIO_TTL", 300),
		PropagationTimeout: env.GetOrDefaultSecond("VERSIO_PROPAGATION_TIMEOUT", 60*time.Second),
		PollingInterval:    env.GetOrDefaultSecond("VERSIO_POLLING_INTERVAL", 5*time.Second),
		SequenceInterval:   env.GetOrDefaultSecond("VERSIO_SEQUENCE_INTERVAL", dns01.DefaultPropagationTimeout),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("VERSIO_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider describes a provider for acme-proxy
type DNSProvider struct {
	config       *Config
	dnsEntriesMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("VERSIO_USERNAME", "VERSIO_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("versio: %v", err)
	}

	config := NewDefaultConfig()
	config.Username = values["VERSIO_USERNAME"]
	config.Password = values["VERSIO_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider .
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("versio: the configuration of the DNS provider is nil")
	}
	if config.Username == "" {
		return nil, errors.New("versio: the versio username is missing")
	}
	if config.Password == "" {
		return nil, errors.New("versio: the versio password is missing")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("versio: %v", err)
	}

	// use mutex to prevent race condition from getDNSRecords until postDNSRecords
	d.dnsEntriesMu.Lock()
	defer d.dnsEntriesMu.Unlock()

	zoneName := dns01.UnFqdn(authZone)
	domains, err := d.getDNSRecords(zoneName)
	if err != nil {
		return fmt.Errorf("versio: %v", err)
	}

	txtRecord := record{
		Type:  "TXT",
		Name:  fqdn,
		Value: `"` + value + `"`,
		TTL:   d.config.TTL,
	}
	// Add new txtRercord to existing array of DNSRecords
	msg := &domains.Record
	msg.DNSRecords = append(msg.DNSRecords, txtRecord)

	err = d.postDNSRecords(zoneName, msg)
	if err != nil {
		return fmt.Errorf("versio: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("versio: %v", err)
	}

	// use mutex to prevent race condition from getDNSRecords until postDNSRecords
	d.dnsEntriesMu.Lock()
	defer d.dnsEntriesMu.Unlock()

	zoneName := dns01.UnFqdn(authZone)
	domains, err := d.getDNSRecords(zoneName)
	if err != nil {
		return fmt.Errorf("versio: %v", err)
	}

	// loop through the existing entries and remove the specific record
	msg := &dnsRecord{}
	for _, e := range domains.Record.DNSRecords {
		if e.Name != fqdn {
			msg.DNSRecords = append(msg.DNSRecords, e)
		}
	}

	err = d.postDNSRecords(zoneName, msg)
	if err != nil {
		return fmt.Errorf("versio: %v", err)
	}
	return nil
}
