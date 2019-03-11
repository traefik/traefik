// Package gandiv5 implements a DNS provider for solving the DNS-01 challenge using Gandi LiveDNS api.
package gandiv5

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

// Gandi API reference:       http://doc.livedns.gandi.net/

const (
	// defaultBaseURL endpoint is the Gandi API endpoint used by Present and CleanUp.
	defaultBaseURL = "https://dns.api.gandi.net/api/v5"
	minTTL         = 300
)

// inProgressInfo contains information about an in-progress challenge
type inProgressInfo struct {
	fieldName string
	authZone  string
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("GANDIV5_TTL", minTTL),
		PropagationTimeout: env.GetOrDefaultSecond("GANDIV5_PROPAGATION_TIMEOUT", 20*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("GANDIV5_POLLING_INTERVAL", 20*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("GANDIV5_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the
// acme.ChallengeProviderTimeout interface that uses Gandi's LiveDNS
// API to manage TXT records for a domain.
type DNSProvider struct {
	config          *Config
	inProgressFQDNs map[string]inProgressInfo
	inProgressMu    sync.Mutex
	// findZoneByFqdn determines the DNS zone of an fqdn. It is overridden during tests.
	findZoneByFqdn func(fqdn string) (string, error)
}

// NewDNSProvider returns a DNSProvider instance configured for Gandi.
// Credentials must be passed in the environment variable: GANDIV5_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("GANDIV5_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("gandi: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["GANDIV5_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gandi.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gandiv5: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("gandiv5: no API Key given")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("gandiv5: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	return &DNSProvider{
		config:          config,
		inProgressFQDNs: make(map[string]inProgressInfo),
		findZoneByFqdn:  dns01.FindZoneByFqdn,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// find authZone
	authZone, err := d.findZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("gandiv5: findZoneByFqdn failure: %v", err)
	}

	// determine name of TXT record
	if !strings.HasSuffix(
		strings.ToLower(fqdn), strings.ToLower("."+authZone)) {
		return fmt.Errorf("gandiv5: unexpected authZone %s for fqdn %s", authZone, fqdn)
	}
	name := fqdn[:len(fqdn)-len("."+authZone)]

	// acquire lock and check there is not a challenge already in
	// progress for this value of authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	// add TXT record into authZone
	err = d.addTXTRecord(dns01.UnFqdn(authZone), name, value, d.config.TTL)
	if err != nil {
		return err
	}

	// save data necessary for CleanUp
	d.inProgressFQDNs[fqdn] = inProgressInfo{
		authZone:  authZone,
		fieldName: name,
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	// acquire lock and retrieve authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.inProgressFQDNs[fqdn]; !ok {
		// if there is no cleanup information then just return
		return nil
	}

	fieldName := d.inProgressFQDNs[fqdn].fieldName
	authZone := d.inProgressFQDNs[fqdn].authZone
	delete(d.inProgressFQDNs, fqdn)

	// delete TXT record from authZone
	err := d.deleteTXTRecord(dns01.UnFqdn(authZone), fieldName)
	if err != nil {
		return fmt.Errorf("gandiv5: %v", err)
	}
	return nil
}

// Timeout returns the values (20*time.Minute, 20*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with Gandi.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
